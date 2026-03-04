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

package gemini

import (
	"unsafe"

	"github.com/goplus/xai"
	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

type msgBuilder struct {
	content []*genai.Part
	role    string
}

func buildMessages(msgs []xai.MsgBuilder) []*genai.Content {
	ret := make([]*genai.Content, len(msgs))
	for i, msg := range msgs {
		m := msg.(*msgBuilder)
		ret[i] = &genai.Content{
			Parts: m.content,
			Role:  m.role,
		}
	}
	return ret
}

func (p *Provider) UserMsg() xai.MsgBuilder {
	return &msgBuilder{role: genai.RoleUser}
}

func (p *Provider) AssistantMsg() xai.MsgBuilder {
	return &msgBuilder{role: genai.RoleModel}
}

func (p *msgBuilder) Text(text string) xai.MsgBuilder {
	p.content = append(p.content, genai.NewPartFromText(text))
	return p
}

func (p *msgBuilder) Image(image xai.ImageData) xai.MsgBuilder {
	p.content = append(p.content, &genai.Part{
		InlineData: (*genai.Blob)(image.(*imageData)),
	})
	return p
}

func (p *msgBuilder) ImageURL(mime xai.ImageType, url string) xai.MsgBuilder {
	p.content = append(p.content, genai.NewPartFromURI(
		url, string(mime),
	))
	return p
}

func (p *msgBuilder) ImageFile(mime xai.ImageType, fileID string) xai.MsgBuilder {
	p.content = append(p.content, genai.NewPartFromURI(
		fileID, string(mime),
	))
	return p
}

func (p *msgBuilder) Doc(doc xai.DocumentData) xai.MsgBuilder {
	p.content = append(p.content, (*genai.Part)(doc.(*docData)))
	return p
}

func (p *msgBuilder) DocURL(mime xai.DocumentType, url string) xai.MsgBuilder {
	p.content = append(p.content, genai.NewPartFromURI(
		url, string(mime),
	))
	return p
}

func (p *msgBuilder) DocFile(mime xai.DocumentType, fileID string) xai.MsgBuilder {
	p.content = append(p.content, genai.NewPartFromURI(
		fileID, string(mime),
	))
	return p
}

func (p *msgBuilder) Thinking(v xai.Thinking) xai.MsgBuilder {
	p.content = append(p.content, &genai.Part{
		Text:             v.Text,
		ThoughtSignature: unsafe.Slice(unsafe.StringData(v.Signature), len(v.Signature)),
		Thought:          true,
	})
	return p
}

// -----------------------------------------------------------------------------

type textBuilder struct {
	parts []*genai.Part
}

func (p *textBuilder) Text(text string) xai.TextBuilder {
	p.parts = append(p.parts, genai.NewPartFromText(text))
	return p
}

func (p *Provider) Texts(texts ...string) xai.TextBuilder {
	var parts []*genai.Part
	if len(texts) > 0 {
		parts = make([]*genai.Part, len(texts))
		for i, text := range texts {
			parts[i] = genai.NewPartFromText(text)
		}
	}
	return &textBuilder{parts: parts}
}

func buildTexts(in xai.TextBuilder) *genai.Content {
	// SystemInstruction set Role to "system" by default, so we don't need to set it here.
	return &genai.Content{
		Parts: in.(*textBuilder).parts,
	}
}

// -----------------------------------------------------------------------------
