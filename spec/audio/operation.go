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

package audio

import (
	"context"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/audio/internal"
)

var (
	ErrAudioRequired = internal.ErrAudioRequired
	ErrInputRequired = internal.ErrInputRequired
)

func (p *Service) Actions(model xai.Model) []xai.Action {
	m := string(model)
	var actions []xai.Action
	if IsASRModel(m) {
		actions = append(actions, xai.Transcribe)
	}
	if IsTTSModel(m) {
		actions = append(actions, xai.Synthesize)
	}
	return actions
}

func (p *Service) Operation(model xai.Model, action xai.Action) (op xai.Operation, err error) {
	m := string(model)
	switch action {
	case xai.Transcribe:
		if !IsASRModel(m) {
			return nil, xai.ErrNotFound
		}
		op = &genTranscribe{model: m}
	case xai.Synthesize:
		if !IsTTSModel(m) {
			return nil, xai.ErrNotFound
		}
		op = &genSynthesize{model: m}
	default:
		err = xai.ErrNotFound
	}
	return
}

// -----------------------------------------------------------------------------

type genTranscribe struct {
	model  string
	params *Params
}

func (p *genTranscribe) InputSchema() xai.InputSchema {
	return &inputSchema{model: p.model, isTTS: false, fields: SchemaForTranscribe(p.model)}
}

func (p *genTranscribe) Params() xai.Params {
	if p.params == nil {
		p.params = NewParams()
	}
	return p.params
}

func (p *genTranscribe) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (xai.OperationResponse, error) {
	s := svc.(*Service)
	if s.asrExec == nil {
		return nil, xai.ErrNotFound
	}
	params := p.Params().(*Params)
	if !hasAudio(params) {
		return nil, ErrAudioRequired
	}
	if o, ok := opts.(*Options); ok && o.UserID != "" {
		params.Set("_user_id", o.UserID)
	}
	return s.asrExec.Transcribe(ctx, xai.Model(p.model), params)
}

// -----------------------------------------------------------------------------

type genSynthesize struct {
	model  string
	params *Params
}

func (p *genSynthesize) InputSchema() xai.InputSchema {
	return &inputSchema{model: p.model, isTTS: true, fields: SchemaForSynthesize(p.model)}
}

func (p *genSynthesize) Params() xai.Params {
	if p.params == nil {
		p.params = NewParams()
	}
	return p.params
}

func (p *genSynthesize) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (xai.OperationResponse, error) {
	s := svc.(*Service)
	if s.ttsExec == nil {
		return nil, xai.ErrNotFound
	}
	params := p.Params().(*Params)
	if !params.HasNonEmptyString(ParamInput) {
		return nil, ErrInputRequired
	}
	if o, ok := opts.(*Options); ok && o.UserID != "" {
		params.Set("_user_id", o.UserID)
	}
	return s.ttsExec.Synthesize(ctx, xai.Model(p.model), params)
}

// inputSchema implements xai.InputSchema.
type inputSchema struct {
	model  string
	isTTS  bool
	fields []xai.Field
}

func (s *inputSchema) Fields() []xai.Field { return s.fields }

func (s *inputSchema) Restrict(name string) *xai.Restriction {
	return nil
}

// hasAudio returns true if params has valid audio input (URL or {format, url}).
func hasAudio(p *Params) bool {
	v, ok := p.Get(internal.ParamAudio)
	if !ok {
		return false
	}
	switch a := v.(type) {
	case string:
		return len(a) > 0
	case map[string]interface{}:
		urlVal, _ := a["url"].(string)
		return len(urlVal) > 0
	case map[string]string:
		return len(a["url"]) > 0
	}
	return false
}
