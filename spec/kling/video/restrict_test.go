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

package video

import (
	"testing"

	"github.com/goplus/xai/spec/kling/internal"
)

// allVideoModels lists every video model for exhaustive testing.
var allVideoModels = []string{
	internal.ModelKlingV21Video,
	internal.ModelKlingV25Turbo,
	internal.ModelKlingVideoO1,
	internal.ModelKlingV26,
	internal.ModelKlingV27,
	internal.ModelKlingV28,
	internal.ModelKlingV29,
	internal.ModelKlingV3,
	internal.ModelKlingV3Omni,
}

// --- Restrict: mode ---

func TestRestrict_Mode_AllModels(t *testing.T) {
	validValues := []string{"std", "pro"}
	invalidValues := []string{"turbo", "fast", "standard", "PRO", "STD"}

	for _, model := range allVideoModels {
		r := Restrict(model, internal.ParamMode)
		if r == nil {
			t.Errorf("model %q: Restrict(mode) returned nil, expected non-nil", model)
			continue
		}
		for _, v := range validValues {
			if err := r.ValidateString(internal.ParamMode, v); err != nil {
				t.Errorf("model %q: mode=%q should be valid, got: %v", model, v, err)
			}
		}
		for _, v := range invalidValues {
			if err := r.ValidateString(internal.ParamMode, v); err == nil {
				t.Errorf("model %q: mode=%q should be rejected", model, v)
			}
		}
	}
}

// --- Restrict: seconds ---

func TestRestrict_Seconds_V21_V25_O1_V26(t *testing.T) {
	models := []string{
		internal.ModelKlingV21Video,
		internal.ModelKlingV25Turbo,
		internal.ModelKlingVideoO1,
		internal.ModelKlingV26,
		internal.ModelKlingV27,
		internal.ModelKlingV28,
		internal.ModelKlingV29,
	}
	validValues := []string{"5", "10"}
	invalidValues := []string{"3", "7", "15", "0", "20", "seconds"}

	for _, model := range models {
		r := Restrict(model, internal.ParamSeconds)
		if r == nil {
			t.Errorf("model %q: Restrict(seconds) returned nil", model)
			continue
		}
		for _, v := range validValues {
			if err := r.ValidateString(internal.ParamSeconds, v); err != nil {
				t.Errorf("model %q: seconds=%q should be valid, got: %v", model, v, err)
			}
		}
		for _, v := range invalidValues {
			if err := r.ValidateString(internal.ParamSeconds, v); err == nil {
				t.Errorf("model %q: seconds=%q should be rejected", model, v)
			}
		}
	}
}

func TestRestrict_Seconds_V3_V3Omni(t *testing.T) {
	models := []string{internal.ModelKlingV3, internal.ModelKlingV3Omni}
	validValues := []string{"3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"}
	invalidValues := []string{"0", "1", "2", "16", "20", "seconds"}

	for _, model := range models {
		r := Restrict(model, internal.ParamSeconds)
		if r == nil {
			t.Errorf("model %q: Restrict(seconds) returned nil", model)
			continue
		}
		for _, v := range validValues {
			if err := r.ValidateString(internal.ParamSeconds, v); err != nil {
				t.Errorf("model %q: seconds=%q should be valid, got: %v", model, v, err)
			}
		}
		for _, v := range invalidValues {
			if err := r.ValidateString(internal.ParamSeconds, v); err == nil {
				t.Errorf("model %q: seconds=%q should be rejected", model, v)
			}
		}
	}
}

// --- Restrict: size ---

func TestRestrict_Size_AllModels(t *testing.T) {
	validValues := []string{"1920x1080", "1080x1920", "1280x720", "720x1280", "1080x1080", "720x720"}
	invalidValues := []string{"1920x1920", "640x480", "4k", "1080p", "auto"}

	for _, model := range allVideoModels {
		r := Restrict(model, internal.ParamSize)
		if r == nil {
			t.Errorf("model %q: Restrict(size) returned nil", model)
			continue
		}
		for _, v := range validValues {
			if err := r.ValidateString(internal.ParamSize, v); err != nil {
				t.Errorf("model %q: size=%q should be valid, got: %v", model, v, err)
			}
		}
		for _, v := range invalidValues {
			if err := r.ValidateString(internal.ParamSize, v); err == nil {
				t.Errorf("model %q: size=%q should be rejected", model, v)
			}
		}
	}
}

