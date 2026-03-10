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

package vidu

import (
	"context"

	xai "github.com/goplus/xai/spec"
)

// Scheme is the URI scheme for Vidu: "vidu".
const Scheme = "vidu"

// Video model constants.
const (
	ModelViduQ1       = "vidu-q1"
	ModelViduQ2       = "vidu-q2"
	ModelViduQ2Turbo  = "viduq2-turbo"
	ModelViduQ2Pro    = "viduq2-pro"
)

var videoModels = []string{ModelViduQ1, ModelViduQ2, ModelViduQ2Turbo, ModelViduQ2Pro}

// IsVideoModel returns true if the model supports video generation.
func IsVideoModel(model string) bool {
	switch normalizeModel(model) {
	case ModelViduQ1, ModelViduQ2, ModelViduQ2Turbo, ModelViduQ2Pro:
		return true
	default:
		return false
	}
}

// VideoModels returns all supported video model IDs.
func VideoModels() []string {
	out := make([]string, len(videoModels))
	copy(out, videoModels)
	return out
}

// Register registers the Vidu service with xai.
// Accepts any xai.Service implementation (e.g. *vidu.Service or *qiniu.Service).
func Register(svc xai.Service) {
	xai.Register(Scheme, func(ctx context.Context, uri string) (xai.Service, error) {
		return svc, nil
	})
}
