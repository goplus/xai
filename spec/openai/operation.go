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
	"context"

	xai "github.com/goplus/xai/spec"
)

// -----------------------------------------------------------------------------

func (p *Service) Actions(model xai.Model) []xai.Action {
	if isSoraModel(model) {
		return []xai.Action{xai.GenVideo}
	}
	return nil
}

func (p *Service) Operation(model xai.Model, action xai.Action) (op xai.Operation, err error) {
	if action != xai.GenVideo || !isSoraModel(model) {
		return nil, xai.ErrNotFound
	}
	return &genVideo{model: string(model)}, nil
}

// GetTask returns the current status for an existing Sora video task.
func (p *Service) GetTask(ctx context.Context, model xai.Model, action xai.Action, taskID string) (xai.OperationResponse, error) {
	if action != xai.GenVideo || !isSoraModel(model) {
		return nil, xai.ErrNotFound
	}
	task, err := p.getVideoTask(ctx, p.baseURL, taskID)
	if err != nil {
		return nil, err
	}
	return newVideoResp(task, p.baseURL), nil
}

// -----------------------------------------------------------------------------
