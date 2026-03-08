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
	"testing"

	"github.com/goplus/xai/spec/kling/internal"
)

// allImageModels lists every image model for exhaustive testing.
var allImageModels = []string{
	internal.ModelKlingV1,
	internal.ModelKlingV15,
	internal.ModelKlingV2,
	internal.ModelKlingV2New,
	internal.ModelKlingV21,
	internal.ModelKlingImageO1,
}

// --- Restrict: aspect_ratio ---

func TestRestrict_AspectRatio_AllModels(t *testing.T) {
	validValues := []string{"1:1", "16:9", "9:16", "4:3", "3:4", "3:2", "2:3", "21:9"}
	invalidValues := []string{"1:2", "5:4", "widescreen", "square", "1920x1080"}

	for _, model := range allImageModels {
		r := Restrict(model, internal.ParamAspectRatio)
		if r == nil {
			t.Errorf("model %q: Restrict(aspect_ratio) returned nil", model)
			continue
		}
		for _, v := range validValues {
			if err := r.ValidateString(internal.ParamAspectRatio, v); err != nil {
				t.Errorf("model %q: aspect_ratio=%q should be valid, got: %v", model, v, err)
			}
		}
		for _, v := range invalidValues {
			if err := r.ValidateString(internal.ParamAspectRatio, v); err == nil {
				t.Errorf("model %q: aspect_ratio=%q should be rejected", model, v)
			}
		}
	}
}

func TestRestrict_AspectRatio_O1_SupportsAuto(t *testing.T) {
	r := Restrict(internal.ModelKlingImageO1, internal.ParamAspectRatio)
	if r == nil {
		t.Fatal("Restrict(aspect_ratio) for O1 returned nil")
	}
	if err := r.ValidateString(internal.ParamAspectRatio, "auto"); err != nil {
		t.Errorf("kling-image-o1: aspect_ratio='auto' should be valid, got: %v", err)
	}
}

// --- Restrict: image_reference (kling-v1-5 only) ---

func TestRestrict_ImageReference_V15(t *testing.T) {
	r := Restrict(internal.ModelKlingV15, internal.ParamImageReference)
	if r == nil {
		t.Fatal("Restrict(image_reference) for kling-v1-5 returned nil")
	}
	validValues := []string{"subject", "face"}
	invalidValues := []string{"scene", "style", "background", "person"}

	for _, v := range validValues {
		if err := r.ValidateString(internal.ParamImageReference, v); err != nil {
			t.Errorf("image_reference=%q should be valid, got: %v", v, err)
		}
	}
	for _, v := range invalidValues {
		if err := r.ValidateString(internal.ParamImageReference, v); err == nil {
			t.Errorf("image_reference=%q should be rejected", v)
		}
	}
}

func TestRestrict_ImageReference_NilForNonV15(t *testing.T) {
	nonV15 := []string{
		internal.ModelKlingV1, internal.ModelKlingV2, internal.ModelKlingV2New,
		internal.ModelKlingV21, internal.ModelKlingImageO1,
	}
	for _, model := range nonV15 {
		r := Restrict(model, internal.ParamImageReference)
		if r != nil {
			t.Errorf("model %q: Restrict(image_reference) should return nil", model)
		}
	}
}

// --- Restrict: resolution (kling-image-o1 only) ---

func TestRestrict_Resolution_O1(t *testing.T) {
	r := Restrict(internal.ModelKlingImageO1, internal.ParamResolution)
	if r == nil {
		t.Fatal("Restrict(resolution) for kling-image-o1 returned nil")
	}
	validValues := []string{"1K", "2K", "4K"}
	invalidValues := []string{"720p", "1080p", "8K", "HD", "auto"}

	for _, v := range validValues {
		if err := r.ValidateString(internal.ParamResolution, v); err != nil {
			t.Errorf("resolution=%q should be valid, got: %v", v, err)
		}
	}
	for _, v := range invalidValues {
		if err := r.ValidateString(internal.ParamResolution, v); err == nil {
			t.Errorf("resolution=%q should be rejected", v)
		}
	}
}

func TestRestrict_Resolution_NilForNonO1(t *testing.T) {
	nonO1 := []string{
		internal.ModelKlingV1, internal.ModelKlingV15, internal.ModelKlingV2,
		internal.ModelKlingV2New, internal.ModelKlingV21,
	}
	for _, model := range nonO1 {
		r := Restrict(model, internal.ParamResolution)
		if r != nil {
			t.Errorf("model %q: Restrict(resolution) should return nil", model)
		}
	}
}

// --- Restrict: unrestricted params return nil ---

func TestRestrict_UnrestrictedParams(t *testing.T) {
	unrestricted := []string{
		internal.ParamPrompt,
		internal.ParamImage,
		internal.ParamNegativePrompt,
		internal.ParamSubjectImageList,
		internal.ParamSceneImage,
		internal.ParamStyleImage,
		internal.ParamN,
		internal.ParamImageFidelity,
		internal.ParamHumanFidelity,
		"unknown_param",
	}
	for _, model := range allImageModels {
		for _, name := range unrestricted {
			r := Restrict(model, name)
			if r != nil {
				t.Errorf("model %q, param %q: expected nil Restriction, got %+v", model, name, r)
			}
		}
	}
}

// --- Schema completeness ---

func TestSchemaForImage_AllModelsReturnFields(t *testing.T) {
	for _, model := range allImageModels {
		fields := SchemaForImage(model)
		if len(fields) == 0 {
			t.Errorf("model %q: SchemaForImage returned no fields", model)
		}
		hasPrompt := false
		for _, f := range fields {
			if f.Name == internal.ParamPrompt {
				hasPrompt = true
			}
		}
		if !hasPrompt {
			t.Errorf("model %q: schema missing required field 'prompt'", model)
		}
	}
}

func TestSchemaForImage_UnknownModelReturnsFallback(t *testing.T) {
	fields := SchemaForImage("unknown-model")
	if len(fields) == 0 {
		t.Error("unknown model should return default schema, got empty")
	}
}
