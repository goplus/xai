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
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"

	oshared "github.com/goplus/xai/examples/openai/shared"
)

func runChatText() {
	service := oshared.NewService("")
	ctx := context.Background()

	params := service.Params().
		Model(xai.Model("gemini-2.5-flash-image")).
		Messages(service.UserMsg().Text("画一只可爱的橘猫，坐在窗台上看着夕阳"))

	resp, err := oshared.GenOrStream(ctx, service, params, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if resp == nil {
		return
	}
	oshared.PrintResponseBlocksWithTitle("response", resp)
}
