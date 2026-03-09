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

func runTTS() {
	svc := newService()
	ctx := context.Background()

	op, err := svc.Operation(xai.Model(audio.ModelTTSV1), xai.Synthesize)
	if err != nil {
		fmt.Println("Operation error:", err)
		return
	}
	op.Params().Set(audio.ParamInput, "Hello, world! This is a text-to-speech demo.")
	op.Params().Set(audio.ParamFormat, "mp3")

	resp, err := xai.CallSync(ctx, svc, op, svc.Options())
	if err != nil {
		fmt.Println("Call error:", err)
		return
	}

	if !resp.Done() {
		fmt.Println("Unexpected: TTS is sync, should be done immediately")
		return
	}

	results := resp.Results()
	if results.Len() == 0 {
		fmt.Println("No results")
		return
	}

	out := results.At(0).(*xai.OutputAudio)
	fmt.Println("Audio URL:", out.Audio)
	fmt.Println("Format:", out.Format)
	fmt.Println("Duration:", out.Duration)
}

func runTTSWithVoice() {
	svc := newService()
	ctx := context.Background()

	op, err := svc.Operation(xai.Model(audio.ModelTTSV1), xai.Synthesize)
	if err != nil {
		fmt.Println("Operation error:", err)
		return
	}
	op.Params().Set(audio.ParamInput, "你好，世界！这是语音合成演示。")
	op.Params().Set(audio.ParamVoice, "qiniu_zh_female_wwxkjx")
	op.Params().Set(audio.ParamFormat, "mp3")
	op.Params().Set(audio.ParamSpeed, 1.0)

	resp, err := xai.CallSync(ctx, svc, op, svc.Options())
	if err != nil {
		fmt.Println("Call error:", err)
		return
	}

	if !resp.Done() {
		fmt.Println("Unexpected: TTS is sync, should be done immediately")
		return
	}

	results := resp.Results()
	if results.Len() == 0 {
		fmt.Println("No results")
		return
	}

	out := results.At(0).(*xai.OutputAudio)
	fmt.Println("Audio URL:", out.Audio)
	fmt.Println("Format:", out.Format)
	fmt.Println("Duration:", out.Duration)
}
