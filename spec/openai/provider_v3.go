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
	"unsafe"

	xai "github.com/goplus/xai/spec"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

// -----------------------------------------------------------------------------

// v3Provider implements provider using the Responses API.
type v3Provider struct {
	responses responses.ResponseService
}

func newV3Provider(opts []option.RequestOption) *v3Provider {
	return &v3Provider{
		responses: responses.NewResponseService(opts...),
	}
}

func (p *v3Provider) Features() xai.Feature {
	return xai.FeatureGen | xai.FeatureGenStream | xai.FeatureOperation
}

func (p *v3Provider) Gen(ctx context.Context, req *genRequest, opts []option.RequestOption) (genResponse, error) {
	params := p.buildParams(req)
	resp, err := p.responses.New(ctx, params, opts...)
	if err != nil {
		return nil, err
	}
	return &v3Response{msg: resp}, nil
}

func (p *v3Provider) GenStream(ctx context.Context, req *genRequest, opts []option.RequestOption) iter.Seq2[genResponse, error] {
	params := p.buildParams(req)
	stream := p.responses.NewStreaming(ctx, params, opts...)
	return p.buildRespIter(stream)
}

func (p *v3Provider) buildParams(req *genRequest) responses.ResponseNewParams {
	var params responses.ResponseNewParams
	params.Model = shared.ResponsesModel(req.Model)

	if req.MaxOutputTokens > 0 {
		params.MaxOutputTokens = param.NewOpt(req.MaxOutputTokens)
	}
	if req.HasTemperature {
		params.Temperature = param.NewOpt(req.Temperature)
	}
	if req.HasTopP {
		params.TopP = param.NewOpt(req.TopP)
	}

	// Build tools
	if len(req.Tools) > 0 {
		params.Tools = make([]responses.ToolUnionParam, len(req.Tools))
		for i, t := range req.Tools {
			if t.IsWebSearch {
				params.Tools[i].OfWebSearch = &responses.WebSearchToolParam{
					Type: "web_search_2025_08_26",
				}
			} else {
				params.Tools[i].OfFunction = &responses.FunctionToolParam{
					Name:        t.Name,
					Description: param.NewOpt(t.Description),
					Strict:      param.NewOpt(false),
					Parameters: map[string]any{
						"type":       "object",
						"properties": map[string]any{},
					},
				}
			}
		}
	}

	// Build input messages
	params.Input = p.buildInput(req)
	return params
}

func (p *v3Provider) buildInput(req *genRequest) responses.ResponseNewParamsInputUnion {
	var result responses.ResponseNewParamsInputUnion

	// Calculate capacity
	n := len(req.Messages)
	if len(req.System) > 0 {
		n++
	}

	msgs := make([]responses.ResponseInputItemUnionParam, 0, n)

	// Add system message
	if len(req.System) > 0 {
		sysContent := make(responses.ResponseInputMessageContentListParam, len(req.System))
		for i, t := range req.System {
			sysContent[i] = responses.ResponseInputContentParamOfInputText(t.Text)
		}
		msgs = append(msgs, responses.ResponseInputItemParamOfMessage(sysContent, responses.EasyInputMessageRoleSystem))
	}

	// Add messages
	for _, msg := range req.Messages {
		items := p.buildMessage(msg)
		msgs = append(msgs, items...)
	}

	result.OfInputItemList = msgs
	return result
}

func (p *v3Provider) buildMessage(msg *message) []responses.ResponseInputItemUnionParam {
	var result []responses.ResponseInputItemUnionParam
	var currentContent responses.ResponseInputMessageContentListParam
	var role responses.EasyInputMessageRole

	switch msg.Role {
	case "user":
		role = responses.EasyInputMessageRoleUser
	case "assistant":
		role = responses.EasyInputMessageRoleAssistant
	case "system":
		role = responses.EasyInputMessageRoleSystem
	default:
		role = responses.EasyInputMessageRoleUser
	}

	flushContent := func() {
		if len(currentContent) > 0 {
			result = append(result, responses.ResponseInputItemParamOfMessage(currentContent, role))
			currentContent = nil
		}
	}

	for _, c := range msg.Contents {
		switch c.Type {
		case contentText:
			currentContent = append(currentContent, responses.ResponseInputContentParamOfInputText(c.Text))

		case contentImageURL:
			img := &responses.ResponseInputImageParam{
				ImageURL: param.NewOpt(c.ImageURL),
			}
			if c.ImageDetail != "" {
				img.Detail = responses.ResponseInputImageDetail(c.ImageDetail)
			}
			currentContent = append(currentContent, responses.ResponseInputContentUnionParam{OfInputImage: img})

		case contentImageFile:
			currentContent = append(currentContent, responses.ResponseInputContentUnionParam{
				OfInputImage: &responses.ResponseInputImageParam{
					FileID: param.NewOpt(c.FileID),
				},
			})

		case contentDocURL:
			currentContent = append(currentContent, responses.ResponseInputContentUnionParam{
				OfInputFile: &responses.ResponseInputFileParam{
					FileURL: param.NewOpt(c.FileURL),
				},
			})

		case contentDocFile:
			currentContent = append(currentContent, responses.ResponseInputContentUnionParam{
				OfInputFile: &responses.ResponseInputFileParam{
					FileID: param.NewOpt(c.FileID),
				},
			})

		case contentToolUse:
			flushContent()
			args := unsafe.String(unsafe.SliceData(c.ToolUse.Input), len(c.ToolUse.Input))
			result = append(result, responses.ResponseInputItemParamOfFunctionCall(c.ToolUse.ID, args, c.ToolUse.Name))

		case contentToolResult:
			flushContent()
			ret := unsafe.String(unsafe.SliceData(c.ToolResult.Result), len(c.ToolResult.Result))
			result = append(result, responses.ResponseInputItemParamOfFunctionCallOutput(c.ToolResult.ID, ret))

		case contentThinking:
			flushContent()
			result = append(result, responses.ResponseInputItemUnionParam{
				OfReasoning: &responses.ResponseReasoningItemParam{
					ID: c.Thinking.Signature,
					Content: []responses.ResponseReasoningItemContentParam{
						{Text: c.Thinking.Text},
					},
				},
			})

		case contentCompaction:
			flushContent()
			result = append(result, responses.ResponseInputItemParamOfCompaction(c.Compaction))
		}
	}

	flushContent()
	return result
}

func (p *v3Provider) buildRespIter(stream *ssestream.Stream[responses.ResponseStreamEventUnion]) iter.Seq2[genResponse, error] {
	return func(yield func(genResponse, error) bool) {
		defer stream.Close()
		for stream.Next() {
			ev := stream.Current()
			switch ev.Type {
			case "response.output_text.delta":
				delta := ev.AsResponseOutputTextDelta()
				if delta.Delta != "" {
					if !yield(&v3StreamChunk{text: delta.Delta}, nil) {
						return
					}
				}
			case "response.reasoning_text.delta":
				delta := ev.AsResponseReasoningTextDelta()
				if delta.Delta != "" {
					if !yield(&v3StreamChunk{text: delta.Delta}, nil) {
						return
					}
				}
			case "error":
				errEv := ev.AsError()
				if errEv.Message != "" {
					if !yield(&v3StreamChunk{}, &streamError{msg: errEv.Message}) {
						return
					}
				}
			}
		}
		if err := stream.Err(); err != nil {
			yield(&v3StreamChunk{}, err)
		}
	}
}

// -----------------------------------------------------------------------------
