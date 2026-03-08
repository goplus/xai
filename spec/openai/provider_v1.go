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
	"context"
	"iter"
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
type v1Provider struct {
	chat openai.ChatCompletionService
}

func newV1Provider(opts []option.RequestOption) *v1Provider {
	return &v1Provider{
		chat: openai.NewChatCompletionService(opts...),
	}
}

func (p *v1Provider) Features() xai.Feature {
	return xai.FeatureGen | xai.FeatureGenStream
}

func (p *v1Provider) Gen(ctx context.Context, req *genRequest, opts []option.RequestOption) (genResponse, error) {
	params := p.buildParams(req)
	resp, err := p.chat.New(ctx, params, opts...)
	if err != nil {
		return nil, err
	}
	return &v1Response{msg: resp}, nil
}

func (p *v1Provider) GenStream(ctx context.Context, req *genRequest, opts []option.RequestOption) iter.Seq2[genResponse, error] {
	params := p.buildParams(req)
	stream := p.chat.NewStreaming(ctx, params, opts...)
	return p.buildRespIter(stream)
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
