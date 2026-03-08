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

package image

import (
	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling/internal"
	"github.com/goplus/xai/types"
)

// SchemaV1Image returns the InputSchema fields for kling-v1.
func SchemaV1Image() []xai.Field {
	return []xai.Field{
		{Name: internal.ParamPrompt, Kind: types.String},
		{Name: internal.ParamAspectRatio, Kind: types.String},
		{Name: internal.ParamImage, Kind: types.String},
		{Name: internal.ParamNegativePrompt, Kind: types.String},
		{Name: internal.ParamImageFidelity, Kind: types.Float},
		{Name: internal.ParamHumanFidelity, Kind: types.Float},
		{Name: internal.ParamN, Kind: types.Int},
	}
}
