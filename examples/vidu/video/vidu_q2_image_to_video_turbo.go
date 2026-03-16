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

func runViduQ2ImageToVideoTurbo(ctx context.Context, svc xai.Service, model xai.Model) error {
	op, err := svc.Operation(model, xai.GenVideo)
	if err != nil {
		return err
	}
	op.Params().
		Set(vidu.ParamPrompt, "Fast-paced handheld motion as the camera circles a runner entering a futuristic subway station.").
		Set(vidu.ParamSeed, 24).
		Set(vidu.ParamDuration, 5).
		Set(vidu.ParamResolution, vidu.Resolution720p).
		Set(vidu.ParamMovementAmplitude, vidu.MovementMedium).
		Set(vidu.ParamWatermark, false).
		Set(vidu.ParamImageURL, DemoVideoURLs.ImageInput)

	results, err := xai.Call(ctx, svc, op, newViduOptions(svc, "demo-user-q2-image-turbo"), progressPrinter("q2-image-turbo"))
	if err != nil {
		return err
	}
	printVideoResults(string(model), "image-to-video-turbo", results)
	return nil
}
