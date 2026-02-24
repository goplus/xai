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
	"github.com/openai/openai-go/v3/responses"
)

// -----------------------------------------------------------------------------

type msgBuilder struct {
	msgs []responses.ResponseInputItemUnionParam
}

func (p *msgBuilder) User(content xai.ContentBuilder) xai.MessageBuilder {
	return p
}

func (p *msgBuilder) Assistant(content xai.ContentBuilder) xai.MessageBuilder {
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

func (p *Provider) Contents() xai.ContentBuilder {
	panic("todo")
}

// -----------------------------------------------------------------------------

func (p *Provider) Parts() xai.MultipartBuilder {
	panic("todo")
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
