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
	"iter"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/goplus/xai"
	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

type Service struct {
	models genai.Models
	ops    genai.Operations
	tools  tools
}

func (p *Service) Features() xai.Feature {
	return xai.FeatureGen | xai.FeatureGenStream | xai.FeatureOperation
}

func (p *Service) Gen(params xai.GenParams) (xai.GenResponse, error) {
	ctx, model, contents, config := buildGenParams(params)
	resp, err := p.models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return nil, err // TODO(xsw): translate error
	}
	return response{resp}, nil
}

func (p *Service) GenStream(params xai.GenParams) iter.Seq2[xai.GenResponse, error] {
	ctx, model, contents, config := buildGenParams(params)
	iter := p.models.GenerateContentStream(ctx, model, contents, config)
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
// `timeout` is the request timeout duration (e.g., "30s").
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
	if timeout := params["timeout"]; len(timeout) > 0 {
		d, err := time.ParseDuration(timeout[0])
		if err != nil {
			return nil, err
		}
		conf.HTTPOptions.Timeout = &d
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
	return &Service{
		models: *cli.Models,
		ops:    *cli.Operations,
		tools:  make(tools),
	}, nil
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
