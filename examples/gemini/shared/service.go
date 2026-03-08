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
	"github.com/goplus/xai/spec/gemini/provider/qiniu"
)

// Gemini image models on Qiniu.
const (
	ModelFlashImage      = "gemini-2.5-flash-image"
	ModelProImagePreview = "gemini-3.0-pro-image-preview"
	ModelV31ImagePreview = "gemini-3.1-flash-image-preview"
)

var streamMode atomic.Bool

func init() {
	streamMode.Store(parseStreamMode(os.Getenv("STREAM")))
}

// NewService creates a Gemini-capable xai.Service backed by Qiniu provider.
//
// The returned value is intentionally typed as xai.Service so callers depend on
// the portable service contract instead of a concrete provider type.
func NewService(token string) xai.Service {
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

// GenOrStream runs service.Gen (non-stream) or service.GenStream based on
// StreamMode().
// For stream mode, prints each delta to stdout. For non-stream, returns full response.
func GenOrStream(ctx context.Context, service xai.Service, params xai.ParamBuilder, opts xai.OptionBuilder) (xai.GenResponse, error) {
	if StreamMode() {
		return nil, runStream(ctx, service, params, opts)
	}
	return service.Gen(ctx, params, opts)
}

func runStream(ctx context.Context, service xai.Service, params xai.ParamBuilder, opts xai.OptionBuilder) error {
	iter := service.GenStream(ctx, params, opts)
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
