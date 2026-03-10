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

package geno

import (
	"context"
	"encoding/json"
	"net/http"

	xai "github.com/goplus/xai/spec"
)

// -----------------------------------------------------------------------------

type ActionInfo struct {
	Path           string   // path to call for the action, e.g. "/v1/images/generations"
	ModelParamName string   // name of the parameter for the model, e.g. "model_name"
	QueryPath      string   // path to query status for the action, e.g. "/v1/images/generations/{id}"
	QueryID        []string // e.g. ["data", "task_id"] for resp["data"]["task_id"]
}

type adapter interface {
	ActionInfo(action xai.Action) ActionInfo
}

// -----------------------------------------------------------------------------

type ServiceBase[T adapter] struct {
	c Client
}

// NewServiceBase creates a new ServiceBase with the provided HTTP client.
func NewServiceBase[T adapter](client *http.Client) *ServiceBase[T] {
	return &ServiceBase[T]{
		c: *NewClient(client),
	}
}

// HTTPClient returns the underlying HTTP client used by the service.
func (p *ServiceBase[T]) HTTPClient() *Client {
	return &p.c
}

// to implement xai.Service
func (p *ServiceBase[T]) Options() xai.OptionBuilder {
	return new(HTTPOptions)
}

// to implement xai.Service
func (p *ServiceBase[T]) Operation(model xai.Model, action xai.Action) (xai.Operation, error) {
	var adapter T
	ai := adapter.ActionInfo(action)
	req, err := p.c.NewRequest(http.MethodPost, ai.Path)
	if err != nil {
		return nil, err
	}
	body := make(map[string]any, 16)
	body[ai.ModelParamName] = model
	return &Operation{
		req:  req,
		body: body,
	}, nil
}

// -----------------------------------------------------------------------------

type Operation struct {
	req  *Request
	body map[string]any
}

func (p *Operation) InputSchema() xai.InputSchema {
	panic("todo")
}

func (p *Operation) Params() xai.Params {
	return p
}

func (p *Operation) Set(name string, val any) xai.Params {
	p.body[name] = val
	return p
}

func (p *Operation) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	ret, err := p.req.Do(ctx, opts.(*HTTPOptions))
	if err != nil {
		return nil, err
	}
	defer ret.Body.Close()

	var body map[string]any
	dec := json.NewDecoder(ret.Body)
	err = dec.Decode(&body)
	if err != nil {
		return nil, err
	}

	_ = body
	panic("todo")
}

// -----------------------------------------------------------------------------
