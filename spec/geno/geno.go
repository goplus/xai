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
	Path           string          // path to call for the action, e.g. "/v1/images/generations"
	ModelParamName string          // name of the parameter for the model, e.g. "model_name"
	InputSchema    xai.InputSchema // input schema for the action
	NewResponse    ResponseCreator
}

type serviceAdapter interface {
	// ActionInfo returns the ActionInfo for the given action. The implementation of
	// this method should determine the path, model parameter name, and response creator
	// function based on the action.
	ActionInfo(action xai.Action) ActionInfo
	Actions(model xai.Model) []xai.Action
}

type Service[T serviceAdapter] struct {
	c       Client
	adapter T
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
func (p *Service[T]) Actions(model xai.Model) []xai.Action {
	return p.adapter.Actions(model)
}

// implement xai.Service
func (p *Service[T]) Operation(model xai.Model, action xai.Action) (xai.Operation, error) {
	ai := p.adapter.ActionInfo(action)
	req, err := p.c.NewRequest(http.MethodPost, ai.Path)
	if err != nil {
		return nil, err
	}
	body := make(map[string]any, 16)
	body[ai.ModelParamName] = model
	return &Operation{
		req:         req,
		body:        body,
		inputSchema: ai.InputSchema,
		newResponse: ai.NewResponse,
	}, nil
}

// -----------------------------------------------------------------------------

type Operation struct {
	req         *Request
	body        map[string]any
	inputSchema xai.InputSchema
	newResponse ResponseCreator
}

func (p *Operation) InputSchema() xai.InputSchema {
	return p.inputSchema
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
	return call(ctx, req, p.newResponse, opts)
}

func call(ctx context.Context, req *Request, newResponse ResponseCreator, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	ret, err := req.Do(ctx, opts)
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
	return newResponse(req.c, body)
}

// -----------------------------------------------------------------------------

type results struct {
	result map[string]any
	items  []any
}

func newResults(result map[string]any, itemsName string) results {
	items, _ := result[itemsName].([]any)
	return results{
		result: result,
		items:  items,
	}
}

func (p *results) XGo_Attr(name string) any {
	return p.result[name]
}

func (p *results) Len() int {
	return len(p.items)
}

// -----------------------------------------------------------------------------

type imageResultAdapter interface {
	OutputImage(item any) *xai.OutputImage
}

type ImageResults[T imageResultAdapter] struct {
	results
}

func NewImageResults[T imageResultAdapter](result map[string]any, itemsName string) xai.Results {
	return &ImageResults[T]{
		results: newResults(result, itemsName),
	}
}

func (p *ImageResults[T]) At(i int) xai.Generated {
	var adapter T
	return adapter.OutputImage(p.items[i])
}

// -----------------------------------------------------------------------------

type imageMaskResultAdapter interface {
	OutputImageMask(item any) *xai.OutputImageMask
}

type ImageMaskResults[T imageMaskResultAdapter] struct {
	results
}

func NewImageMaskResults[T imageMaskResultAdapter](result map[string]any, itemsName string) xai.Results {
	return &ImageMaskResults[T]{
		results: newResults(result, itemsName),
	}
}

func (p *ImageMaskResults[T]) At(i int) xai.Generated {
	var adapter T
	return adapter.OutputImageMask(p.items[i])
}

// -----------------------------------------------------------------------------

type videoResultAdapter interface {
	OutputVideo(item any) *xai.OutputVideo
}

type VideoResults[T videoResultAdapter] struct {
	results
}

func NewVideoResults[T videoResultAdapter](result map[string]any, itemsName string) xai.Results {
	return &VideoResults[T]{
		results: newResults(result, itemsName),
	}
}

func (p *VideoResults[T]) At(i int) xai.Generated {
	var adapter T
	return adapter.OutputVideo(p.items[i])
}

// -----------------------------------------------------------------------------

type QueryOpInfo struct {
	Path        string
	QueryBody   map[string]any
	NewResponse ResponseCreator
}

type responseAdapter interface {
	Done(body map[string]any) bool
	Sleep(body map[string]any)
	Results(body map[string]any) xai.Results
	QueryOpInfo(body map[string]any) QueryOpInfo
}

type OperationResponse[T responseAdapter] struct {
	body    map[string]any
	c       *Client
	adapter T
}

func NewOperationResponse[T responseAdapter](c *Client, body map[string]any) *OperationResponse[T] {
	return &OperationResponse[T]{c: c, body: body}
}

func (p *OperationResponse[T]) Done() bool {
	return p.adapter.Done(p.body)
}

func (p *OperationResponse[T]) Sleep() {
	p.adapter.Sleep(p.body)
}

func (p *OperationResponse[T]) Retry(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	qoi := p.adapter.QueryOpInfo(p.body)
	req, err := p.c.NewRequest(http.MethodGet, qoi.Path)
	if err != nil {
		return
	}
	err = req.Json(qoi.QueryBody)
	if err != nil {
		return
	}
	return call(ctx, req, qoi.NewResponse, opts)
}

func (p *OperationResponse[T]) Results() xai.Results {
	return p.adapter.Results(p.body)
}

// -----------------------------------------------------------------------------
