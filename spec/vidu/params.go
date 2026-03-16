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
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/types"
)

// Param name constants.
const (
	ParamPrompt            = "prompt"
	ParamSeed              = "seed"
	ParamDuration          = "duration"
	ParamResolution        = "resolution"
	ParamMovementAmplitude = "movement_amplitude"
	ParamWatermark         = "watermark"
	ParamAspectRatio       = "aspect_ratio"
	ParamBGM               = "bgm"
	ParamStyle             = "style"
	ParamAudio             = "audio"
	ParamIsRec             = "is_rec"
	ParamVoiceID           = "voice_id"

	ParamReferenceImageURLs = "reference_image_urls"
	ParamSubjects           = "subjects"
	ParamImageURL           = "image_url"
	ParamStartImageURL      = "start_image_url"
	ParamEndImageURL        = "end_image_url"
)

// Resolution options.
const (
	Resolution720p  = "720p"
	Resolution1080p = "1080p"
)

// Aspect ratio options.
const (
	AspectRatio16_9 = "16:9"
	AspectRatio9_16 = "9:16"
	AspectRatio3_4  = "3:4"
	AspectRatio4_3  = "4:3"
	AspectRatio1_1  = "1:1"
)

// Movement amplitude options.
const (
	MovementAuto   = "auto"
	MovementSmall  = "small"
	MovementMedium = "medium"
	MovementLarge  = "large"
)

// Style options (q1 only; q2 ignores).
const (
	StyleGeneral = "general"
	StyleAnime   = "anime"
)

var (
	// ErrPromptRequired is returned when prompt is not set or empty.
	ErrPromptRequired = errors.New("vidu: prompt is required")

	// ErrUnsupportedModel is returned when model is not vidu-q1, vidu-q2, viduq2-turbo or viduq2-pro.
	ErrUnsupportedModel = errors.New("vidu: unsupported video model")

	// ErrRouteNotSupported is returned when selected mode is unsupported by model.
	ErrRouteNotSupported = errors.New("vidu: generation route is not supported by model")

	// ErrStartEndPairRequired is returned when start/end input is incomplete.
	ErrStartEndPairRequired = errors.New("vidu: start_image_url and end_image_url must be set together")

	// ErrReferenceInputsConflict is returned when two reference formats are mixed.
	ErrReferenceInputsConflict = errors.New("vidu: reference_image_urls and subjects cannot be used together")

	// ErrConflictingGenerationMode is returned when route-specific params are mixed.
	ErrConflictingGenerationMode = errors.New("vidu: conflicting generation mode parameters")

	// ErrInvalidDuration is returned when duration <= 0.
	ErrInvalidDuration = errors.New("vidu: duration must be greater than 0")

	// ErrInvalidQ1Duration is returned when vidu-q1 duration is not 5 seconds.
	ErrInvalidQ1Duration = errors.New("vidu: vidu-q1 duration must be 5 seconds")

	// ErrInvalidSubjects is returned when subjects payload is malformed.
	ErrInvalidSubjects = errors.New("vidu: subjects payload is invalid")

	// ErrInvalidReferenceImageURLs is returned when reference_image_urls payload is malformed.
	ErrInvalidReferenceImageURLs = errors.New("vidu: reference_image_urls payload is invalid")

	// ErrPromptTooLong is returned when prompt exceeds 2000 characters.
	ErrPromptTooLong = errors.New("vidu: prompt exceeds 2000 characters")

	// ErrInvalidReferenceCount is returned when reference_image_urls has not 1-7 images.
	ErrInvalidReferenceCount = errors.New("vidu: reference_image_urls must have 1-7 images")

	// ErrInvalidSubjectsCount is returned when subjects has not 1-7 items.
	ErrInvalidSubjectsCount = errors.New("vidu: subjects must have 1-7 items")

	// ErrInvalidSubjectImages is returned when subject images violate 1-3 per subject, 1-7 total.
	ErrInvalidSubjectImages = errors.New("vidu: each subject has at most 3 images, total 1-7")

	// ErrInvalidQ2Duration is returned when vidu-q2 duration is out of range.
	ErrInvalidQ2Duration = errors.New("vidu: vidu-q2 text-to-video duration must be 1-10, others must be 5")
)

