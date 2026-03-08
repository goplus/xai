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

// Chat completions with thinking mode: enabled vs disabled.
// Uses DeepSeek model and WithThinking option (api.qnaigc.com).
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/openai"

	"github.com/goplus/xai/examples/openai/shared"
)

func runChatThinking() {
	svc := shared.NewService("")
	ctx := context.Background()

	params := svc.Params().
		Model(xai.Model(shared.ModelDeepSeekV32)).
		Messages(svc.UserMsg().Text("What is the Sun?"))

	// Thinking enabled
	fmt.Println("--- thinking enabled ---")
	opts := openai.WithThinking(svc.Options(), true)
	resp, err := shared.GenOrStream(ctx, svc, params, opts)
	if err != nil {
		fmt.Println("Error (thinking enabled):", err)
		return
	}
	if resp != nil {
		shared.PrintResponseBlocks(resp)
	}

	// Thinking disabled
	fmt.Println("\n--- thinking disabled ---")
	opts = openai.WithThinking(svc.Options(), false)
	resp, err = shared.GenOrStream(ctx, svc, params, opts)
	if err != nil {
		fmt.Println("Error (thinking disabled):", err)
		return
	}
	if resp != nil {
		shared.PrintResponseBlocks(resp)
	}
}
