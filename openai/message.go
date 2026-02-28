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
	"github.com/goplus/xai"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
)

// -----------------------------------------------------------------------------

type msgBuilder struct {
	msgs []responses.ResponseInputItemUnionParam
}

func (p *msgBuilder) User(content xai.ContentBuilder) xai.MessageBuilder {
	p.msgs = append(p.msgs, buildContents(content)...)
	return p
}

func (p *msgBuilder) Assistant(content xai.ContentBuilder) xai.MessageBuilder {
	contents := buildContents(content)
	for _, c := range contents {
		if c.OfMessage != nil {
			c.OfMessage.Role = responses.EasyInputMessageRoleAssistant
		}
	}
	p.msgs = append(p.msgs, contents...)
	return p
}

func (p *Provider) Messages() xai.MessageBuilder {
	// we reserve the first slot for system prompt, which is optional but commonly used
	msgs := make([]responses.ResponseInputItemUnionParam, 1, 2)
	return &msgBuilder{msgs: msgs}
}

func buildMessages(in xai.MessageBuilder, sys responses.ResponseInputItemUnionParam) (ret responses.ResponseNewParamsInputUnion) {
	p := in.(*msgBuilder)
	msgs := p.msgs
	if sys.OfMessage != nil {
		msgs[0] = sys // system prompt
	} else {
		msgs = msgs[1:]
	}
	ret.OfInputItemList = msgs
	return
}

// -----------------------------------------------------------------------------

type contentBuilder struct {
	content []responses.ResponseInputItemUnionParam
	msg     *responses.EasyInputMessageParam
}

func (p *contentBuilder) addMsg(v responses.ResponseInputContentUnionParam) xai.ContentBuilder {
	if p.msg == nil {
		content := responses.ResponseInputItemParamOfMessage(responses.ResponseInputMessageContentListParam{v}, responses.EasyInputMessageRoleUser)
		p.content = append(p.content, content)
		p.msg = content.OfMessage
	} else {
		p.msg.Content.OfInputItemContentList = append(p.msg.Content.OfInputItemContentList, v)
	}
	return p
}

func (p *contentBuilder) addNonMsg(v responses.ResponseInputItemUnionParam) xai.ContentBuilder {
	p.content = append(p.content, v)
	p.msg = nil
	return p
}

func (p *contentBuilder) Text(text string) xai.ContentBuilder {
	return p.addMsg(responses.ResponseInputContentParamOfInputText(text))
}

func (p *contentBuilder) Image(image xai.ImageData) xai.ContentBuilder {
	panic("todo")
}

func (p *contentBuilder) ImageURL(mime xai.ImageType, url string) xai.ContentBuilder {
	return p.addMsg(responses.ResponseInputContentUnionParam{
		OfInputImage: &responses.ResponseInputImageParam{
			ImageURL: param.NewOpt(url),
		},
	})
}

func (p *contentBuilder) ImageFile(mime xai.ImageType, fileID string) xai.ContentBuilder {
	return p.addMsg(responses.ResponseInputContentUnionParam{
		OfInputImage: &responses.ResponseInputImageParam{
			FileID: param.NewOpt(fileID),
		},
	})
}

func (p *contentBuilder) Doc(doc xai.DocumentData) xai.ContentBuilder {
	panic("todo")
}

func (p *contentBuilder) DocURL(mime xai.DocumentType, url string) xai.ContentBuilder {
	return p.addMsg(responses.ResponseInputContentUnionParam{
		OfInputFile: &responses.ResponseInputFileParam{
			FileURL: param.NewOpt(url),
		},
	})
}

func (p *contentBuilder) DocFile(mime xai.DocumentType, fileID string) xai.ContentBuilder {
	return p.addMsg(responses.ResponseInputContentUnionParam{
		OfInputFile: &responses.ResponseInputFileParam{
			FileID: param.NewOpt(fileID),
		},
	})
}

func (p *contentBuilder) Thinking(signature, thinking string) xai.ContentBuilder {
	return p.addNonMsg(responses.ResponseInputItemUnionParam{
		OfReasoning: &responses.ResponseReasoningItemParam{
			ID: signature,
			Content: []responses.ResponseReasoningItemContentParam{
				{Text: thinking},
			},
		},
	})
}

func (p *contentBuilder) RedactedThinking(data string) xai.ContentBuilder {
	panic("todo")
}

func (p *Provider) Contents() xai.ContentBuilder {
	return &contentBuilder{}
}

func buildContents(in xai.ContentBuilder) []responses.ResponseInputItemUnionParam {
	return in.(*contentBuilder).content
}

// -----------------------------------------------------------------------------

type textBuilder struct {
	content responses.ResponseInputMessageContentListParam
}

func (p *textBuilder) Text(text string) xai.TextBuilder {
	p.content = append(p.content, responses.ResponseInputContentParamOfInputText(text))
	return p
}

func (p *Provider) Texts() xai.TextBuilder {
	return &textBuilder{}
}

func buildTexts(in xai.TextBuilder) responses.ResponseInputMessageContentListParam {
	return in.(*textBuilder).content
}

// -----------------------------------------------------------------------------

func (p *Provider) Tools() xai.ToolBuilder {
	panic("todo")
}

// -----------------------------------------------------------------------------
