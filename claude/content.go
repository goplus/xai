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
	"unsafe"

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

func (p *contentBuilder) Text(s string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewTextBlock(s))
	return p
}

func (p *contentBuilder) ImageURL(url string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewImageBlock(anthropic.URLImageSourceParam{
		URL: url,
	}))
	return p
}

func (p *contentBuilder) ImageBase64(mime xai.ImageType, base64 []byte) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewImageBlockBase64(
		string(mime), unsafe.String(unsafe.SliceData(base64), len(base64)),
	))
	return p
}

func (p *Provider) Contents() xai.ContentBuilder {
	return &contentBuilder{}
}

func buildContents(in xai.ContentBuilder) []anthropic.ContentBlockParamUnion {
	return in.(*contentBuilder).content
}

// -----------------------------------------------------------------------------

type textBuilder struct {
	content []anthropic.TextBlockParam
}

func (p *textBuilder) Text(s string) xai.TextBuilder {
	p.content = append(p.content, anthropic.TextBlockParam{Text: s})
	return p
}

func (p *Provider) Texts() xai.TextBuilder {
	return &textBuilder{}
}

func buildTexts(in xai.TextBuilder) []anthropic.TextBlockParam {
	return in.(*textBuilder).content
}

// -----------------------------------------------------------------------------

func (p *Provider) Tools() xai.ToolBuilder {
	panic("todo")
}

// -----------------------------------------------------------------------------
