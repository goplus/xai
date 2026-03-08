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
	"bytes"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/openai/openai-go/v3/option"
)

// -----------------------------------------------------------------------------

type options struct {
	opts []option.RequestOption
}

func (p *options) WithBaseURL(base string) xai.OptionBuilder {
	p.opts = append(p.opts, option.WithBaseURL(base))
	return p
}

func (p *Service) Options() xai.OptionBuilder {
	return &options{}
}

// WithThinking returns an OptionBuilder with thinking enabled or disabled.
// Pass svc.Options() as the first argument. Only effective for OpenAI-compatible services.
// Use thinking-enabled models like deepseek-v3.2-251201 for best results.
func WithThinking(ob xai.OptionBuilder, enabled bool) xai.OptionBuilder {
	if p, ok := ob.(*options); ok {
		return p.withThinking(enabled)
	}
	return ob
}

// WithDebugCurl logs an equivalent curl command for each outgoing request.
// Pass svc.Options() as the first argument.
func WithDebugCurl(ob xai.OptionBuilder, enabled bool) xai.OptionBuilder {
	if p, ok := ob.(*options); ok {
		return p.withDebugCurl(enabled)
	}
	return ob
}

func (p *options) withThinking(enabled bool) *options {
	typ := "disabled"
	if enabled {
		typ = "enabled"
	}
	p.opts = append(p.opts, option.WithJSONSet("thinking", map[string]string{"type": typ}))
	return p
}

func (p *options) withDebugCurl(enabled bool) *options {
	if !enabled {
		return p
	}
	p.opts = append(p.opts, option.WithMiddleware(func(req *http.Request, nxt option.MiddlewareNext) (*http.Response, error) {
		body, err := snapshotBody(req)
		if err != nil {
			log.Printf("[openai] failed to read request body: %v", err)
		}
		log.Printf("[openai] curl command:\n%s", buildCurlCommand(req, body))
		return nxt(req)
	}))
	return p
}

func snapshotBody(req *http.Request) ([]byte, error) {
	if req == nil || req.Body == nil {
		return nil, nil
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	return body, nil
}

func buildCurlCommand(req *http.Request, body []byte) string {
	if req == nil || req.URL == nil {
		return "curl"
	}
	var sb strings.Builder
	sb.WriteString("curl -X ")
	sb.WriteString(req.Method)

	keys := make([]string, 0, len(req.Header))
	for key := range req.Header {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		values := req.Header.Values(key)
		for _, value := range values {
			sb.WriteString(" -H ")
			sb.WriteString(shellQuote(key + ": " + value))
		}
	}

	if len(body) > 0 {
		sb.WriteString(" --data-raw ")
		sb.WriteString(shellQuote(string(body)))
	}

	sb.WriteString(" ")
	sb.WriteString(shellQuote(req.URL.String()))
	return sb.String()
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func buildOptions(opts xai.OptionBuilder) (ret []option.RequestOption) {
	if p, ok := opts.(*options); ok {
		ret = p.opts
	}
	return
}

// -----------------------------------------------------------------------------
