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

package vidu

import (
	"errors"
	"testing"
)

func TestIsVideoModel(t *testing.T) {
	if !IsVideoModel(ModelViduQ1) {
		t.Fatal("expected vidu-q1 to be a video model")
	}
	if !IsVideoModel(ModelViduQ2) {
		t.Fatal("expected vidu-q2 to be a video model")
	}
	if !IsVideoModel(ModelViduQ2Turbo) {
		t.Fatal("expected viduq2-turbo to be a video model")
	}
	if !IsVideoModel(ModelViduQ2Pro) {
		t.Fatal("expected viduq2-pro to be a video model")
	}
	if IsVideoModel("kling-v2-1") {
		t.Fatal("expected kling-v2-1 not to be a vidu model")
	}
}

func TestBuildVideoParamsRoutes(t *testing.T) {
	tests := []struct {
		name      string
		model     string
		setup     func(p *Params)
		wantRoute GenerationRoute
	}{
		{
			name:  "q1 text",
			model: ModelViduQ1,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat running")
			},
			wantRoute: RouteTextToVideo,
		},
		{
			name:  "q1 reference",
			model: ModelViduQ1,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat running")
				p.Set(ParamReferenceImageURLs, []string{"https://example.com/1.png"})
			},
			wantRoute: RouteReferenceToVideo,
		},
		{
			name:  "q2 image",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat running")
				p.Set(ParamImageURL, "https://example.com/1.png")
			},
			wantRoute: RouteImageToVideo,
		},
		{
			name:  "q2 start end",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a dragon lands")
				p.Set(ParamStartImageURL, "https://example.com/start.png")
				p.Set(ParamEndImageURL, "https://example.com/end.png")
			},
			wantRoute: RouteStartEndToVideo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParams()
			tt.setup(p)
			vp, err := BuildVideoParams(tt.model, p)
			if err != nil {
				t.Fatalf("BuildVideoParams error: %v", err)
			}
			if got := vp.Route(); got != tt.wantRoute {
				t.Fatalf("Route() = %s, want %s", got, tt.wantRoute)
			}
		})
	}
}