var (
	limitResolution        = &xai.StringEnum{Values: []string{Resolution720p, Resolution1080p}}
	limitMovementAmplitude = &xai.StringEnum{Values: []string{MovementAuto, MovementSmall, MovementMedium, MovementLarge}}
	limitAspectRatioQ1     = &xai.StringEnum{Values: []string{AspectRatio16_9, AspectRatio9_16, AspectRatio1_1}}
	limitAspectRatioQ2     = &xai.StringEnum{Values: []string{AspectRatio16_9, AspectRatio9_16, AspectRatio3_4, AspectRatio4_3, AspectRatio1_1}}
	limitStyle             = &xai.StringEnum{Values: []string{StyleGeneral, StyleAnime}}
)

// GenerationRoute is the selected endpoint mode for a request.
type GenerationRoute string

const (
	RouteTextToVideo      GenerationRoute = "text_to_video"
	RouteReferenceToVideo GenerationRoute = "reference_to_video"
	RouteImageToVideo     GenerationRoute = "image_to_video"
	RouteStartEndToVideo  GenerationRoute = "start_end_to_video"
)

// Subject defines a named reference entity in reference-to-video mode.
type Subject struct {
	ID      string   `json:"id"`
	Images  []string `json:"images"`
	VoiceID string   `json:"voice_id,omitempty"`
}

// VideoParams is the typed Vidu video request params.
type VideoParams struct {
	ModelName string

	Prompt            string
	Seed              *int
	Duration          *int
	Resolution        string
	MovementAmplitude string
	Watermark         *bool

	AspectRatio string
	BGM         *bool
	Style       string
	Audio       *bool
	IsRec       *bool
	VoiceID     string

	ReferenceImageURLs []string
	Subjects           []Subject
	ImageURL           string
	StartImageURL      string
	EndImageURL        string
}

// Model returns the model name.
func (p *VideoParams) Model() string {
	return p.ModelName
}

// Route returns the selected generation route derived from params.
func (p *VideoParams) Route() GenerationRoute {
	hasStartEnd := strings.TrimSpace(p.StartImageURL) != "" || strings.TrimSpace(p.EndImageURL) != ""
	if hasStartEnd {
		return RouteStartEndToVideo
	}
	if strings.TrimSpace(p.ImageURL) != "" {
		return RouteImageToVideo
	}
	if len(p.ReferenceImageURLs) > 0 || len(p.Subjects) > 0 {
		return RouteReferenceToVideo
	}
	return RouteTextToVideo
}

// BuildVideoParams builds typed VideoParams from Params for the given model.
func BuildVideoParams(model string, p *Params) (*VideoParams, error) {
	if p == nil {
		return nil, errors.New("vidu: params is nil")
	}

	m := normalizeModel(model)
	if !IsVideoModel(m) {
		return nil, fmt.Errorf("%w %q", ErrUnsupportedModel, model)
	}

	_, hasRawReferenceImageURLs := p.Get(ParamReferenceImageURLs)
	_, hasRawSubjects := p.Get(ParamSubjects)
	parsedReferenceImageURLs := p.GetStringSlice(ParamReferenceImageURLs)
	parsedSubjects := p.GetSubjects(ParamSubjects)

	// If caller explicitly sets reference inputs but parsing failed or produced empty,
	// fail fast instead of silently falling back to text-to-video route.
	if hasRawReferenceImageURLs {
		if parsedReferenceImageURLs == nil {
			return nil, ErrInvalidReferenceImageURLs
		}
		if len(parsedReferenceImageURLs) == 0 {
			return nil, ErrInvalidReferenceCount
		}
	}
	if hasRawSubjects {
		if parsedSubjects == nil {
			return nil, ErrInvalidSubjects
		}
		if len(parsedSubjects) == 0 {
			return nil, ErrInvalidSubjectsCount
		}
	}

	vp := &VideoParams{
		ModelName:         m,
		Prompt:            p.GetString(ParamPrompt),
		Seed:              p.GetInt(ParamSeed),
		Duration:          p.GetInt(ParamDuration),
		Resolution:        p.GetString(ParamResolution),
		MovementAmplitude: p.GetString(ParamMovementAmplitude),
		Watermark:         p.GetBool(ParamWatermark),

		AspectRatio: p.GetString(ParamAspectRatio),
		BGM:         p.GetBool(ParamBGM),
		Style:       p.GetString(ParamStyle),
		Audio:       p.GetBool(ParamAudio),
		IsRec:       p.GetBool(ParamIsRec),
		VoiceID:     p.GetString(ParamVoiceID),

		ReferenceImageURLs: parsedReferenceImageURLs,
		Subjects:           parsedSubjects,
		ImageURL:           p.GetString(ParamImageURL),
		StartImageURL:      p.GetString(ParamStartImageURL),
		EndImageURL:        p.GetString(ParamEndImageURL),
	}

	if err := Validate(vp); err != nil {
		return nil, err
	}
	return vp, nil
}

