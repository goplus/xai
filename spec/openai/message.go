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
	xai "github.com/goplus/xai/spec"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
)

// -----------------------------------------------------------------------------

type msgBuilder struct {
	content []responses.ResponseInputItemUnionParam
	msg     *responses.EasyInputMessageParam
	role    responses.EasyInputMessageRole
}

func buildMessages(in []xai.MsgBuilder, sysPrompt responses.ResponseInputItemUnionParam) (ret responses.ResponseNewParamsInputUnion) {
	sys := sysPrompt.OfMessage != nil
	n := len(in)
	if sys {
		n++
	}
	msgs := make([]responses.ResponseInputItemUnionParam, 0, n)
	if sys {
		msgs = append(msgs, sysPrompt)
	}
	for _, v := range in {
		msgs = append(msgs, v.(*msgBuilder).content[0])
	}
	ret.OfInputItemList = msgs
	return
}

func (p *Service) UserMsg() xai.MsgBuilder {
	return &msgBuilder{role: responses.EasyInputMessageRoleUser}
}

func (p *Service) AssistantMsg() xai.MsgBuilder {
	return &msgBuilder{role: responses.EasyInputMessageRoleAssistant}
}

func (p *msgBuilder) addMsg(v responses.ResponseInputContentUnionParam) xai.MsgBuilder {
	if p.msg == nil {
		content := responses.ResponseInputItemParamOfMessage(responses.ResponseInputMessageContentListParam{v}, p.role)
		p.content = append(p.content, content)
		p.msg = content.OfMessage
	} else {
		p.msg.Content.OfInputItemContentList = append(p.msg.Content.OfInputItemContentList, v)
	}
	return p
}

func (p *msgBuilder) addNonMsg(v responses.ResponseInputItemUnionParam) xai.MsgBuilder {
	p.content = append(p.content, v)
	p.msg = nil
	return p
}

func (p *msgBuilder) Text(text string) xai.MsgBuilder {
	return p.addMsg(responses.ResponseInputContentParamOfInputText(text))
}

func (p *msgBuilder) Image(image xai.ImageData) xai.MsgBuilder {
	panic("todo")
}

func (p *msgBuilder) ImageURL(mime xai.ImageType, url string) xai.MsgBuilder {
	return p.addMsg(responses.ResponseInputContentUnionParam{
		OfInputImage: &responses.ResponseInputImageParam{
			ImageURL: param.NewOpt(url),
		},
	})
}

func (p *msgBuilder) ImageFile(mime xai.ImageType, fileID string) xai.MsgBuilder {
	return p.addMsg(responses.ResponseInputContentUnionParam{
		OfInputImage: &responses.ResponseInputImageParam{
			FileID: param.NewOpt(fileID),
		},
	})
}

func (p *msgBuilder) Doc(doc xai.DocumentData) xai.MsgBuilder {
	panic("todo")
}

func (p *msgBuilder) DocURL(mime xai.DocumentType, url string) xai.MsgBuilder {
	return p.addMsg(responses.ResponseInputContentUnionParam{
		OfInputFile: &responses.ResponseInputFileParam{
			FileURL: param.NewOpt(url),
		},
	})
}

func (p *msgBuilder) DocFile(mime xai.DocumentType, fileID string) xai.MsgBuilder {
	return p.addMsg(responses.ResponseInputContentUnionParam{
		OfInputFile: &responses.ResponseInputFileParam{
			FileID: param.NewOpt(fileID),
		},
	})
}

func (p *msgBuilder) Part(part xai.Part) xai.MsgBuilder {
	p.content = append(p.content, buildPart(part))
	return p
}

func (p *msgBuilder) Thinking(v xai.Thinking) xai.MsgBuilder {
	if v.Redacted {
		panic("todo")
	}
	return p.addNonMsg(responses.ResponseInputItemUnionParam{
		OfReasoning: &responses.ResponseReasoningItemParam{
			ID: v.Signature,
			Content: []responses.ResponseReasoningItemContentParam{
				{Text: v.Text},
			},
		},
	})
}

func (p *msgBuilder) Compaction(data string) xai.MsgBuilder {
	return p.addNonMsg(responses.ResponseInputItemParamOfCompaction(data))
}

// -----------------------------------------------------------------------------
