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
	"strings"

	xai "github.com/goplus/xai/spec"
)

func (p *Service) Actions(model xai.Model) []xai.Action {
	if IsVideoModel(string(model)) {
		return []xai.Action{xai.GenVideo}
	}
	return nil
}

func (p *Service) Operation(model xai.Model, action xai.Action) (xai.Operation, error) {
	if action != xai.GenVideo || !IsVideoModel(string(model)) {
		return nil, xai.ErrNotFound
	}
	return &genVideo{model: normalizeModel(string(model))}, nil
}

type genVideo struct {
	model  string
	params *Params
}

func (p *genVideo) InputSchema() xai.InputSchema {
	return &inputSchema{model: p.model, fields: SchemaForVideo(p.model)}
}

func (p *genVideo) Params() xai.Params {
	if p.params == nil {
		p.params = NewParams()
	}
	return p.params
}

func (p *genVideo) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (xai.OperationResponse, error) {
	s := svc.(*Service)
	if s.backend == nil {
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
	if _, err := BuildVideoParams(p.model, params); err != nil {
		return nil, err
	}

	if o, ok := opts.(*Options); ok && strings.TrimSpace(o.UserID) != "" {
		params.Set("_user_id", strings.TrimSpace(o.UserID))
	}
	return s.backend.Submit(ctx, xai.Model(p.model), params)
}

// inputSchema implements xai.InputSchema.
type inputSchema struct {
	model  string
	fields []xai.Field
}

func (s *inputSchema) Fields() []xai.Field { return s.fields }

func (s *inputSchema) Restrict(name string) *xai.Restriction {
	return Restrict(s.model, name)
}

// validateParamsAgainstRestriction checks each set string param against Restriction.
func validateParamsAgainstRestriction(schema xai.InputSchema, params *Params) error {
	for name, val := range params.Export() {
		if strings.HasPrefix(name, "_") {
			continue
		}
		s, ok := val.(string)
		if !ok {
			continue
		}
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if r := schema.Restrict(name); r != nil {
			if err := r.ValidateString(name, s); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetTask returns an OperationResponse for an existing task by taskID.
func (p *Service) GetTask(ctx context.Context, model xai.Model, action xai.Action, taskID string) (xai.OperationResponse, error) {
	if action != xai.GenVideo || p.backend == nil {
		return nil, xai.ErrNotFound
	}
	return p.backend.GetTaskStatus(ctx, taskID)
}
