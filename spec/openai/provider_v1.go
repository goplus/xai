/*
 * Copyright (c) 2026 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"log"
	"net/http"
	"os"
	"strings"
	"unsafe"

	xai "github.com/goplus/xai/spec"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/shared"
)

// -----------------------------------------------------------------------------

// v1Provider implements provider using the Chat Completions API.
// When baseURL and apiKey are set, it uses custom HTTP and extended response
// parsing for APIs that return an "images" field (e.g. Qiniu gemini-2.5-flash-image).
type v1Provider struct {
	chat    openai.ChatCompletionService
	baseURL string
	apiKey  string
}

func newV1Provider(opts []option.RequestOption, baseURL, apiKey string) *v1Provider {
	p := &v1Provider{
		chat: openai.NewChatCompletionService(opts...),
	}
	if baseURL != "" && apiKey != "" {
		p.baseURL = baseURL
		p.apiKey = apiKey
	}
	return p
}

func (p *v1Provider) Features() xai.Feature {
	return xai.FeatureGen | xai.FeatureGenStream
}

func (p *v1Provider) Gen(ctx context.Context, req *genRequest, opts []option.RequestOption) (genResponse, error) {
	params := p.buildParams(req)
	if p.baseURL != "" && p.apiKey != "" {
		return p.genWithExtendedParsing(ctx, params)
	}
	resp, err := p.chat.New(ctx, params, opts...)
	if err != nil {
		return nil, err
	}
	return &v1Response{msg: resp}, nil
}

func (p *v1Provider) genWithExtendedParsing(ctx context.Context, params openai.ChatCompletionNewParams) (genResponse, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	log.Printf("[openai] curl command:\n%s", buildCurlCommand(httpReq, body))

	if os.Getenv("QINIU_MOCK_CURL") != "" {
		return mockExtendedChatCompletionResponse(body)
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chat completions: %s: %s", resp.Status, string(respBody))
	}
	return parseChatCompletionResponseExtended(respBody)
}

func mockExtendedChatCompletionResponse(body []byte) (genResponse, error) {
	hasTools := bytes.Contains(body, []byte(`"tools"`))
	hasToolResult := bytes.Contains(body, []byte(`"role":"tool"`))
	hasImageURL := bytes.Contains(body, []byte(`"image_url"`))

	switch {
	case hasTools && !hasToolResult:
		return parseChatCompletionResponseExtended([]byte(`{
			"id":"chatcmpl-mock-tool",
			"choices":[{
				"index":0,
				"finish_reason":"tool_calls",
				"message":{
					"role":"assistant",
					"content":"",
					"tool_calls":[{
						"id":"call_mock_weather",
						"type":"function",
						"function":{
							"name":"get_weather",
							"arguments":"{\"city\":\"Shanghai\"}"
						}
					}]
				}
			}]
		}`))
	case hasTools && hasToolResult:
		return parseChatCompletionResponseExtended([]byte(`{
			"id":"chatcmpl-mock-tool-final",
			"choices":[{
				"index":0,
				"finish_reason":"stop",
				"message":{
					"role":"assistant",
					"content":"Shanghai is sunny and 26C."
				}
			}]
		}`))
	case hasImageURL:
		return parseChatCompletionResponseExtended([]byte(`{
			"id":"chatcmpl-mock-image",
			"choices":[{
				"index":0,
				"finish_reason":"stop",
				"message":{
					"role":"assistant",
					"content":"I changed the image to a red style.",
					"images":[{
						"type":"image_url",
						"image_url":{"url":"data:image/png;base64,aGVsbG8="}
					}]
				}
			}]
		}`))
	default:
		return parseChatCompletionResponseExtended([]byte(`{
			"id":"chatcmpl-mock-text",
			"choices":[{
				"index":0,
				"finish_reason":"stop",
				"message":{
					"role":"assistant",
					"content":"Gemini mock response."
				}
			}]
		}`))
	}
}

func (p *v1Provider) GenStream(ctx context.Context, req *genRequest, opts []option.RequestOption) iter.Seq2[genResponse, error] {
	params := p.buildParams(req)
	stream := p.chat.NewStreaming(ctx, params, opts...)
	return p.buildRespIter(stream)
}

// chatCompletionResponseRaw parses the extended response format with images.
type chatCompletionResponseRaw struct {
	ID      string `json:"id"`
	Choices []struct {
		Index        int32  `json:"index"`
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Role    string `json:"role"`
			Content string `json:"content"`
			Images  []struct {
				Type     string `json:"type"`
				ImageURL struct {
					URL string `json:"url"`
				} `json:"image_url"`
			} `json:"images"`
			ToolCalls []openai.ChatCompletionMessageToolCallUnion `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
}

func parseChatCompletionResponseExtended(body []byte) (genResponse, error) {
	var raw chatCompletionResponseRaw
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if len(raw.Choices) == 0 {
		return nil, fmt.Errorf("chat completions: empty choices")
	}

	choices := make([]openai.ChatCompletionChoice, len(raw.Choices))
	imagesPerChoice := make([][]xai.Blob, len(raw.Choices))
	rawPerChoice := make([]*chatCompletionMessageRaw, len(raw.Choices))

	for i, c := range raw.Choices {
		choices[i] = openai.ChatCompletionChoice{
			Index:        int64(c.Index),
			FinishReason: c.FinishReason,
			Message: openai.ChatCompletionMessage{
				Content:   c.Message.Content,
				ToolCalls: c.Message.ToolCalls,
			},
		}
		var blobs []xai.Blob
		for _, img := range c.Message.Images {
			if url := img.ImageURL.URL; url != "" {
				b := blobFromImageURL(url)
				if b.BlobData != nil {
					blobs = append(blobs, b)
				}
			}
		}
		imagesPerChoice[i] = blobs
		if len(blobs) > 0 {
			rawPerChoice[i] = &chatCompletionMessageRaw{
				Content:   c.Message.Content,
				ToolCalls: c.Message.ToolCalls,
			}
			for j := range c.Message.Images {
				rawPerChoice[i].Images = append(rawPerChoice[i].Images, struct {
					Type     string `json:"type"`
					ImageURL struct {
						URL string `json:"url"`
					} `json:"image_url"`
				}{
					Type: c.Message.Images[j].Type,
					ImageURL: struct {
						URL string `json:"url"`
					}{URL: c.Message.Images[j].ImageURL.URL},
				})
			}
		}
	}

	return &v1ResponseWithImages{
		msg: &openai.ChatCompletion{
			Choices: choices,
		},
		images: imagesPerChoice,
		raw:    rawPerChoice,
	}, nil
}

func blobFromImageURL(rawURL string) xai.Blob {
	if !strings.HasPrefix(rawURL, "data:") {
		return xai.Blob{}
	}
	payload := strings.TrimPrefix(rawURL, "data:")
	pos := strings.Index(payload, ",")
	if pos < 0 {
		return xai.Blob{}
	}
	header := payload[:pos]
	mime := strings.TrimSpace(strings.Split(header, ";")[0])
	if mime == "" {
		mime = "image/png"
	}
	b64 := payload[pos+1:]
	buf, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return xai.Blob{}
	}
	return xai.Blob{
		MIME:     mime,
		BlobData: xai.BlobFromRaw(buf),
	}
}

func (p *v1Provider) buildParams(req *genRequest) openai.ChatCompletionNewParams {
	var params openai.ChatCompletionNewParams
	params.Model = shared.ChatModel(req.Model)

	if req.MaxOutputTokens > 0 {
		params.MaxCompletionTokens = param.NewOpt(req.MaxOutputTokens)
	}
	if req.HasTemperature {
		params.Temperature = param.NewOpt(req.Temperature)
	}
	if req.HasTopP {
		params.TopP = param.NewOpt(req.TopP)
	}

	// Build tools
	if len(req.Tools) > 0 {
		var tools []openai.ChatCompletionToolUnionParam
		for _, t := range req.Tools {
			if t.IsWebSearch {
				continue
			}
			tools = append(tools, openai.ChatCompletionToolUnionParam{
				OfFunction: &openai.ChatCompletionFunctionToolParam{
					Function: shared.FunctionDefinitionParam{
						Name:        t.Name,
						Description: param.NewOpt(t.Description),
						// Keep schema explicit for OpenAI-compatible backends that require parameters.type.
						Parameters: shared.FunctionParameters{
							"type":       "object",
							"properties": map[string]any{},
						},
					},
				},
			})
		}
		params.Tools = tools
	}

	// Build messages
	params.Messages = p.buildMessages(req)
	return params
}

func (p *v1Provider) buildMessages(req *genRequest) []openai.ChatCompletionMessageParamUnion {
	var result []openai.ChatCompletionMessageParamUnion

	// Add system message
	if len(req.System) > 0 {
		var sb strings.Builder
		for i, t := range req.System {
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(t.Text)
		}
		result = append(result, openai.SystemMessage(sb.String()))
	}

	// Add messages
	for _, msg := range req.Messages {
		items := p.buildMessage(msg)
		result = append(result, items...)
	}

	return result
}

func (p *v1Provider) buildMessage(msg *message) []openai.ChatCompletionMessageParamUnion {
	var result []openai.ChatCompletionMessageParamUnion

	switch msg.Role {
	case "user":
		var content []openai.ChatCompletionContentPartUnionParam
		for _, c := range msg.Contents {
			switch c.Type {
			case contentText:
				content = append(content, openai.ChatCompletionContentPartUnionParam{
					OfText: &openai.ChatCompletionContentPartTextParam{
						Text: c.Text,
					},
				})
			case contentImageURL:
				imgParam := openai.ChatCompletionContentPartImageImageURLParam{
					URL: c.ImageURL,
				}
				if c.ImageDetail != "" {
					imgParam.Detail = c.ImageDetail
				}
				content = append(content, openai.ChatCompletionContentPartUnionParam{
					OfImageURL: &openai.ChatCompletionContentPartImageParam{
						ImageURL: imgParam,
					},
				})
			case contentImageFile:
				content = append(content, openai.ChatCompletionContentPartUnionParam{
					OfFile: &openai.ChatCompletionContentPartFileParam{
						File: openai.ChatCompletionContentPartFileFileParam{
							FileID: param.NewOpt(c.FileID),
						},
					},
				})
			case contentDocFile:
				content = append(content, openai.ChatCompletionContentPartUnionParam{
					OfFile: &openai.ChatCompletionContentPartFileParam{
						File: openai.ChatCompletionContentPartFileFileParam{
							FileID: param.NewOpt(c.FileID),
						},
					},
				})
			}
		}
		if len(content) > 0 {
			result = append(result, openai.ChatCompletionMessageParamUnion{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfArrayOfContentParts: content,
					},
				},
			})
		}

	case "assistant":
		var textContent string
		var toolCalls []openai.ChatCompletionMessageToolCallUnionParam
		for _, c := range msg.Contents {
			switch c.Type {
			case contentText:
				textContent = c.Text
			case contentToolUse:
				args := unsafe.String(unsafe.SliceData(c.ToolUse.Input), len(c.ToolUse.Input))
				toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallUnionParam{
					OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
						ID: c.ToolUse.ID,
						Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
							Name:      c.ToolUse.Name,
							Arguments: args,
						},
					},
				})
			}
		}
		assistantMsg := &openai.ChatCompletionAssistantMessageParam{}
		if textContent != "" {
			assistantMsg.Content.OfString = param.NewOpt(textContent)
		}
		if len(toolCalls) > 0 {
			assistantMsg.ToolCalls = toolCalls
		}
		result = append(result, openai.ChatCompletionMessageParamUnion{
			OfAssistant: assistantMsg,
		})

	case "tool":
		for _, c := range msg.Contents {
			if c.Type == contentToolResult {
				ret := unsafe.String(unsafe.SliceData(c.ToolResult.Result), len(c.ToolResult.Result))
				result = append(result, openai.ChatCompletionMessageParamUnion{
					OfTool: &openai.ChatCompletionToolMessageParam{
						ToolCallID: c.ToolResult.ID,
						Content: openai.ChatCompletionToolMessageParamContentUnion{
							OfString: param.NewOpt(ret),
						},
					},
				})
			}
		}
	}

	return result
}

func (p *v1Provider) buildRespIter(stream *ssestream.Stream[openai.ChatCompletionChunk]) iter.Seq2[genResponse, error] {
	return func(yield func(genResponse, error) bool) {
		defer stream.Close()
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta.Content != "" {
					if !yield(&v1StreamChunk{text: delta.Content}, nil) {
						return
					}
				}
			}
		}
		if err := stream.Err(); err != nil {
			yield(&v1StreamChunk{}, err)
		}
	}
}

// -----------------------------------------------------------------------------
