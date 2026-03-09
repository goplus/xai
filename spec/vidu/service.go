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

package vidu

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"iter"
	"os"

	xai "github.com/goplus/xai/spec"
)

var errGenNotSupported = errors.New("vidu: Gen/GenStream not supported, use Operation for video generation")

// Options implements xai.OptionBuilder for executor implementations.
type Options struct {
	BaseURL string
	UserID  string
}

func (p *Options) WithBaseURL(base string) xai.OptionBuilder  { p.BaseURL = base; return p }
func (p *Options) WithUserID(userID string) xai.OptionBuilder { p.UserID = userID; return p }

// Service implements xai.Service for Vidu video generation.
type Service struct {
	backend Backend
	tools   map[string]xai.Tool
}

// Option configures the Service.
type Option func(*Service)

// NewWithBackend creates a Vidu Service with a custom backend implementation.
func NewWithBackend(backend Backend, opts ...Option) *Service {
	if backend == nil {
		panic("vidu: nil backend")
	}
	s := &Service{
		backend: backend,
		tools:   make(map[string]xai.Tool),
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// NewService creates a Vidu Service (alias for NewWithBackend for backward compatibility).
func NewService(backend Backend, opts ...Option) *Service {
	return NewWithBackend(backend, opts...)
}

func (p *Service) Features() xai.Feature { return xai.FeatureOperation }

func (p *Service) Gen(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) (xai.GenResponse, error) {
	return nil, errGenNotSupported
}

func (p *Service) GenStream(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) iter.Seq2[xai.GenResponse, error] {
	return func(yield func(xai.GenResponse, error) bool) {
		yield(nil, errGenNotSupported)
	}
}

func (p *Service) Options() xai.OptionBuilder {
	return &Options{}
}

func (p *Service) Params() xai.ParamBuilder { return &noopParamBuilder{} }

type noopParamBuilder struct{}

func (p *noopParamBuilder) System(xai.TextBuilder) xai.ParamBuilder     { return p }
func (p *noopParamBuilder) Messages(...xai.MsgBuilder) xai.ParamBuilder { return p }
func (p *noopParamBuilder) Tools(...xai.ToolBase) xai.ParamBuilder      { return p }
func (p *noopParamBuilder) Model(xai.Model) xai.ParamBuilder            { return p }
func (p *noopParamBuilder) MaxOutputTokens(int64) xai.ParamBuilder      { return p }
func (p *noopParamBuilder) Compact(int64) xai.ParamBuilder              { return p }
func (p *noopParamBuilder) Container(string) xai.ParamBuilder           { return p }
func (p *noopParamBuilder) InferenceGeo(string) xai.ParamBuilder        { return p }
func (p *noopParamBuilder) Temperature(float64) xai.ParamBuilder        { return p }
func (p *noopParamBuilder) TopK(int64) xai.ParamBuilder                 { return p }
func (p *noopParamBuilder) TopP(float64) xai.ParamBuilder               { return p }

type noopImageBuilder struct{}

func (p noopImageBuilder) From(mime xai.ImageType, displayName string, src io.Reader) (xai.ImageData, error) {
	return nil, errGenNotSupported
}
func (p noopImageBuilder) FromLocal(mime xai.ImageType, fileName string) (xai.ImageData, error) {
	return nil, errGenNotSupported
}
func (p noopImageBuilder) FromBase64(mime xai.ImageType, displayName string, base64 string) (xai.ImageData, error) {
	return nil, errGenNotSupported
}
func (p noopImageBuilder) FromBytes(mime xai.ImageType, displayName string, data []byte) xai.ImageData {
	return nil
}

func (p *Service) Images() xai.ImageBuilder { return noopImageBuilder{} }

type noopDocBuilder struct{}

func (p noopDocBuilder) From(mime xai.DocumentType, displayName string, src io.Reader) (xai.DocumentData, error) {
	return nil, errGenNotSupported
}
func (p noopDocBuilder) FromLocal(mime xai.DocumentType, fileName string) (xai.DocumentData, error) {
	return nil, errGenNotSupported
}
func (p noopDocBuilder) FromBase64(mime xai.DocumentType, displayName string, base64 string) (xai.DocumentData, error) {
	return nil, errGenNotSupported
}
func (p noopDocBuilder) FromBytes(mime xai.DocumentType, displayName string, data []byte) xai.DocumentData {
	return nil
}
func (p noopDocBuilder) PlainText(text string) xai.DocumentData { return nil }

func (p *Service) Docs() xai.DocumentBuilder { return noopDocBuilder{} }

type noopTextBuilder struct{}

func (p noopTextBuilder) Text(text string) xai.TextBuilder { return p }

func (p *Service) Texts(texts ...string) xai.TextBuilder { return noopTextBuilder{} }

type noopMsgBuilder struct{}

func (p noopMsgBuilder) Text(text string) xai.MsgBuilder                             { return p }
func (p noopMsgBuilder) Image(image xai.ImageData) xai.MsgBuilder                    { return p }
func (p noopMsgBuilder) ImageURL(mime xai.ImageType, url string) xai.MsgBuilder      { return p }
func (p noopMsgBuilder) ImageFile(mime xai.ImageType, fileID string) xai.MsgBuilder  { return p }
func (p noopMsgBuilder) Doc(doc xai.DocumentData) xai.MsgBuilder                     { return p }
func (p noopMsgBuilder) DocURL(mime xai.DocumentType, url string) xai.MsgBuilder     { return p }
func (p noopMsgBuilder) DocFile(mime xai.DocumentType, fileID string) xai.MsgBuilder { return p }
func (p noopMsgBuilder) Part(part xai.Part) xai.MsgBuilder                           { return p }
func (p noopMsgBuilder) Thinking(t xai.Thinking) xai.MsgBuilder                      { return p }
func (p noopMsgBuilder) ToolUse(v xai.ToolUse) xai.MsgBuilder                        { return p }
func (p noopMsgBuilder) ToolResult(v xai.ToolResult) xai.MsgBuilder                  { return p }
func (p noopMsgBuilder) Compaction(data string) xai.MsgBuilder                       { return p }

func (p *Service) UserMsg() xai.MsgBuilder      { return noopMsgBuilder{} }
func (p *Service) AssistantMsg() xai.MsgBuilder { return noopMsgBuilder{} }

type noopWebSearchTool struct{}

func (p noopWebSearchTool) UnderlyingAssignTo(any)                     {}
func (p noopWebSearchTool) MaxUses(int64) xai.WebSearchTool            { return p }
func (p noopWebSearchTool) AllowedDomains(...string) xai.WebSearchTool { return p }
func (p noopWebSearchTool) BlockedDomains(...string) xai.WebSearchTool { return p }

func (p *Service) WebSearchTool() xai.WebSearchTool { return noopWebSearchTool{} }

type noopTool struct {
	name string
}

func (p noopTool) UnderlyingAssignTo(any)           {}
func (p noopTool) Description(desc string) xai.Tool { return p }

func (p *Service) ToolDef(name string) xai.Tool {
	if _, ok := p.tools[name]; ok {
		panic("vidu: tool already defined: " + name)
	}
	t := noopTool{name: name}
	p.tools[name] = t
	return t
}

func (p *Service) Tool(name string) xai.Tool { return p.tools[name] }

// objectFactory
func (p *Service) ImageFrom(mime xai.ImageType, src io.Reader) (xai.Image, error) {
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return p.ImageFromBytes(mime, data), nil
}

func (p *Service) ImageFromLocal(mime xai.ImageType, fileName string) (xai.Image, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return p.ImageFromBytes(mime, data), nil
}

func (p *Service) ImageFromBase64(mime xai.ImageType, b64 string) (xai.Image, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}
	return p.ImageFromBytes(mime, data), nil
}

func (p *Service) ImageFromBytes(mime xai.ImageType, data []byte) xai.Image {
	return newImageFromBytes(mime, data)
}

func (p *Service) ImageFromStgUri(mime xai.ImageType, stgUri string) xai.Image {
	return newImageFromURI(mime, stgUri)
}

func (p *Service) VideoFrom(mime xai.VideoType, src io.Reader) (xai.Video, error) {
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return p.VideoFromBytes(mime, data), nil
}

func (p *Service) VideoFromLocal(mime xai.VideoType, fileName string) (xai.Video, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return p.VideoFromBytes(mime, data), nil
}

func (p *Service) VideoFromBase64(mime xai.VideoType, b64 string) (xai.Video, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}
	return p.VideoFromBytes(mime, data), nil
}

func (p *Service) VideoFromBytes(mime xai.VideoType, data []byte) xai.Video {
	return newVideoFromBytes(mime, data)
}

func (p *Service) VideoFromStgUri(mime xai.VideoType, stgUri string) xai.Video {
	return newVideoFromURI(mime, stgUri)
}

func (p *Service) ReferenceImage(img xai.Image, id int32, typ xai.ReferenceImageType) (xai.ReferenceImage, xai.Configurable) {
	return nil, nil
}

func (p *Service) GenVideoReferenceImages(imgs ...xai.GenVideoReferenceImage) xai.GenVideoReferenceImages {
	return nil
}

func (p *Service) GenVideoMask(img xai.Image, maskMode string) xai.GenVideoMask {
	return nil
}
