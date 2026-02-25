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
	"encoding/base64"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
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

func (p *contentBuilder) Image(mime xai.ImageType, data []byte) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaImageBlock(
		anthropic.BetaBase64ImageSourceParam{
			Data:      base64.StdEncoding.EncodeToString(data),
			MediaType: anthropic.BetaBase64ImageSourceMediaType(mime),
		},
	))
	return p
}

func (p *contentBuilder) ImageBase64(mime xai.ImageType, base64 string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaImageBlock(
		anthropic.BetaBase64ImageSourceParam{
			Data:      base64,
			MediaType: anthropic.BetaBase64ImageSourceMediaType(mime),
		},
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

func (p *contentBuilder) DocText(text string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaDocumentBlock(anthropic.BetaPlainTextSourceParam{
		Data: text,
	}))
	return p
}

func (p *contentBuilder) DocPDFURL(url string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaDocumentBlock(anthropic.BetaURLPDFSourceParam{
		URL: url,
	}))
	return p
}

func (p *contentBuilder) DocPDFBase64(base64 string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaDocumentBlock(anthropic.BetaBase64PDFSourceParam{
		Data: base64,
	}))
	return p
}

func (p *contentBuilder) DocMultipart(multi xai.MultipartBuilder) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaDocumentBlock(anthropic.BetaContentBlockSourceParam{
		Content: anthropic.BetaContentBlockSourceContentUnionParam{
			OfBetaContentBlockSourceContent: buildMultipart(multi),
		},
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

type multipartBuilder struct {
	content []anthropic.BetaContentBlockSourceContentUnionParam
}

func (p *multipartBuilder) Text(text string) xai.MultipartBuilder {
	p.content = append(p.content, anthropic.BetaContentBlockSourceContentUnionParam{
		OfString: param.NewOpt(text),
	})
	return p
}

func (p *multipartBuilder) ImageURL(url string) xai.MultipartBuilder {
	panic("todo")
	/* p.content = append(p.content, anthropic.BetaContentBlockSourceContentUnionParam{
		OfImage: &anthropic.ImageBlockParam{
			Source: anthropic.ImageBlockParamSourceUnion{
				OfURL: &anthropic.URLImageSourceParam{
					URL: url,
				},
			},
		},
	})
	return p */
}

func (p *multipartBuilder) ImageBase64(mime xai.ImageType, base64 string) xai.MultipartBuilder {
	panic("todo")
	/* p.content = append(p.content, anthropic.BetaContentBlockSourceContentUnionParam{
		OfImage: &anthropic.ImageBlockParam{
			Source: anthropic.ImageBlockParamSourceUnion{
				OfBase64: &anthropic.Base64ImageSourceParam{
					MediaType: anthropic.Base64ImageSourceMediaType(mime),
					Data:      base64,
				},
			},
		},
	})
	return p */
}

func (p *Provider) Parts() xai.MultipartBuilder {
	return &multipartBuilder{}
}

func buildMultipart(in xai.MultipartBuilder) []anthropic.BetaContentBlockSourceContentUnionParam {
	return in.(*multipartBuilder).content
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
