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
 * WITHOUT WARRANTIES OR CONDITIONS OF KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package image

import (
	"strings"

	"github.com/goplus/xai/spec/kling/internal"
)

var imageModels = []string{
	internal.ModelKlingV1, internal.ModelKlingV15, internal.ModelKlingV2,
	internal.ModelKlingV2New, internal.ModelKlingV21, internal.ModelKlingImageO1,
}

// IsImageModel returns true if the model supports image generation.
func IsImageModel(m string) bool {
	m = strings.ToLower(m)
	if !strings.HasPrefix(m, "kling") {
		return false
	}
	// Exclude video-only models
	if m == internal.ModelKlingVideoO1 || strings.Contains(m, "v2-5") || strings.Contains(m, "v2-6") ||
		strings.Contains(m, "v2-7") || strings.Contains(m, "v2-8") || strings.Contains(m, "v2-9") {
		return false
	}
	return true
}

// ImageModels returns all supported image model IDs.
func ImageModels() []string {
	out := make([]string, len(imageModels))
	copy(out, imageModels)
	return out
}
