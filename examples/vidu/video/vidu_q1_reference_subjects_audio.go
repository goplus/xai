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

func runViduQ1ReferenceToVideoSubjectsAudio(ctx context.Context, svc xai.Service, model xai.Model) error {
	op, err := svc.Operation(model, xai.GenVideo)
	if err != nil {
		return err
	}
	op.Params().
		Set(vidu.ParamPrompt, "The @narrator walks past the @lantern on the @street and softly introduces the scene.").
		Set(vidu.ParamSeed, 5).
		Set(vidu.ParamDuration, 5).
		Set(vidu.ParamResolution, vidu.Resolution1080p).
		Set(vidu.ParamMovementAmplitude, vidu.MovementAuto).
		Set(vidu.ParamAspectRatio, vidu.AspectRatio9_16).
		Set(vidu.ParamAudio, true).
		Set(vidu.ParamWatermark, false).
		Set(vidu.ParamSubjects, []vidu.Subject{
			{ID: "narrator", Images: []string{DemoVideoURLs.Reference1}, VoiceID: "voice-narrator"},
			{ID: "lantern", Images: []string{DemoVideoURLs.Reference3}},
			{ID: "street", Images: []string{DemoVideoURLs.Reference2}},
		})

	results, err := xai.Call(ctx, svc, op, newViduOptions(svc, "demo-user-q1-ref-subjects-audio"), progressPrinter("q1-ref-subjects-audio"))
	if err != nil {
		return err
	}
	printVideoResults(string(model), "reference-to-video-subjects-audio", results)
	return nil
}
