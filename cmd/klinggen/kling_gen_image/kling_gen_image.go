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

package main

import (
	"github.com/goplus/xai/cmd/klinggen"
)

// -----------------------------------------------------------------------------

var imageModelSels = map[string]klinggen.None{
	"imageGeneration":   {},
	"OmniImage":         {},
	"multiImageToImage": {},
}

var knownParams = map[string]string{
	"model_name":   "",
	"prompt":       "Prompt",
	"n":            "NumberOfImages",
	"aspect_ratio": "AspectRatio",
	"resolution":   "ImageSize",
}

var knownParamTypes = map[string]string{
	"image_list": "[]xai.Image",
}

func main() {
	util := &klinggen.Util{
		ModelSels:       imageModelSels,
		KnownParams:     knownParams,
		KnownParamTypes: knownParamTypes,
	}
	util.DoFile("../../../spec/kling/klingai.json")
}

// -----------------------------------------------------------------------------
