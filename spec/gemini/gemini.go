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
	"os"
	"reflect"
	"strings"

	xai "github.com/goplus/xai/spec"
	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

type Service struct {
	models genai.Models
	tools  tools
}

func (p *Service) Gen(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) (xai.GenResponse, error) {
	model, contents, config := buildParams(params)
	buildOptions(config, opts)
	resp, err := p.models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return nil, err // TODO(xsw): translate error
	}
	return response{resp}, nil
}

func (p *Service) GenStream(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) iter.Seq2[xai.GenResponse, error] {
	model, contents, config := buildParams(params)
	buildOptions(config, opts)
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
// uri should be in the format of "gemini:base=xxx&project=xxx", where "base" is
// the base URL of the API endpoint. "project" is the project ID, which is required
// when using the Vertex AI.
//
// For example, "gemini:base=https://generativelanguage.googleapis.com".
func New(ctx context.Context, uri string) (xai.Service, error) {
	params, err := url.ParseQuery(strings.TrimPrefix(uri, Scheme+":"))
	if err != nil {
		return nil, err
	}
	var conf genai.ClientConfig
	setEnvVarProvider(&conf)
	if base := params["base"]; len(base) > 0 {
		conf.HTTPOptions.BaseURL = base[0]
	}
	if project := params["project"]; len(project) > 0 {
		conf.Project = project[0]
		conf.Backend = genai.BackendVertexAI
	}
	if key := params["key"]; len(key) > 0 {
		conf.APIKey = key[0]
	}
	cli, err := genai.NewClient(ctx, &conf)
	if err != nil {
		return nil, err
	}
	return &Service{
		models: *cli.Models,
		tools:  make(tools),
	}, nil
}

// Remove calls to genai.defaultEnvVarProvider because we don't suggest users
// to set environment variables for API key and base URL. Instead, they should
// provide these parameters directly in the URI.
func setEnvVarProvider(conf *genai.ClientConfig) {
	v := reflect.ValueOf(conf).Elem().FieldByName("envVarProvider")
	if v.IsValid() {
		*(*func() map[string]string)(v.Addr().UnsafePointer()) = envVarProvider
	}
}

// envVarProvider only returns GOOGLE_CLOUD_LOCATION and GOOGLE_CLOUD_REGION.
func envVarProvider() map[string]string {
	vars := make(map[string]string)
	if v, ok := os.LookupEnv("GOOGLE_CLOUD_LOCATION"); ok {
		vars["GOOGLE_CLOUD_LOCATION"] = v
	}
	if v, ok := os.LookupEnv("GOOGLE_CLOUD_REGION"); ok {
		vars["GOOGLE_CLOUD_REGION"] = v
	}
	return vars
}

func init() {
	xai.Register(Scheme, New)
}

// -----------------------------------------------------------------------------