// --- Restrict: sound ---

func TestRestrict_Sound_V26_V3(t *testing.T) {
	modelsWithSound := []string{
		internal.ModelKlingV26, internal.ModelKlingV27,
		internal.ModelKlingV28, internal.ModelKlingV29,
		internal.ModelKlingV3, internal.ModelKlingV3Omni,
	}
	validValues := []string{"on", "off"}
	invalidValues := []string{"yes", "no", "true", "1", "mute"}

	for _, model := range modelsWithSound {
		r := Restrict(model, internal.ParamSound)
		if r == nil {
			t.Errorf("model %q: Restrict(sound) returned nil, expected non-nil", model)
			continue
		}
		for _, v := range validValues {
			if err := r.ValidateString(internal.ParamSound, v); err != nil {
				t.Errorf("model %q: sound=%q should be valid, got: %v", model, v, err)
			}
		}
		for _, v := range invalidValues {
			if err := r.ValidateString(internal.ParamSound, v); err == nil {
				t.Errorf("model %q: sound=%q should be rejected", model, v)
			}
		}
	}
}

func TestRestrict_Sound_NilForOlderModels(t *testing.T) {
	modelsWithoutSound := []string{
		internal.ModelKlingV21Video,
		internal.ModelKlingV25Turbo,
		internal.ModelKlingVideoO1,
	}
	for _, model := range modelsWithoutSound {
		r := Restrict(model, internal.ParamSound)
		if r != nil {
			t.Errorf("model %q: Restrict(sound) should return nil for models without sound support", model)
		}
	}
}

// --- Restrict: keep_original_sound ---

func TestRestrict_KeepOriginalSound(t *testing.T) {
	modelsWithKeepSound := []string{
		internal.ModelKlingVideoO1,
		internal.ModelKlingV26, internal.ModelKlingV27,
		internal.ModelKlingV28, internal.ModelKlingV29,
		internal.ModelKlingV3Omni,
	}
	validValues := []string{"yes", "no"}
	invalidValues := []string{"on", "off", "true", "1"}

	for _, model := range modelsWithKeepSound {
		r := Restrict(model, internal.ParamKeepOriginalSound)
		if r == nil {
			t.Errorf("model %q: Restrict(keep_original_sound) returned nil, expected non-nil", model)
			continue
		}
		for _, v := range validValues {
			if err := r.ValidateString(internal.ParamKeepOriginalSound, v); err != nil {
				t.Errorf("model %q: keep_original_sound=%q should be valid, got: %v", model, v, err)
			}
		}
		for _, v := range invalidValues {
			if err := r.ValidateString(internal.ParamKeepOriginalSound, v); err == nil {
				t.Errorf("model %q: keep_original_sound=%q should be rejected", model, v)
			}
		}
	}
}

// --- Restrict: unrestricted params return nil ---

func TestRestrict_UnrestrictedParams(t *testing.T) {
	unrestricted := []string{
		internal.ParamPrompt,
		internal.ParamInputReference,
		internal.ParamImageTail,
		internal.ParamNegativePrompt,
		internal.ParamImageList,
		internal.ParamVideoList,
		"unknown_param",
	}
	for _, model := range allVideoModels {
		for _, name := range unrestricted {
			r := Restrict(model, name)
			if r != nil {
				t.Errorf("model %q, param %q: expected nil Restriction, got %+v", model, name, r)
			}
		}
	}
}

// --- Validate: required params ---

func TestValidate_V21_RequiresInputReference(t *testing.T) {
	p := &mockChecker{vals: map[string]string{internal.ParamPrompt: "test"}}
	if err := Validate(internal.ModelKlingV21Video, p); err != ErrInputReferenceRequired {
		t.Errorf("expected ErrInputReferenceRequired, got: %v", err)
	}

	p.vals[internal.ParamInputReference] = "http://example.com/img.jpg"
	if err := Validate(internal.ModelKlingV21Video, p); err != nil {
		t.Errorf("expected nil with input_reference set, got: %v", err)
	}
}

