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

package shared

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/openai"
	"github.com/goplus/xai/spec/openai/provider/qiniu"
)

// ModelGeminiPro is the default model for qiniu chat completions.
const ModelGeminiPro = "gemini-3.0-pro-preview"

// ModelDeepSeekV32 is a thinking-enabled model (supports thinking param).
const ModelDeepSeekV32 = "deepseek/deepseek-v3.2-251201"

var streamMode atomic.Bool

func init() {
	streamMode.Store(parseStreamMode(os.Getenv("STREAM")))
}

// NewService creates an OpenAI-compatible Service backed by Qiniu API.
// It is wired to provider_v1 (Chat Completions API) through qiniu.NewService.
// Uses QINIU_API_KEY from environment when token is empty.
func NewService(token string) *openai.Service {
	if token == "" {
		token = os.Getenv("QINIU_API_KEY")
	}
	return qiniu.NewService(token)
}

// SetStreamMode sets whether GenOrStream uses streaming mode.
func SetStreamMode(enabled bool) {
	streamMode.Store(enabled)
}

// StreamMode returns current streaming mode.
func StreamMode() bool {
	return streamMode.Load()
}

func parseStreamMode(v string) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	if v == "1" || v == "true" || v == "yes" || v == "on" {
		return true
	}
	return false
}

// DebugOptions ensures OpenAI examples always print equivalent curl commands.
func DebugOptions(svc *openai.Service, opts xai.OptionBuilder) xai.OptionBuilder {
	if opts == nil {
		opts = svc.Options()
	}
	return openai.WithDebugCurl(opts, true)
}

// GenOrStream runs svc.Gen (non-stream) or svc.GenStream based on StreamMode().
// For stream mode, prints each delta to stdout. For non-stream, returns full response.
func GenOrStream(ctx context.Context, svc *openai.Service, params xai.ParamBuilder, opts xai.OptionBuilder) (xai.GenResponse, error) {
	opts = DebugOptions(svc, opts)
	if StreamMode() {
		return nil, runStream(ctx, svc, params, opts)
	}
	return svc.Gen(ctx, params, opts)
}

func runStream(ctx context.Context, svc *openai.Service, params xai.ParamBuilder, opts xai.OptionBuilder) error {
	iter := svc.GenStream(ctx, params, opts)
	chunk := 0
	for resp, err := range iter {
		if err != nil {
			return err
		}
		if resp != nil && resp.Len() > 0 {
			PrintResponseBlocksWithTitle(fmt.Sprintf("stream_chunk[%d]", chunk), resp)
			chunk++
		}
	}
	return nil
}
