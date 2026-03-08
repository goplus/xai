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
	"fmt"

	gshared "github.com/goplus/xai/examples/gemini/shared"
	xai "github.com/goplus/xai/spec"
)

func printImageResults(results xai.Results) {
	fmt.Printf("results { images: %d }\n", results.Len())
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputImage)
		fmt.Printf("  image[%d]: %s\n", i, gshared.TruncateForPrint(out.Image.StgUri(), 120))
	}
}
