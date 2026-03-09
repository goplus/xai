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
	"encoding/json"
	"fmt"

	xai "github.com/goplus/xai/spec"
)

// -----------------------------------------------------------------------------

type msgBuilder struct {
	msg *message
}

func (p *Service) UserMsg() xai.MsgBuilder {
	return &msgBuilder{msg: &message{Role: "user"}}
}

func (p *Service) AssistantMsg() xai.MsgBuilder {
	return &msgBuilder{msg: &message{Role: "assistant"}}
}

// MsgBuilderExt extends MsgBuilder with provider-specific methods.
// Use UserMsgExt() when you need image detail or video file input.
type MsgBuilderExt interface {
	xai.MsgBuilder
	TextExt(text string) MsgBuilderExt
	ImageURLExt(mime xai.ImageType, url string) MsgBuilderExt
	ImageURLWithDetail(url string, detail string) MsgBuilderExt
	VideoFile(fileID string, format string) MsgBuilderExt
}

// UserMsgExt returns a MsgBuilderExt for building user messages with image/video content.
func (p *Service) UserMsgExt() MsgBuilderExt {
	return msgBuilderExt{&msgBuilder{msg: &message{Role: "user"}}}
}

type msgBuilderExt struct {
	*msgBuilder
}

func (p msgBuilderExt) TextExt(text string) MsgBuilderExt {
	p.msgBuilder.Text(text)
	return p
}

func (p msgBuilderExt) ImageURLExt(mime xai.ImageType, url string) MsgBuilderExt {
	p.msgBuilder.ImageURL(mime, url)
	return p
}

func (p msgBuilderExt) ImageURLWithDetail(url string, detail string) MsgBuilderExt {
	p.msgBuilder.ImageURLWithDetail(url, detail)
	return p
}

func (p msgBuilderExt) VideoFile(fileID string, format string) MsgBuilderExt {
	p.msgBuilder.VideoFile(fileID, format)
	return p
}

func (p *msgBuilder) addContent(c *content) xai.MsgBuilder {
	p.msg.Contents = append(p.msg.Contents, c)
	return p
}

func (p *msgBuilder) Text(text string) xai.MsgBuilder {
	return p.addContent(&content{Type: contentText, Text: text})
}

func (p *msgBuilder) Image(image xai.ImageData) xai.MsgBuilder {
	// TODO: not yet implemented; use ImageURL or ImageFile instead
	return p
}

func (p *msgBuilder) ImageURL(mime xai.ImageType, url string) xai.MsgBuilder {
	return p.addContent(&content{Type: contentImageURL, ImageURL: url})
}

func (p *msgBuilder) ImageURLWithDetail(url string, detail string) xai.MsgBuilder {
	return p.addContent(&content{Type: contentImageURL, ImageURL: url, ImageDetail: detail})
}

func (p *msgBuilder) ImageFile(mime xai.ImageType, fileID string) xai.MsgBuilder {
	return p.addContent(&content{Type: contentImageFile, FileID: fileID})
}

func (p *msgBuilder) Doc(doc xai.DocumentData) xai.MsgBuilder {
	// TODO: not yet implemented; use DocURL or DocFile instead
	return p
}

func (p *msgBuilder) DocURL(mime xai.DocumentType, url string) xai.MsgBuilder {
	return p.addContent(&content{Type: contentDocURL, FileURL: url, FileMIME: string(mime)})
}

func (p *msgBuilder) DocFile(mime xai.DocumentType, fileID string) xai.MsgBuilder {
	return p.addContent(&content{Type: contentDocFile, FileID: fileID, FileMIME: string(mime)})
}

func (p *msgBuilder) VideoFile(fileID string, format string) xai.MsgBuilder {
	return p.DocFile(xai.DocumentType(format), fileID)
}

func (p *msgBuilder) Part(part xai.Part) xai.MsgBuilder {
	// TODO: not yet implemented
	return p
}

func (p *msgBuilder) Thinking(v xai.Thinking) xai.MsgBuilder {
	return p.addContent(&content{
		Type: contentThinking,
		Thinking: &thinkingContent{
			Text:      v.Text,
			Signature: v.Signature,
			Redacted:  v.Redacted,
		},
	})
}

func (p *msgBuilder) Compaction(data string) xai.MsgBuilder {
	return p.addContent(&content{Type: contentCompaction, Compaction: data})
}

func (p *msgBuilder) ToolUse(v xai.ToolUse) xai.MsgBuilder {
	input, err := json.Marshal(v.Input)
	if err != nil {
		panic("invalid tool input: " + err.Error())
	}
	return p.addContent(&content{
		Type: contentToolUse,
		ToolUse: &toolUseContent{
			ID:    v.ID,
			Name:  v.Name,
			Input: input,
		},
	})
}

func (p *msgBuilder) ToolResult(v xai.ToolResult) xai.MsgBuilder {
	var result any = v.Result
	if v.IsError && v.Result != nil {
		if err, ok := v.Result.(error); ok {
			result = map[string]any{"error": err.Error()}
		} else {
			result = map[string]any{"error": fmt.Sprintf("%v", v.Result)}
		}
	}
	resultJSON, err := json.Marshal(result)
	if err != nil {
		panic("invalid tool result: " + err.Error())
	}
	// Tool results create a new "tool" role message
	p.msg.Role = "tool"
	return p.addContent(&content{
		Type: contentToolResult,
		ToolResult: &toolResultContent{
			ID:      v.ID,
			Name:    v.Name,
			Result:  resultJSON,
			IsError: v.IsError,
		},
	})
}

// getMessage returns the internal message for building requests.
func (p *msgBuilder) getMessage() *message {
	return p.msg
}

// -----------------------------------------------------------------------------

type textBuilder struct {
	texts []textContent
}

func (p *textBuilder) Text(text string) xai.TextBuilder {
	p.texts = append(p.texts, textContent{Text: text})
	return p
}

func (p *Service) Texts(texts ...string) xai.TextBuilder {
	tb := &textBuilder{}
	for _, t := range texts {
		tb.texts = append(tb.texts, textContent{Text: t})
	}
	return tb
}

func buildTexts(in xai.TextBuilder) []textContent {
	return in.(*textBuilder).texts
}

// -----------------------------------------------------------------------------
