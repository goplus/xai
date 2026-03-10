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

// ResponseCreator is a function type that creates an xai.OperationResponse from
// a given HTTP response body.
type ResponseCreator func(c *Client, body map[string]any) (xai.OperationResponse, error)

// ActionInfo contains information about an action, including the path to call for
// the action, the name of the parameter for the model, and a function to create a
// response from the HTTP response body.
type ActionInfo struct {
	Path           string // path to call for the action, e.g. "/v1/images/generations"
	ModelParamName string // name of the parameter for the model, e.g. "model_name"
	NewResponse    ResponseCreator
}

type serviceAdapter interface {
	// ActionInfo returns the ActionInfo for the given action. The implementation of
	// this method should determine the path, model parameter name, and response creator
	// function based on the action.
	ActionInfo(action xai.Action) ActionInfo
}

type Service[T serviceAdapter] struct {
	c Client
}

// NewService creates a new Service with the provided HTTP client.
func NewService[T serviceAdapter](client *http.Client) *Service[T] {
	return &Service[T]{
		c: *NewClient(client),
	}
}

// HTTPClient returns the underlying HTTP client used by the service.
func (p *Service[T]) HTTPClient() *Client {
	return &p.c
}

// implement xai.Service
func (p *Service[T]) Options() xai.OptionBuilder {
	return new(HTTPOptions)
}

// implement xai.Service
func (p *Service[T]) Operation(model xai.Model, action xai.Action) (xai.Operation, error) {
	var adapter T
	ai := adapter.ActionInfo(action)
	req, err := p.c.NewRequest(http.MethodPost, ai.Path)
	if err != nil {
		return nil, err
	}
	body := make(map[string]any, 16)
	body[ai.ModelParamName] = model
	return &Operation{
		req:         req,
		body:        body,
		newResponse: ai.NewResponse,
	}, nil
}

// -----------------------------------------------------------------------------

type Operation struct {
	req         *Request
	body        map[string]any
	newResponse ResponseCreator
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
	req := p.req
	err = req.Json(p.body)
	if err != nil {
		return
	}

	ret, err := req.Do(ctx, opts.(*HTTPOptions))
	if err != nil {
		return
	}
	defer ret.Body.Close()

	var body map[string]any
	dec := json.NewDecoder(ret.Body)
	err = dec.Decode(&body)
	if err != nil {
		return
	}

	return p.newResponse(p.req.c, body)
}

// -----------------------------------------------------------------------------

type OperationResponse struct {
}

func (p *OperationResponse) Done() bool {
	panic("todo")
}

func (p *OperationResponse) Sleep() {
	panic("todo")
}

func (p *OperationResponse) Retry(ctx context.Context, svc xai.Service) (xai.OperationResponse, error) {
	panic("todo")
}

func (p *OperationResponse) Results() xai.Results {
	panic("todo")
}

// -----------------------------------------------------------------------------