func TestBuildVideoParamsValidation(t *testing.T) {
	tests := []struct {
		name    string
		model   string
		setup   func(p *Params)
		wantErr error
	}{
		{
			name:  "q1 image route not supported",
			model: ModelViduQ1,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat running")
				p.Set(ParamImageURL, "https://example.com/1.png")
			},
			wantErr: ErrRouteNotSupported,
		},
		{
			name:  "missing end image",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat running")
				p.Set(ParamStartImageURL, "https://example.com/start.png")
			},
			wantErr: ErrStartEndPairRequired,
		},
		{
			name:  "reference conflict",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat running")
				p.Set(ParamReferenceImageURLs, []string{"https://example.com/1.png"})
				p.Set(ParamSubjects, []Subject{{ID: "cat", Images: []string{"https://example.com/2.png"}}})
			},
			wantErr: ErrReferenceInputsConflict,
		},
		{
			name:  "invalid duration",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat running")
				p.Set(ParamDuration, 0)
			},
			wantErr: ErrInvalidDuration,
		},
		{
			name:  "q1 duration must be 5 seconds",
			model: ModelViduQ1,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat running")
				p.Set(ParamDuration, 4)
			},
			wantErr: ErrInvalidQ1Duration,
		},
		{
			name:  "malformed reference_image_urls",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat running")
				p.Set(ParamReferenceImageURLs, 12345)
			},
			wantErr: ErrInvalidReferenceImageURLs,
		},
		{
			name:  "malformed subjects",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat running")
				p.Set(ParamSubjects, "not-a-list")
			},
			wantErr: ErrInvalidSubjects,
		},
		{
			name:  "prompt too long",
			model: ModelViduQ2,
			setup: func(p *Params) {
				runes := make([]rune, 2001)
				for i := range runes {
					runes[i] = 'a'
				}
				p.Set(ParamPrompt, string(runes))
			},
			wantErr: ErrPromptTooLong,
		},
		{
			name:  "reference_image_urls empty",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat")
				p.Set(ParamReferenceImageURLs, []string{})
			},
			wantErr: ErrInvalidReferenceCount,
		},
		{
			name:  "reference_image_urls too many",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat")
				urls := make([]string, 8)
				for i := range urls {
					urls[i] = "https://example.com/" + string(rune('0'+i)) + ".png"
				}
				p.Set(ParamReferenceImageURLs, urls)
			},
			wantErr: ErrInvalidReferenceCount,
		},
		{
			name:  "subjects empty",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat")
				p.Set(ParamSubjects, []Subject{})
			},
			wantErr: ErrInvalidSubjectsCount,
		},
		{
			name:  "subjects too many",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat")
				subs := make([]Subject, 8)
				for i := range subs {
					subs[i] = Subject{ID: string(rune('a' + i)), Images: []string{"https://example.com/1.png"}}
				}
				p.Set(ParamSubjects, subs)
			},
			wantErr: ErrInvalidSubjectsCount,
		},
		{
			name:  "subject too many images",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat")
				p.Set(ParamSubjects, []Subject{{
					ID:     "cat",
					Images: []string{"https://a.png", "https://b.png", "https://c.png", "https://d.png"},
				}})
			},
			wantErr: ErrInvalidSubjectImages,
		},
		{
			name:  "q2 text-to-video duration too low",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat running")
				p.Set(ParamDuration, 0)
			},
			wantErr: ErrInvalidDuration,
		},
		{
			name:  "q2 text-to-video duration too high",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat running")
				p.Set(ParamDuration, 11)
			},
			wantErr: ErrInvalidQ2Duration,
		},
		{
			name:  "q2 reference duration not 5",
			model: ModelViduQ2,
			setup: func(p *Params) {
				p.Set(ParamPrompt, "a cat")
				p.Set(ParamReferenceImageURLs, []string{"https://example.com/1.png"})
				p.Set(ParamDuration, 4)
			},
			wantErr: ErrInvalidQ2Duration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParams()
			tt.setup(p)
			_, err := BuildVideoParams(tt.model, p)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestRestrict(t *testing.T) {
	tests := []struct {
		model    string
		name     string
		wantNil  bool
		validVal string
		badVal   string
	}{
		{ModelViduQ1, ParamResolution, false, Resolution1080p, "4k"},
		{ModelViduQ2, ParamResolution, false, Resolution720p, "4k"},
		{ModelViduQ1, ParamMovementAmplitude, false, MovementAuto, "invalid"},
		{ModelViduQ2, ParamMovementAmplitude, false, MovementSmall, "invalid"},
		{ModelViduQ1, ParamAspectRatio, false, AspectRatio16_9, AspectRatio3_4},
		{ModelViduQ2, ParamAspectRatio, false, AspectRatio3_4, "2:1"},
		{ModelViduQ1, ParamStyle, false, StyleGeneral, "invalid"},
		{ModelViduQ2, ParamStyle, true, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.model+"_"+tt.name, func(t *testing.T) {
			r := Restrict(tt.model, tt.name)
			if tt.wantNil {
				if r != nil {
					t.Fatalf("Restrict(%q, %q) expected nil, got %+v", tt.model, tt.name, r)
				}
				return
			}
			if r == nil {
				t.Fatalf("Restrict(%q, %q) expected non-nil", tt.model, tt.name)
			}
			if err := r.ValidateString(tt.name, tt.validVal); err != nil {
				t.Errorf("ValidateString(%q) valid value: %v", tt.validVal, err)
			}
			if err := r.ValidateString(tt.name, tt.badVal); err == nil {
				t.Errorf("ValidateString(%q) bad value: expected error, got nil", tt.badVal)
			}
		})
	}
}
