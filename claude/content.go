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

package claude

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/goplus/xai"
)

// -----------------------------------------------------------------------------

type msgBuilder struct {
	msgs []anthropic.MessageParam
}

func (p *msgBuilder) User(content xai.ContentBuilder) xai.MessageBuilder {
	p.msgs = append(p.msgs, anthropic.NewUserMessage(buildContents(content)...))
	return p
}

func (p *msgBuilder) Assistant(content xai.ContentBuilder) xai.MessageBuilder {
	p.msgs = append(p.msgs, anthropic.NewAssistantMessage(buildContents(content)...))
	return p
}

func (p *Provider) Messages() xai.MessageBuilder {
	return &msgBuilder{}
}

func buildMessages(in xai.MessageBuilder) []anthropic.MessageParam {
	return in.(*msgBuilder).msgs
}

// -----------------------------------------------------------------------------

type contentBuilder struct {
	content []anthropic.ContentBlockParamUnion
}

func (p *Provider) Contents() xai.ContentBuilder {
	return &contentBuilder{}
}

func buildContents(in xai.ContentBuilder) []anthropic.ContentBlockParamUnion {
	return in.(*contentBuilder).content
}

// -----------------------------------------------------------------------------

type textBuilder struct {
	msgs []anthropic.TextBlockParam
}

func (p *Provider) Texts() xai.TextBuilder {
	panic("todo")
}

func buildTexts(in xai.TextBuilder) []anthropic.TextBlockParam {
	return in.(*textBuilder).msgs
}

// -----------------------------------------------------------------------------

func (p *Provider) Tools() xai.ToolBuilder {
	panic("todo")
}

// -----------------------------------------------------------------------------
