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
	xai "github.com/goplus/xai/spec"
)

// -----------------------------------------------------------------------------

type msgBuilder struct {
	content []anthropic.BetaContentBlockParamUnion
	role    anthropic.BetaMessageParamRole
}

func buildMessages(msgs []xai.MsgBuilder) []anthropic.BetaMessageParam {
	ret := make([]anthropic.BetaMessageParam, len(msgs))
	for i, msg := range msgs {
		m := msg.(*msgBuilder)
		ret[i] = anthropic.BetaMessageParam{
			Content: m.content,
			Role:    m.role,
		}
	}
	return ret
}

func (p *Service) UserMsg() xai.MsgBuilder {
	return &msgBuilder{role: anthropic.BetaMessageParamRoleUser}
}

func (p *Service) AssistantMsg() xai.MsgBuilder {
	return &msgBuilder{role: anthropic.BetaMessageParamRoleAssistant}
}

func (p *msgBuilder) Text(text string) xai.MsgBuilder {
	p.content = append(p.content, anthropic.NewBetaTextBlock(text))
	return p
}

func (p *msgBuilder) Image(image xai.ImageData) xai.MsgBuilder {
	p.content = append(p.content, anthropic.NewBetaImageBlock(
		*(*anthropic.BetaBase64ImageSourceParam)(image.(*imageData)),
	))
	return p
}

func (p *msgBuilder) ImageURL(mime xai.ImageType, url string) xai.MsgBuilder {
	p.content = append(p.content, anthropic.NewBetaImageBlock(
		anthropic.BetaURLImageSourceParam{
			URL: url,
		},
	))
	return p
}

func (p *msgBuilder) ImageFile(mime xai.ImageType, fileID string) xai.MsgBuilder {
	p.content = append(p.content, anthropic.NewBetaImageBlock(
		anthropic.BetaFileImageSourceParam{
			FileID: fileID,
		},
	))
	return p
}

func (p *msgBuilder) Doc(doc xai.DocumentData) xai.MsgBuilder {
	p.content = append(p.content, (doc.(*docData).data))
	return p
}

func (p *msgBuilder) DocURL(mime xai.DocumentType, url string) xai.MsgBuilder {
	if mime != xai.DocPDF {
		panic("todo")
	}
	p.content = append(p.content, anthropic.NewBetaDocumentBlock(anthropic.BetaURLPDFSourceParam{
		URL: url,
	}))
	return p
}

func (p *msgBuilder) DocFile(mime xai.DocumentType, fileID string) xai.MsgBuilder {
	p.content = append(p.content, anthropic.NewBetaDocumentBlock(anthropic.BetaFileDocumentSourceParam{
		FileID: fileID,
	}))
	return p
}

func (p *msgBuilder) Part(part xai.Part) xai.MsgBuilder {
	p.content = append(p.content, buildPart(part))
	return p
}

func (p *msgBuilder) Thinking(v xai.Thinking) xai.MsgBuilder {
	var content anthropic.BetaContentBlockParamUnion
	if v.Redacted {
		content = anthropic.NewBetaThinkingBlock(v.Signature, v.Text)
	} else {
		content = anthropic.NewBetaRedactedThinkingBlock(v.Signature)
	}
	p.content = append(p.content, content)
	return p
}

func (p *msgBuilder) Compaction(data string) xai.MsgBuilder {
	p.content = append(p.content, anthropic.NewBetaCompactionBlock(data))
	return p
}

// -----------------------------------------------------------------------------

type textBuilder struct {
	content []anthropic.BetaTextBlockParam
}

func (p *textBuilder) Text(text string) xai.TextBuilder {
	p.content = append(p.content, anthropic.BetaTextBlockParam{Text: text})
	return p
}

func (p *Service) Texts(texts ...string) xai.TextBuilder {
	var content []anthropic.BetaTextBlockParam
	if len(texts) > 0 {
		content = make([]anthropic.BetaTextBlockParam, len(texts))
		for i, text := range texts {
			content[i].Text = text
		}
	}
	return &textBuilder{content: content}
}

func buildTexts(in xai.TextBuilder) []anthropic.BetaTextBlockParam {
	return in.(*textBuilder).content
}

// -----------------------------------------------------------------------------
