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
	"encoding/base64"

	"github.com/goplus/xai"
	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

type msgBuilder struct {
	msgs []*genai.Content
}

func (p *msgBuilder) User(content xai.ContentBuilder) xai.MessageBuilder {
	p.msgs = append(p.msgs, &genai.Content{
		Parts: buildContents(content),
		Role:  genai.RoleUser,
	})
	return p
}

func (p *msgBuilder) Assistant(content xai.ContentBuilder) xai.MessageBuilder {
	p.msgs = append(p.msgs, &genai.Content{
		Parts: buildContents(content),
		Role:  genai.RoleModel,
	})
	return p
}

func (p *Provider) Messages() xai.MessageBuilder {
	return &msgBuilder{}
}

func buildMessages(in xai.MessageBuilder) []*genai.Content {
	return in.(*msgBuilder).msgs
}

// -----------------------------------------------------------------------------

type contentBuilder struct {
	content []*genai.Part
	lastErr error
}

func (p *contentBuilder) Text(text string) xai.ContentBuilder {
	p.content = append(p.content, genai.NewPartFromText(text))
	return p
}

func (p *contentBuilder) Image(mime xai.ImageType, data []byte) xai.ContentBuilder {
	p.content = append(p.content, genai.NewPartFromBytes(
		data, string(mime),
	))
	return p
}

func (p *contentBuilder) ImageBase64(mime xai.ImageType, data string) xai.ContentBuilder {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		p.lastErr = err
	} else {
		p.content = append(p.content, genai.NewPartFromBytes(
			b, string(mime),
		))
	}
	return p
}

func (p *contentBuilder) ImageURL(mime xai.ImageType, url string) xai.ContentBuilder {
	p.content = append(p.content, genai.NewPartFromURI(
		url, string(mime),
	))
	return p
}

func (p *contentBuilder) DocText(text string) xai.ContentBuilder {
	p.content = append(p.content, genai.NewPartFromText(text))
	return p
}

func (p *contentBuilder) DocPDFURL(url string) xai.ContentBuilder {
	p.content = append(p.content, genai.NewPartFromURI(
		url, "application/pdf",
	))
	return p
}

func (p *contentBuilder) DocPDFBase64(data string) xai.ContentBuilder {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		p.lastErr = err
	} else {
		p.content = append(p.content, genai.NewPartFromBytes(
			b, "application/pdf",
		))
	}
	return p
}

func (p *contentBuilder) DocMultipart(multi xai.MultipartBuilder) xai.ContentBuilder {
	panic("todo")
}

func (p *contentBuilder) SearchResult(content xai.TextBuilder, source, title string) xai.ContentBuilder {
	panic("todo")
}

func (p *contentBuilder) Thinking(signature, thinking string) xai.ContentBuilder {
	panic("todo")
}

func (p *contentBuilder) RedactedThinking(data string) xai.ContentBuilder {
	panic("todo")
}

func (p *contentBuilder) ToolUse(id string, input any, name string) xai.ContentBuilder {
	// TODO(xsw): name as toolUseID
	p.content = append(p.content, genai.NewPartFromFunctionCall(name, input.(map[string]any)))
	return p
}

func (p *contentBuilder) ToolResult(toolUseID string, content any, isError bool) xai.ContentBuilder {
	// TODO(xsw): validate content
	p.content = append(p.content, genai.NewPartFromFunctionResponse(toolUseID, content.(map[string]any)))
	return p
}

func (p *contentBuilder) ServerToolUse(id string, input any, name xai.ServerToolName) xai.ContentBuilder {
	panic("todo")
}

func (p *Provider) Contents() xai.ContentBuilder {
	return &contentBuilder{}
}

func buildContents(in xai.ContentBuilder) []*genai.Part {
	return in.(*contentBuilder).content
}

// -----------------------------------------------------------------------------

func (p *Provider) Parts() xai.MultipartBuilder {
	panic("todo")
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
