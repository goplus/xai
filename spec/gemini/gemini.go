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

package gemini

import (
	"context"
	"errors"
	"iter"
	"net/url"
	"reflect"
	"strings"

	xai "github.com/goplus/xai/spec"
	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

type Service struct {
	backend Backend
	tools   tools
}

// NewWithBackend creates a gemini Service with a custom backend implementation.
func NewWithBackend(backend Backend) *Service {
	if backend == nil {
		panic("gemini: nil backend")
	}
	return &Service{
		backend: backend,
		tools:   make(tools),
	}
}

func (p *Service) Features() xai.Feature {
	return xai.FeatureGen | xai.FeatureGenStream | xai.FeatureOperation
}

// Backend returns the underlying Backend implementation.
func (p *Service) Backend() Backend {
	return p.backend
}

func (p *Service) Gen(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) (xai.GenResponse, error) {
	model, contents, config := buildGenParams(params)
	buildOptions(config, opts)
	if p.backend == nil {
		return nil, errors.New("gemini: backend not configured")
	}
	resp, err := p.backend.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return nil, err // TODO(xsw): translate error
	}
	return response{resp}, nil
}

func (p *Service) GenStream(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) iter.Seq2[xai.GenResponse, error] {
	model, contents, config := buildGenParams(params)
	buildOptions(config, opts)
	if p.backend == nil {
		return func(yield func(xai.GenResponse, error) bool) {
			yield(nil, errors.New("gemini: backend not configured"))
		}
	}
	iter := p.backend.GenerateContentStream(ctx, model, contents, config)
	return func(yield func(xai.GenResponse, error) bool) {
		iter(func(resp *genai.GenerateContentResponse, err error) bool {
			return yield(response{resp}, err)
		})
	}
}

// -----------------------------------------------------------------------------

const (
	Scheme = "gemini"
)

// New creates a new Service instance based on the scheme in the given URI.
// uri should be in the format of "gemini:base=service_base_url&key=api_key".
//
// `base` is the base URL of the API endpoint.
// `key` is the API key for authentication for Gemini backend.
// `project` is the project ID for Vertex AI backend.
// `location` is the location for Vertex AI backend.
//
// For example, "gemini:base=https://generativelanguage.googleapis.com/&key=your_api_key".
func New(ctx context.Context, uri string) (xai.Service, error) {
	params, err := url.ParseQuery(strings.TrimPrefix(uri, Scheme+":"))
	if err != nil {
		return nil, err
	}
	var conf genai.ClientConfig
	setNilEnvVarProvider(&conf)
	if base := params["base"]; len(base) > 0 {
		conf.HTTPOptions.BaseURL = base[0]
	}
	if key := params["key"]; len(key) > 0 {
		conf.APIKey = key[0]
	}
	if project := params["project"]; len(project) > 0 {
		conf.Project = project[0]
		conf.Backend = genai.BackendVertexAI
	}
	if location := params["location"]; len(location) > 0 {
		conf.Location = location[0]
	}
	cli, err := genai.NewClient(ctx, &conf)
	if err != nil {
		return nil, err
	}
	return NewWithBackend(newGenAIBackend(*cli.Models, *cli.Operations)), nil
}

// Remove calls to genai.defaultEnvVarProvider because we don't suggest users
// to set environment variables for API key and base URL. Instead, they should
// provide these parameters directly in the URI.
func setNilEnvVarProvider(conf *genai.ClientConfig) {
	v := reflect.ValueOf(conf).Elem().FieldByName("envVarProvider")
	if v.IsValid() {
		*(*func() map[string]string)(v.Addr().UnsafePointer()) = nilEnvVarProvider
	}
}

func nilEnvVarProvider() map[string]string {
	return nil
}

func init() {
	xai.Register(Scheme, New)
}

// -----------------------------------------------------------------------------