func TestValidate_OtherModels_NoInputReferenceRequired(t *testing.T) {
	models := []string{
		internal.ModelKlingV25Turbo, internal.ModelKlingVideoO1,
		internal.ModelKlingV26, internal.ModelKlingV3, internal.ModelKlingV3Omni,
	}
	for _, model := range models {
		p := &mockChecker{vals: map[string]string{internal.ParamPrompt: "test"}}
		if err := Validate(model, p); err != nil {
			t.Errorf("model %q: expected nil, got: %v", model, err)
		}
	}
}

// --- Validate: keyframe constraints ---

func TestValidate_Keyframe_RequiresModePro(t *testing.T) {
	for _, model := range allVideoModels {
		p := &mockChecker{vals: map[string]string{
			internal.ParamPrompt:         "test",
			internal.ParamInputReference: "http://example.com/img.jpg",
			internal.ParamImageTail:      "http://example.com/tail.jpg",
			internal.ParamMode:           "std",
			internal.ParamSeconds:        "10",
		}}
		if err := Validate(model, p); err != ErrKeyframeModeRequired {
			t.Errorf("model %q: expected ErrKeyframeModeRequired with mode=std + image_tail, got: %v", model, err)
		}
	}
}

func TestValidate_Keyframe_V21_RequiresSeconds10(t *testing.T) {
	p := &mockChecker{vals: map[string]string{
		internal.ParamPrompt:         "test",
		internal.ParamInputReference: "http://example.com/img.jpg",
		internal.ParamImageTail:      "http://example.com/tail.jpg",
		internal.ParamMode:           "pro",
		internal.ParamSeconds:        "5",
	}}
	if err := Validate(internal.ModelKlingV21Video, p); err != ErrKeyframeSecondsRequired {
		t.Errorf("expected ErrKeyframeSecondsRequired, got: %v", err)
	}

	p.vals[internal.ParamSeconds] = "10"
	if err := Validate(internal.ModelKlingV21Video, p); err != nil {
		t.Errorf("expected nil with seconds=10, got: %v", err)
	}
}

func TestValidate_Keyframe_NonV21_AnySeconds(t *testing.T) {
	models := []string{
		internal.ModelKlingV25Turbo, internal.ModelKlingVideoO1,
		internal.ModelKlingV26, internal.ModelKlingV3, internal.ModelKlingV3Omni,
	}
	for _, model := range models {
		p := &mockChecker{vals: map[string]string{
			internal.ParamPrompt:    "test",
			internal.ParamImageTail: "http://example.com/tail.jpg",
			internal.ParamMode:      "pro",
			internal.ParamSeconds:   "5",
		}}
		if err := Validate(model, p); err != nil {
			t.Errorf("model %q: expected nil (non-V21 keyframe allows seconds!=10), got: %v", model, err)
		}
	}
}

func TestValidate_NoImageTail_SkipsKeyframeCheck(t *testing.T) {
	for _, model := range allVideoModels {
		p := &mockChecker{vals: map[string]string{
			internal.ParamPrompt:         "test",
			internal.ParamInputReference: "http://example.com/img.jpg",
			internal.ParamMode:           "std",
		}}
		if err := Validate(model, p); err != nil {
			t.Errorf("model %q: expected nil without image_tail, got: %v", model, err)
		}
	}
}

// --- Schema completeness ---

func TestSchemaForVideo_AllModelsReturnFields(t *testing.T) {
	for _, model := range allVideoModels {
		fields := SchemaForVideo(model)
		if len(fields) == 0 {
			t.Errorf("model %q: SchemaForVideo returned no fields", model)
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

func TestSchemaForVideo_UnknownModelReturnsFallback(t *testing.T) {
	fields := SchemaForVideo("unknown-model")
	if len(fields) == 0 {
		t.Error("unknown model should return default schema, got empty")
	}
}

// --- mockChecker ---

type mockChecker struct {
	vals map[string]string
}

func (m *mockChecker) HasNonEmptyString(name string) bool {
	v, ok := m.vals[name]
	return ok && v != ""
}

func (m *mockChecker) GetString(name string) string {
	return m.vals[name]
}