// Validate runs model-specific and route-specific validation.
func Validate(p *VideoParams) error {
	if p == nil {
		return errors.New("vidu: video params is nil")
	}
	if strings.TrimSpace(p.Prompt) == "" {
		return ErrPromptRequired
	}
	if utf8.RuneCountInString(p.Prompt) > 2000 {
		return ErrPromptTooLong
	}
	if p.Duration != nil && *p.Duration <= 0 {
		return ErrInvalidDuration
	}

	hasRefs := len(p.ReferenceImageURLs) > 0 || len(p.Subjects) > 0
	hasImage := strings.TrimSpace(p.ImageURL) != ""
	hasStart := strings.TrimSpace(p.StartImageURL) != ""
	hasEnd := strings.TrimSpace(p.EndImageURL) != ""
	hasStartEnd := hasStart || hasEnd

	if len(p.ReferenceImageURLs) > 0 && len(p.Subjects) > 0 {
		return ErrReferenceInputsConflict
	}
	if hasImage && hasRefs {
		return ErrConflictingGenerationMode
	}
	if hasStartEnd && (hasImage || hasRefs) {
		return ErrConflictingGenerationMode
	}
	if hasStart != hasEnd {
		return ErrStartEndPairRequired
	}

	if len(p.ReferenceImageURLs) > 0 {
		if n := len(p.ReferenceImageURLs); n < 1 || n > 7 {
			return ErrInvalidReferenceCount
		}
	}

	if len(p.Subjects) > 0 {
		if n := len(p.Subjects); n < 1 || n > 7 {
			return ErrInvalidSubjectsCount
		}
		totalImages := 0
		for i, sb := range p.Subjects {
			if strings.TrimSpace(sb.ID) == "" || len(sb.Images) == 0 {
				return fmt.Errorf("%w: subjects[%d] requires id and images", ErrInvalidSubjects, i)
			}
			if len(sb.Images) > 3 {
				return ErrInvalidSubjectImages
			}
			totalImages += len(sb.Images)
		}
		if totalImages < 1 || totalImages > 7 {
			return ErrInvalidSubjectImages
		}
	}

	switch p.ModelName {
	case ModelViduQ1:
		if p.Duration != nil && *p.Duration != 5 {
			return ErrInvalidQ1Duration
		}
		switch p.Route() {
		case RouteTextToVideo, RouteReferenceToVideo:
			return nil
		default:
			return fmt.Errorf("%w: %s does not support %s", ErrRouteNotSupported, p.ModelName, p.Route())
		}
	case ModelViduQ2, ModelViduQ2Turbo, ModelViduQ2Pro:
		if p.Duration != nil {
			switch p.Route() {
			case RouteTextToVideo:
				if *p.Duration < 1 || *p.Duration > 10 {
					return ErrInvalidQ2Duration
				}
			default:
				if *p.Duration != 5 {
					return ErrInvalidQ2Duration
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("%w %q", ErrUnsupportedModel, p.ModelName)
	}
}

// SchemaForVideo returns the InputSchema fields for Vidu models.
func SchemaForVideo(model string) []xai.Field {
	if !IsVideoModel(model) {
		return nil
	}
	return []xai.Field{
		{Name: ParamPrompt, Kind: types.String},
		{Name: ParamSeed, Kind: types.Int},
		{Name: ParamDuration, Kind: types.Int},
		{Name: ParamResolution, Kind: types.String},
		{Name: ParamMovementAmplitude, Kind: types.String},
		{Name: ParamAspectRatio, Kind: types.String},
		{Name: ParamStyle, Kind: types.String},
		{Name: ParamBGM, Kind: types.Bool},
		{Name: ParamWatermark, Kind: types.Bool},
		{Name: ParamAudio, Kind: types.Bool},
		{Name: ParamIsRec, Kind: types.Bool},
		{Name: ParamVoiceID, Kind: types.String},
		{Name: ParamReferenceImageURLs, Kind: types.String | types.List},
		{Name: ParamSubjects, Kind: types.List},
		{Name: ParamImageURL, Kind: types.String},
		{Name: ParamStartImageURL, Kind: types.String},
		{Name: ParamEndImageURL, Kind: types.String},
	}
}

// Restrict returns the parameter restriction for a Vidu model.
func Restrict(model, name string) *xai.Restriction {
	if !IsVideoModel(model) {
		return nil
	}
	m := normalizeModel(model)
	switch name {
	case ParamPrompt:
		return &xai.Restriction{Required: true}
	case ParamResolution:
		return &xai.Restriction{Limit: limitResolution}
	case ParamMovementAmplitude:
		return &xai.Restriction{Limit: limitMovementAmplitude}
	case ParamAspectRatio:
		if m == ModelViduQ1 {
			return &xai.Restriction{Limit: limitAspectRatioQ1}
		}
		return &xai.Restriction{Limit: limitAspectRatioQ2}
	case ParamStyle:
		// style only effective for q1; q2 ignores, return nil to allow any
		if m == ModelViduQ1 {
			return &xai.Restriction{Limit: limitStyle}
		}
		return nil
	default:
		return nil
	}
}

// Params implements xai.Params with map-based storage.
type Params struct {
	m map[string]any
}

// NewParams creates a new Params instance.
func NewParams() *Params {
	return &Params{m: make(map[string]any)}
}

// Set sets a parameter. Implements xai.Params.
func (p *Params) Set(name string, val any) xai.Params {
	p.m[name] = val
	return p
}

// Export returns a copy of params for executor implementations.
func (p *Params) Export() map[string]any {
	out := make(map[string]any, len(p.m))
	for k, v := range p.m {
		out[k] = v
	}
	return out
}

// Get returns the raw value for a parameter.
func (p *Params) Get(name string) (any, bool) {
	v, ok := p.m[name]
	return v, ok
}

// HasNonEmptyString returns true if param is a non-empty string.
func (p *Params) HasNonEmptyString(name string) bool {
	s := p.GetString(name)
	return strings.TrimSpace(s) != ""
}

// GetString returns a trimmed string value, or "" if missing/invalid.
func (p *Params) GetString(name string) string {
	v, ok := p.m[name]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

// GetInt returns an int pointer, or nil if missing/invalid.
func (p *Params) GetInt(name string) *int {
	v, ok := p.m[name]
	if !ok {
		return nil
	}
	switch x := v.(type) {
	case int:
		v := x
		return &v
	case int64:
		v := int(x)
		return &v
	case float64:
		v := int(x)
		return &v
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return nil
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil
		}
		return &n
	default:
		return nil
	}
}

// GetBool returns a bool pointer, or nil if missing/invalid.
func (p *Params) GetBool(name string) *bool {
	v, ok := p.m[name]
	if !ok {
		return nil
	}
	switch x := v.(type) {
	case bool:
		v := x
		return &v
	case string:
		s := strings.ToLower(strings.TrimSpace(x))
		switch s {
		case "true", "1", "yes", "on":
			v := true
			return &v
		case "false", "0", "no", "off":
			v := false
			return &v
		default:
			return nil
		}
	default:
		return nil
	}
}

// GetStringSlice returns []string parsed from string/[]string/[]interface{}.
func (p *Params) GetStringSlice(name string) []string {
	v, ok := p.m[name]
	if !ok {
		return nil
	}
	switch x := v.(type) {
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return nil
		}
		return []string{s}
	case []string:
		return sanitizeStrings(x)
	case []interface{}:
		out := make([]string, 0, len(x))
		for _, item := range x {
			s, ok := item.(string)
			if !ok {
				continue
			}
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

// GetSubjects returns []Subject parsed from []Subject/[]interface{}.
func (p *Params) GetSubjects(name string) []Subject {
	v, ok := p.m[name]
	if !ok {
		return nil
	}

	switch x := v.(type) {
	case []Subject:
		return sanitizeSubjects(x)
	case []map[string]string:
		return parseSubjectsFromStringMaps(x)
	case []map[string]interface{}:
		return parseSubjectsFromMaps(x)
	case []interface{}:
		return parseSubjectsFromSlice(x)
	default:
		return nil
	}
}

func sanitizeStrings(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func sanitizeSubjects(in []Subject) []Subject {
	out := make([]Subject, 0, len(in))
	for _, sb := range in {
		sb.ID = strings.TrimSpace(sb.ID)
		sb.VoiceID = strings.TrimSpace(sb.VoiceID)
		sb.Images = sanitizeStrings(sb.Images)
		if sb.ID == "" && len(sb.Images) == 0 {
			continue
		}
		out = append(out, sb)
	}
	return out
}

func parseSubjectsFromMaps(items []map[string]interface{}) []Subject {
	out := make([]Subject, 0, len(items))
	for _, m := range items {
		sb := Subject{}
		if id, ok := m["id"].(string); ok {
			sb.ID = strings.TrimSpace(id)
		}
		if voiceID, ok := m["voice_id"].(string); ok {
			sb.VoiceID = strings.TrimSpace(voiceID)
		}
		sb.Images = parseImagesField(m["images"])
		if sb.ID == "" && len(sb.Images) == 0 {
			continue
		}
		out = append(out, sb)
	}
	return out
}

func parseSubjectsFromSlice(items []interface{}) []Subject {
	maps := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		switch m := item.(type) {
		case map[string]interface{}:
			maps = append(maps, m)
		case map[string]string:
			mi := map[string]interface{}{
				"id":       m["id"],
				"voice_id": m["voice_id"],
				"images":   m["images"],
			}
			maps = append(maps, mi)
		}
	}
	return parseSubjectsFromMaps(maps)
}

func parseSubjectsFromStringMaps(items []map[string]string) []Subject {
	out := make([]Subject, 0, len(items))
	for _, m := range items {
		sb := Subject{
			ID:      strings.TrimSpace(m["id"]),
			VoiceID: strings.TrimSpace(m["voice_id"]),
			Images:  parseImagesField(m["images"]),
		}
		if sb.ID == "" && len(sb.Images) == 0 {
			continue
		}
		out = append(out, sb)
	}
	return out
}

func parseImagesField(v any) []string {
	switch x := v.(type) {
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return nil
		}
		return []string{s}
	case []string:
		return sanitizeStrings(x)
	case []interface{}:
		out := make([]string, 0, len(x))
		for _, item := range x {
			s, ok := item.(string)
			if !ok {
				continue
			}
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func normalizeModel(model string) string {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case ModelViduQ1:
		return ModelViduQ1
	case ModelViduQ2:
		return ModelViduQ2
	case ModelViduQ2Turbo:
		return ModelViduQ2Turbo
	case ModelViduQ2Pro:
		return ModelViduQ2Pro
	default:
		return strings.ToLower(strings.TrimSpace(model))
	}
}

func isNilLike(v any) bool {
	if v == nil {
		return true
	}
	switch x := v.(type) {
	case []string:
		return len(x) == 0
	case []Subject:
		return len(x) == 0
	case []map[string]string:
		return len(x) == 0
	case []map[string]interface{}:
		return len(x) == 0
	case []interface{}:
		return len(x) == 0
	}
	return false
}
