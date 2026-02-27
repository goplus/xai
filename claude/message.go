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
	msgs []anthropic.BetaMessageParam
}

func (p *msgBuilder) User(content xai.ContentBuilder) xai.MessageBuilder {
	p.msgs = append(p.msgs, anthropic.NewBetaUserMessage(buildContents(content)...))
	return p
}

func (p *msgBuilder) Assistant(content xai.ContentBuilder) xai.MessageBuilder {
	p.msgs = append(p.msgs, anthropic.BetaMessageParam{
		Role:    anthropic.BetaMessageParamRoleAssistant,
		Content: buildContents(content),
	})
	return p
}

func (p *Provider) Messages() xai.MessageBuilder {
	return &msgBuilder{}
}

func buildMessages(in xai.MessageBuilder) []anthropic.BetaMessageParam {
	return in.(*msgBuilder).msgs
}

// -----------------------------------------------------------------------------

type contentBuilder struct {
	content []anthropic.BetaContentBlockParamUnion
}

func (p *contentBuilder) Text(text string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaTextBlock(text))
	return p
}

func (p *contentBuilder) Image(image xai.ImageData) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaImageBlock(
		*(*anthropic.BetaBase64ImageSourceParam)(image.(*imageData)),
	))
	return p
}

func (p *contentBuilder) ImageURL(mime xai.ImageType, url string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaImageBlock(
		anthropic.BetaURLImageSourceParam{
			URL: url,
		},
	))
	return p
}

func (p *contentBuilder) ImageFile(mime xai.ImageType, fileID string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaImageBlock(
		anthropic.BetaFileImageSourceParam{
			FileID: fileID,
		},
	))
	return p
}

func (p *contentBuilder) Doc(doc xai.DocumentData) xai.ContentBuilder {
	p.content = append(p.content, (doc.(*docData).data))
	return p
}

func (p *contentBuilder) DocURL(mime xai.DocumentType, url string) xai.ContentBuilder {
	if mime != xai.DocPDF {
		panic("todo")
	}
	p.content = append(p.content, anthropic.NewBetaDocumentBlock(anthropic.BetaURLPDFSourceParam{
		URL: url,
	}))
	return p
}

func (p *contentBuilder) DocFile(mime xai.DocumentType, fileID string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaDocumentBlock(anthropic.BetaFileDocumentSourceParam{
		FileID: fileID,
	}))
	return p
}

func (p *contentBuilder) SearchResult(content xai.TextBuilder, source, title string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaSearchResultBlock(buildTexts(content), source, title))
	return p
}

func (p *contentBuilder) Thinking(signature, thinking string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaThinkingBlock(signature, thinking))
	return p
}

func (p *contentBuilder) RedactedThinking(data string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaRedactedThinkingBlock(data))
	return p
}

func (p *contentBuilder) ToolUse(id string, input any, name string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaToolUseBlock(id, input, name))
	return p
}

func (p *contentBuilder) ToolResult(toolUseID string, content any, isError bool) xai.ContentBuilder {
	// TODO(xsw): validate content
	p.content = append(p.content, anthropic.NewBetaToolResultBlock(toolUseID, content.(string), isError))
	return p
}

func (p *contentBuilder) ServerToolUse(id string, input any, name xai.ServerToolName) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaServerToolUseBlock(id, input, anthropic.BetaServerToolUseBlockParamName(name)))
	return p
}

func (p *Provider) Contents() xai.ContentBuilder {
	return &contentBuilder{}
}

func buildContents(in xai.ContentBuilder) []anthropic.BetaContentBlockParamUnion {
	return in.(*contentBuilder).content
}

// -----------------------------------------------------------------------------

type textBuilder struct {
	content []anthropic.BetaTextBlockParam
}

func (p *textBuilder) Text(text string) xai.TextBuilder {
	p.content = append(p.content, anthropic.BetaTextBlockParam{Text: text})
	return p
}

func (p *Provider) Texts() xai.TextBuilder {
	return &textBuilder{}
}

func buildTexts(in xai.TextBuilder) []anthropic.BetaTextBlockParam {
	return in.(*textBuilder).content
}

// -----------------------------------------------------------------------------

func (p *Provider) Tools() xai.ToolBuilder {
	panic("todo")
}

// -----------------------------------------------------------------------------
