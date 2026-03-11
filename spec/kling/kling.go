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
	"net/url"
	"strings"
	"time"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/geno"
	"golang.org/x/oauth2"
)

// -----------------------------------------------------------------------------

func newGenImageResponse(c *geno.Client, body map[string]any) (xai.OperationResponse, error) {
	return geno.NewOperationResponse[adapter](c, xai.GenImage, body), nil
}

// -----------------------------------------------------------------------------

type adapter struct{}

func (adapter) Actions(model xai.Model) []xai.Action {
	panic("todo")
}

func (adapter) ActionInfo(action xai.Action) geno.ActionInfo {
	switch action {
	case xai.GenImage:
		return geno.ActionInfo{
			Path:           "/v1/images/generations",
			ModelParamName: "model_name",
			InputSchema:    nil, // TODO(xsw)
			NewResponse:    newGenImageResponse,
		}
	default:
		panic("unexpected action: " + action)
	}
}

func (adapter) QueryOpInfo(action xai.Action, body map[string]any) (ret geno.QueryOpInfo, err error) {
	switch action {
	case xai.GenImage:
		data, _ := body["data"].(map[string]any)
		if id, ok := data["task_id"].(string); ok {
			ret.Path = "/v1/images/generations/" + id
			ret.NewResponse = newGenImageResponse
		} else {
			err = geno.ErrMissingOperationID
		}
	default:
		panic("unexpected action: " + action)
	}
	return
}

func (adapter) Results(action xai.Action, body map[string]any) xai.Results {
	result, _ := body["task_result"].(map[string]any)
	switch action {
	case xai.GenImage:
		return geno.NewImageResults[adapter](result, "images")
	default:
		panic("unexpected action: " + action)
	}
}

func (adapter) Sleep(action xai.Action, body map[string]any) {
	switch action {
	case xai.GenVideo:
		// sleep 10s for video operations
		time.Sleep(10 * time.Second)
	default:
		// sleep 0.5s for image operations
		time.Sleep(time.Second / 2)
	}
}

func (adapter) Done(action xai.Action, body map[string]any) bool {
	data, _ := body["data"].(map[string]any)
	switch data["task_status"] {
	case "succeed", "failed":
		return true
	default: // submitted, processing
		return false
	}
}

func (adapter) OutputImage(item any) *xai.OutputImage {
	panic("todo")
}

// -----------------------------------------------------------------------------

const (
	Scheme = "kling"
)

// New creates a new Service instance based on the scheme in the given URI.
// uri should be in the format of "kling:base=service_base_url&token=your_token".
//
// `base` is the base URL of the API endpoint.
// `timeout` is the request timeout duration (e.g., "30s").
// `token` is the authentication token for accessing the service.
//
// For example, "kling:base=https://api-singapore.klingai.com/&token=your_token".
func New(ctx context.Context, uri string) (xai.Service, error) {
	params, err := url.ParseQuery(strings.TrimPrefix(uri, Scheme+":"))
	if err != nil {
		return nil, err
	}

	var src oauth2.TokenSource
	if token := params["token"]; len(token) > 0 {
		src = oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: token[0],
		})
	} else {
		panic("token is required")
	}

	svc := geno.NewService[adapter](oauth2.NewClient(ctx, src))
	c := svc.HTTPClient()
	if base := params["base"]; len(base) > 0 {
		c.BaseURL(base[0])
	}
	if timeout := params["timeout"]; len(timeout) > 0 {
		d, err := time.ParseDuration(timeout[0])
		if err != nil {
			return nil, err
		}
		c.Timeout(d)
	}
	return svc, nil
}

func init() {
	xai.Register(Scheme, New)
}

// -----------------------------------------------------------------------------
