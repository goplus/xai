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

package kling

import (
	"context"
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling/image"
	"github.com/goplus/xai/spec/kling/video"
)

// -----------------------------------------------------------------------------

func (p *Service) Actions(model xai.Model) []xai.Action {
	m := string(model)
	var actions []xai.Action
	if IsImageModel(m) {
		actions = append(actions, xai.GenImage)
	}
	if IsVideoModel(m) {
		actions = append(actions, xai.GenVideo)
	}
	return actions
}

func (p *Service) Operation(model xai.Model, action xai.Action) (op xai.Operation, err error) {
	m := string(model)
	switch action {
	case xai.GenImage:
		if !IsImageModel(m) {
			return nil, xai.ErrNotFound
		}
		op = &genImage{model: m}
	case xai.GenVideo:
		if !IsVideoModel(m) {
			return nil, xai.ErrNotFound
		}
		op = &genVideo{model: m}
	default:
		err = xai.ErrNotFound
	}
	return
}

// -----------------------------------------------------------------------------

type genImage struct {
	model string
	params *Params
}

func (p *genImage) InputSchema() xai.InputSchema {
	return &inputSchema{model: p.model, isVideo: false, fields: image.SchemaForImage(p.model)}
}

func (p *genImage) Params() xai.Params {
	if p.params == nil {
		p.params = NewParams()
	}
	return p.params
}

func (p *genImage) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (xai.OperationResponse, error) {
	s := svc.(*Service)
	if s.imgExec == nil {
		return nil, xai.ErrNotFound
	}
	params := p.Params().(*Params)
	if !params.HasNonEmptyString(ParamPrompt) {
		return nil, ErrPromptRequired
	}
	schema := p.InputSchema()
	if err := validateParamsAgainstRestriction(schema, params); err != nil {
		return nil, err
	}
	if o, ok := opts.(*Options); ok && o.UserID != "" {
		params.Set("_user_id", o.UserID)
	}
	return s.imgExec.Submit(ctx, xai.Model(p.model), params)
}

// -----------------------------------------------------------------------------

type genVideo struct {
	model string
	params *Params
}

func (p *genVideo) InputSchema() xai.InputSchema {
	return &inputSchema{model: p.model, isVideo: true, fields: video.SchemaForVideo(p.model)}
}

func (p *genVideo) Params() xai.Params {
	if p.params == nil {
		p.params = NewParams()
	}
	return p.params
}

func (p *genVideo) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (xai.OperationResponse, error) {
	s := svc.(*Service)
	if s.vidExec == nil {
		return nil, xai.ErrNotFound
	}
	params := p.Params().(*Params)
	if !params.HasNonEmptyString(ParamPrompt) {
		return nil, ErrPromptRequired
	}
	if err := video.Validate(p.model, params); err != nil {
		return nil, err
	}
	schema := p.InputSchema()
	if err := validateParamsAgainstRestriction(schema, params); err != nil {
		return nil, err
	}
	if o, ok := opts.(*Options); ok && o.UserID != "" {
		params.Set("_user_id", o.UserID)
	}
	return s.vidExec.Submit(ctx, xai.Model(p.model), params)
}

// inputSchema implements xai.InputSchema.
type inputSchema struct {
	model   string
	isVideo bool
	fields  []xai.Field
}

func (s *inputSchema) Fields() []xai.Field { return s.fields }

func (s *inputSchema) Restrict(name string) *xai.Restriction {
	if s.isVideo {
		return video.Restrict(s.model, name)
	}
	return image.Restrict(s.model, name)
}

// validateParamsAgainstRestriction checks each set string param against its Restriction.
func validateParamsAgainstRestriction(schema xai.InputSchema, params *Params) error {
	for name, val := range params.Export() {
		if strings.HasPrefix(name, "_") {
			continue
		}
		s, ok := val.(string)
		if !ok || strings.TrimSpace(s) == "" {
			continue
		}
		r := schema.Restrict(name)
		if err := r.ValidateString(name, s); err != nil {
			return err
		}
	}
	return nil
}

// GetTask returns an OperationResponse for an existing task by taskID.
// Use when resuming from DB. Routes to imgExec or vidExec based on action.
func (p *Service) GetTask(ctx context.Context, model xai.Model, action xai.Action, taskID string) (xai.OperationResponse, error) {
	switch action {
	case xai.GenImage:
		if p.imgExec == nil {
			return nil, xai.ErrNotFound
		}
		return p.imgExec.GetTaskStatus(ctx, taskID)
	case xai.GenVideo:
		if p.vidExec == nil {
			return nil, xai.ErrNotFound
		}
		return p.vidExec.GetTaskStatus(ctx, taskID)
	default:
		return nil, xai.ErrNotFound
	}
}
