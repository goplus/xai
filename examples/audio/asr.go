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
	"github.com/goplus/xai/spec/audio"
)

func runASR() {
	svc := newService()
	ctx := context.Background()

	op, err := svc.Operation(xai.Model(audio.ModelASR), xai.Transcribe)
	if err != nil {
		fmt.Println("Operation error:", err)
		return
	}
	op.Params().Set(audio.ParamAudio, DemoAudioURL)
	op.Params().Set(audio.ParamFormat, "mp3")

	resp, err := xai.CallSync(ctx, svc, op, svc.Options())
	if err != nil {
		fmt.Println("Call error:", err)
		return
	}

	if !resp.Done() {
		fmt.Println("Unexpected: ASR is sync, should be done immediately")
		return
	}

	results := resp.Results()
	if results.Len() == 0 {
		fmt.Println("No results")
		return
	}

	out := results.At(0).(*xai.OutputText)
	fmt.Println("Transcribed text:", out.Text)
	if out.Duration != nil {
		fmt.Printf("Duration: %.2f seconds\n", *out.Duration)
	}
}
