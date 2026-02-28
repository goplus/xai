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
	msgs []*genai.Content
}

func (p *msgBuilder) User(content xai.ContentBuilder) xai.MessageBuilder {
	parts := buildContents(content)
	p.msgs = append(p.msgs, &genai.Content{
		Parts: parts,
		Role:  genai.RoleUser,
	})
	return p
}

func (p *msgBuilder) Assistant(content xai.ContentBuilder) xai.MessageBuilder {
	parts := buildContents(content)
	p.msgs = append(p.msgs, &genai.Content{
		Parts: parts,
		Role:  genai.RoleModel,
	})
	return p
}

func (p *Provider) Messages() xai.MessageBuilder {
	return &msgBuilder{}
}

func buildMessages(in xai.MessageBuilder) []*genai.Content {
	p := in.(*msgBuilder)
	return p.msgs
}

// -----------------------------------------------------------------------------

type contentBuilder struct {
	content []*genai.Part
}

func (p *contentBuilder) Text(text string) xai.ContentBuilder {
	p.content = append(p.content, genai.NewPartFromText(text))
	return p
}

func (p *contentBuilder) Image(image xai.ImageData) xai.ContentBuilder {
	p.content = append(p.content, &genai.Part{
		InlineData: (*genai.Blob)(image.(*imageData)),
	})
	return p
}

func (p *contentBuilder) ImageURL(mime xai.ImageType, url string) xai.ContentBuilder {
	p.content = append(p.content, genai.NewPartFromURI(
		url, string(mime),
	))
	return p
}

func (p *contentBuilder) ImageFile(mime xai.ImageType, fileID string) xai.ContentBuilder {
	p.content = append(p.content, genai.NewPartFromURI(
		fileID, string(mime),
	))
	return p
}

func (p *contentBuilder) Doc(doc xai.DocumentData) xai.ContentBuilder {
	p.content = append(p.content, (*genai.Part)(doc.(*docData)))
	return p
}

func (p *contentBuilder) DocURL(mime xai.DocumentType, url string) xai.ContentBuilder {
	p.content = append(p.content, genai.NewPartFromURI(
		url, string(mime),
	))
	return p
}

func (p *contentBuilder) DocFile(mime xai.DocumentType, fileID string) xai.ContentBuilder {
	p.content = append(p.content, genai.NewPartFromURI(
		fileID, string(mime),
	))
	return p
}

func (p *contentBuilder) Thinking(signature, thinking string) xai.ContentBuilder {
	p.content = append(p.content, &genai.Part{
		Text:             thinking,
		ThoughtSignature: unsafe.Slice(unsafe.StringData(signature), len(signature)),
		Thought:          true,
	})
	return p
}

func (p *contentBuilder) RedactedThinking(data string) xai.ContentBuilder {
	// TODO(xsw): validate data
	return p
}

func (p *Provider) Contents() xai.ContentBuilder {
	return &contentBuilder{}
}

func buildContents(in xai.ContentBuilder) []*genai.Part {
	p := in.(*contentBuilder)
	return p.content
}

// -----------------------------------------------------------------------------

type textBuilder struct {
	parts []*genai.Part
}

func (p *textBuilder) Text(text string) xai.TextBuilder {
	p.parts = append(p.parts, genai.NewPartFromText(text))
	return p
}

func (p *Provider) Texts() xai.TextBuilder {
	return &textBuilder{}
}

func buildTexts(in xai.TextBuilder) *genai.Content {
	// SystemInstruction set Role to "system" by default, so we don't need to set it here.
	return &genai.Content{
		Parts: in.(*textBuilder).parts,
	}
}

// -----------------------------------------------------------------------------

func (p *Provider) Tools() xai.ToolBuilder {
	panic("todo")
}

// -----------------------------------------------------------------------------
