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
	"net/http"

	xai "github.com/goplus/xai/spec"
)

// -----------------------------------------------------------------------------

type ServiceBase struct {
	c Client
}

func NewServiceBase(client *http.Client) *ServiceBase {
	return &ServiceBase{
		c: *NewClient(client),
	}
}

func (p *ServiceBase) SetOptions(opts *HTTPOptions) {
	p.c.baseURL = opts.baseURL
	if opts.timeout != nil {
		p.c.Timeout(*opts.timeout)
	}
}

func (p *ServiceBase) Options() xai.OptionBuilder {
	return new(HTTPOptions)
}

func (p *ServiceBase) NewOperation(method, path string) (ret Operation, err error) {
	ret.req, err = p.c.NewRequest(method, path)
	return
}

// -----------------------------------------------------------------------------

type Operation struct {
	req *Request
}

func (p Operation) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	ret, err := p.req.Do(ctx, opts.(*HTTPOptions))
	if err != nil {
		return nil, err
	}
	_ = ret
	panic("todo")
}

// -----------------------------------------------------------------------------
