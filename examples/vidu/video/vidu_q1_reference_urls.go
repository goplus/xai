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

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/vidu"
)

func runViduQ1ReferenceToVideoURLs(ctx context.Context, svc *vidu.Service, model xai.Model) error {
	op, err := svc.Operation(model, xai.GenVideo)
	if err != nil {
		return err
	}
	op.Params().
		Set(vidu.ParamPrompt, "The little devil is looking at the apple on the beach and walking around it.").
		Set(vidu.ParamSeed, 2).
		Set(vidu.ParamDuration, 5).
		Set(vidu.ParamResolution, vidu.Resolution1080p).
		Set(vidu.ParamMovementAmplitude, "auto").
		Set(vidu.ParamWatermark, true).
		Set(vidu.ParamReferenceImageURLs, []string{
			DemoVideoURLs.Reference1,
			DemoVideoURLs.Reference2,
			DemoVideoURLs.Reference3,
		})

	results, err := xai.Call(ctx, svc, op, newViduOptions(svc, "demo-user-q1-ref-urls"), progressPrinter("q1-ref-urls"))
	if err != nil {
		return err
	}
	printVideoResults(string(model), "reference-to-video-urls", results)
	return nil
}
