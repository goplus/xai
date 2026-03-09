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

package video

import (
	"strings"

	"github.com/goplus/xai/spec/kling/internal"
)

var videoModels = []string{
	internal.ModelKlingV21Video, internal.ModelKlingV25Turbo, internal.ModelKlingVideoO1,
	internal.ModelKlingV26, internal.ModelKlingV27, internal.ModelKlingV28, internal.ModelKlingV29,
	internal.ModelKlingV3, internal.ModelKlingV3Omni,
}

// IsVideoModel returns true if the model supports video generation.
func IsVideoModel(m string) bool {
	m = strings.ToLower(m)
	return strings.HasPrefix(m, "kling") && (m == internal.ModelKlingVideoO1 ||
		strings.Contains(m, "v2-1") || strings.Contains(m, "v2-5") ||
		strings.Contains(m, "v2-6") || strings.Contains(m, "v2-7") ||
		strings.Contains(m, "v2-8") || strings.Contains(m, "v2-9") ||
		m == internal.ModelKlingV3 || m == internal.ModelKlingV3Omni)
}

// VideoModels returns all supported video model IDs.
func VideoModels() []string {
	out := make([]string, len(videoModels))
	copy(out, videoModels)
	return out
}
