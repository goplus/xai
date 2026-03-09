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

package gemini

import (
	"fmt"
	"strings"

	"google.golang.org/genai"
)

// Veo duration constraints: Veo 2 supports 5-8s; Veo 3 supports 4, 6, 8s.
var (
	veo2DurationSeconds = map[int32]bool{5: true, 6: true, 7: true, 8: true}
	veo3DurationSeconds = map[int32]bool{4: true, 6: true, 8: true}
)

const (
	veo2Prefix = "veo-2."
	veo3Prefix = "veo-3."
	maxSeed    = 4294967295
)

// validateGenVideoConfig validates GenerateVideosConfig and GenerateVideosSource
// against Veo API constraints before calling the backend.
func validateGenVideoConfig(model string, source *genai.GenerateVideosSource, config *genai.GenerateVideosConfig) error {
	if source == nil {
		return fmt.Errorf("xai: video source is required")
	}
	prompt := strings.TrimSpace(source.Prompt)
	if source.Image == nil && source.Video == nil && prompt == "" {
		return fmt.Errorf("xai: Prompt is required for text-to-video")
	}
	if config == nil {
		return nil
	}

	// DurationSeconds: Veo 2 supports 5-8; Veo 3 supports 4, 6, 8
	if config.DurationSeconds != nil {
		d := *config.DurationSeconds
		m := strings.ToLower(model)
		if strings.HasPrefix(m, veo3Prefix) {
			if !veo3DurationSeconds[d] {
				return fmt.Errorf("xai: DurationSeconds %d not in [4, 6, 8] for Veo 3 models", d)
			}
		} else if strings.HasPrefix(m, veo2Prefix) {
			if !veo2DurationSeconds[d] {
				return fmt.Errorf("xai: DurationSeconds %d not in [5, 6, 7, 8] for Veo 2 models", d)
			}
		} else {
			// Unknown model: allow 4, 5, 6, 7, 8
			allowed := veo2DurationSeconds[d] || veo3DurationSeconds[d]
			if !allowed {
				return fmt.Errorf("xai: DurationSeconds %d not in [4, 5, 6, 7, 8]", d)
			}
		}
	}

	// NumberOfVideos: 0 (use default) or 1-4
	if config.NumberOfVideos != 0 && (config.NumberOfVideos < 1 || config.NumberOfVideos > 4) {
		return fmt.Errorf("xai: NumberOfVideos %d not in [1, 4]", config.NumberOfVideos)
	}

	// Seed: 0-4294967295
	if config.Seed != nil {
		s := int64(*config.Seed)
		if s < 0 || s > maxSeed {
			return fmt.Errorf("xai: Seed %d not in [0, %d]", *config.Seed, maxSeed)
		}
	}

	// String enum validation via restriction_genVideo
	schema := newInputSchema(&struct {
		genai.GenerateVideosSource
		genai.GenerateVideosConfig
	}{}, restriction_genVideo)

	for name, val := range map[string]string{
		"AspectRatio":        config.AspectRatio,
		"Resolution":         config.Resolution,
		"PersonGeneration":   config.PersonGeneration,
		"CompressionQuality": string(config.CompressionQuality),
	} {
		if val == "" {
			continue
		}
		if r := schema.Restrict(name); r != nil {
			if err := r.ValidateString(name, val); err != nil {
				return err
			}
		}
	}
	return nil
}
