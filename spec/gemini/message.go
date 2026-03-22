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

	xai "github.com/goplus/xai/spec"
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

func (p *Service) UserMsg() xai.MsgBuilder {
	return &msgBuilder{role: genai.RoleUser}
}

func (p *Service) AssistantMsg() xai.MsgBuilder {
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

func (p *msgBuilder) Part(part xai.Part) xai.MsgBuilder {
	p.content = append(p.content, buildPart(part))
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

func (p *msgBuilder) Compaction(data string) xai.MsgBuilder {
	panic("gemini does not support compaction")
}

// -----------------------------------------------------------------------------
