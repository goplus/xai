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

package qiniu

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"google.golang.org/genai"
)

func (p *backend) GenerateVideosFromSource(ctx context.Context, model string, source *genai.GenerateVideosSource, config *genai.GenerateVideosConfig) (*genai.GenerateVideosOperation, error) {
	body, err := buildVideoGenerationRequest(model, source, config)
	if err != nil {
		return nil, err
	}

	var resp videoTaskSubmitResponse
	if err := p.client.postJSONAt(ctx, p.baseURL(genVideoHTTPOptions(config)), endpointVideosGenerate, body, &resp); err != nil {
		return nil, err
	}

	taskID := resp.taskID()
	if taskID == "" {
		return nil, fmt.Errorf("qiniu: empty video task id")
	}

	done, failed := classifyVideoTaskStatus(resp.Status, resp.Data)
	op := &genai.GenerateVideosOperation{
		Name:     taskID,
		Done:     done,
		Metadata: resp.metadata(),
	}
	if done {
		op.Response = resp.toGenerateVideosResponse()
		if failed {
			op.Error = videoTaskError(resp.Status, resp.Message)
		}
	}
	return op, nil
}

func (p *backend) GetVideosOperation(ctx context.Context, op *genai.GenerateVideosOperation, config *genai.GetOperationConfig) (*genai.GenerateVideosOperation, error) {
	if op == nil {
		return nil, fmt.Errorf("qiniu: nil video operation")
	}
	taskID := videoTaskID(op.Name)
	if taskID == "" {
		return nil, fmt.Errorf("qiniu: empty video task id")
	}

	var resp videoTaskStatusResponse
	endpoint := endpointVideosGenerate + "/" + taskID
	if err := p.client.getJSONAt(ctx, p.baseURL(getVideoHTTPOptions(config)), endpoint, &resp); err != nil {
		return nil, err
	}

	done, failed := classifyVideoTaskStatus(resp.Status, resp.Data)
	ret := &genai.GenerateVideosOperation{
		Name:     coalesceString(resp.taskID(), taskID),
		Done:     done,
		Metadata: resp.metadata(),
	}
	if done {
		ret.Response = resp.toGenerateVideosResponse()
		if failed {
			ret.Error = videoTaskError(resp.Status, resp.Message)
		}
	}
	return ret, nil
}

func genVideoHTTPOptions(cfg *genai.GenerateVideosConfig) *genai.HTTPOptions {
	if cfg == nil {
		return nil
	}
	return cfg.HTTPOptions
}

func getVideoHTTPOptions(cfg *genai.GetOperationConfig) *genai.HTTPOptions {
	if cfg == nil {
		return nil
	}
	return cfg.HTTPOptions
}

func buildVideoGenerationRequest(model string, source *genai.GenerateVideosSource, config *genai.GenerateVideosConfig) (map[string]any, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, fmt.Errorf("qiniu: model is required")
	}
	if source == nil {
		return nil, fmt.Errorf("qiniu: video source is required")
	}

	prompt := strings.TrimSpace(source.Prompt)
	hasRefImgs := config != nil && len(config.ReferenceImages) > 0
	if hasRefImgs {
		if source.Image != nil || source.Video != nil || config.LastFrame != nil {
			return nil, fmt.Errorf("qiniu: ReferenceImages cannot be used with Image, Video, or LastFrame")
		}
		if prompt == "" {
			return nil, fmt.Errorf("qiniu: Prompt is required when using ReferenceImages")
		}
		if !modelSupportsReferenceImages(model) {
			return nil, fmt.Errorf("qiniu: referenceImages only supported by veo-2.0-generate-exp and veo-3.1-generate-preview")
		}
		if err := validateReferenceImages(config.ReferenceImages); err != nil {
			return nil, err
		}
	} else if source.Image == nil && source.Video == nil && prompt == "" {
		return nil, fmt.Errorf("qiniu: Prompt, Image, Video, or ReferenceImages is required")
	}

	inst := map[string]any{}
	if prompt != "" {
		inst["prompt"] = prompt
	}
	if hasRefImgs {
		refObjs, err := buildVeoReferenceImages(config.ReferenceImages)
		if err != nil {
			return nil, err
		}
		inst["referenceImages"] = refObjs
	} else {
		if source.Image != nil {
			imgObj, err := buildVeoImageInput(source.Image)
			if err != nil {
				return nil, err
			}
			inst["image"] = imgObj
		}
		if source.Video != nil {
			videoObj, err := buildVeoVideoInput(source.Video)
			if err != nil {
				return nil, err
			}
			inst["video"] = videoObj
		}
		if config != nil && config.LastFrame != nil {
			lastFrameObj, err := buildVeoImageInput(config.LastFrame)
			if err != nil {
				return nil, err
			}
			inst["lastFrame"] = lastFrameObj
		}
	}
	if len(inst) == 0 {
		return nil, fmt.Errorf("qiniu: Prompt, Image, Video, or ReferenceImages is required")
	}

	body := map[string]any{
		"model":     model,
		"instances": []map[string]any{inst},
	}

	if config == nil {
		return body, nil
	}
	if config.Mask != nil {
		return nil, fmt.Errorf("qiniu: Mask is not supported by /v1/videos/generations")
	}

	// referenceImages mode: duration must be 8
	if hasRefImgs && config.DurationSeconds != nil && *config.DurationSeconds != 8 {
		return nil, fmt.Errorf("qiniu: durationSeconds must be 8 when using referenceImages")
	}

	// Enforce Veo API constraints (see veo.md)
	if config.DurationSeconds != nil {
		d := *config.DurationSeconds
		if d != 4 && d != 5 && d != 6 && d != 7 && d != 8 {
			return nil, fmt.Errorf("qiniu: durationSeconds %d not in [4, 5, 6, 7, 8]", d)
		}
	}
	if config.NumberOfVideos != 0 && (config.NumberOfVideos < 1 || config.NumberOfVideos > 4) {
		return nil, fmt.Errorf("qiniu: sampleCount %d not in [1, 4]", config.NumberOfVideos)
	}
	if config.Seed != nil {
		s := int64(*config.Seed)
		if s < 0 || s > 4294967295 {
			return nil, fmt.Errorf("qiniu: seed %d not in [0, 4294967295]", *config.Seed)
		}
	}

	params := make(map[string]any)
	if v := normalizeVeoAspectRatio(config.AspectRatio); v != "" {
		if v != "16:9" && v != "9:16" {
			return nil, fmt.Errorf("qiniu: aspectRatio %q not in [16:9, 9:16]", v)
		}
		params["aspectRatio"] = v
	}
	if hasRefImgs {
		params["durationSeconds"] = 8
	} else if config.DurationSeconds != nil {
		params["durationSeconds"] = *config.DurationSeconds
	}
	if config.NumberOfVideos > 0 {
		params["sampleCount"] = config.NumberOfVideos
	}
	if config.Seed != nil {
		params["seed"] = *config.Seed
	}
	if v := strings.TrimSpace(config.NegativePrompt); v != "" {
		params["negativePrompt"] = v
	}
	if v := strings.TrimSpace(config.PersonGeneration); v != "" {
		v = strings.ToLower(v)
		if v != "dont_allow" && v != "allow_adult" {
			return nil, fmt.Errorf("qiniu: personGeneration %q not in [dont_allow, allow_adult]", v)
		}
		params["personGeneration"] = v
	}
	if config.GenerateAudio != nil {
		params["generateAudio"] = *config.GenerateAudio
	}
	if v := strings.TrimSpace(config.Resolution); v != "" {
		if v != "720p" && v != "1080p" {
			return nil, fmt.Errorf("qiniu: resolution %q not in [720p, 1080p]", v)
		}
		params["resolution"] = v
	}
	if config.FPS != nil {
		params["fps"] = *config.FPS
	}
	if config.EnhancePrompt {
		params["enhancePrompt"] = true
	}
	if v := strings.TrimSpace(string(config.CompressionQuality)); v != "" {
		params["compressionQuality"] = v
	}
	if len(params) > 0 {
		body["parameters"] = params
	}

	// Qiniu Veo uses callback_url instead of Pub/Sub topic.
	if callbackURL := strings.TrimSpace(config.PubsubTopic); callbackURL != "" {
		body["callback_url"] = callbackURL
	}
	return body, nil
}

func modelSupportsReferenceImages(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	return m == "veo-2.0-generate-exp" || m == "veo-3.1-generate-preview"
}

func validateReferenceImages(refs []*genai.VideoGenerationReferenceImage) error {
	if len(refs) == 0 {
		return nil
	}
	assetCount, styleCount := 0, 0
	for _, r := range refs {
		rt := strings.ToLower(strings.TrimSpace(string(r.ReferenceType)))
		if rt == "style" {
			styleCount++
		} else {
			assetCount++
		}
	}
	if styleCount > 0 && assetCount > 0 {
		return fmt.Errorf("qiniu: referenceImages cannot mix asset and style types")
	}
	if styleCount > 1 {
		return fmt.Errorf("qiniu: referenceImages allows at most 1 style image")
	}
	if assetCount > 3 {
		return fmt.Errorf("qiniu: referenceImages allows at most 3 asset images")
	}
	return nil
}

// buildVeoReferenceImages converts []*VideoGenerationReferenceImage to Qiniu instances[0].referenceImages.
func buildVeoReferenceImages(refs []*genai.VideoGenerationReferenceImage) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(refs))
	for _, r := range refs {
		if r == nil || r.Image == nil {
			continue
		}
		imgObj, err := buildVeoImageInput(r.Image)
		if err != nil {
			return nil, err
		}
		refType := strings.ToLower(strings.TrimSpace(string(r.ReferenceType)))
		if refType == "" {
			refType = "asset"
		}
		if refType != "asset" && refType != "style" {
			refType = "asset"
		}
		out = append(out, map[string]any{
			"image":         imgObj,
			"referenceType": refType,
		})
	}
	return out, nil
}

// buildVeoImageInput converts genai.Image to Qiniu instances[0].image format.
// Supports bytesBase64Encoded (from ImageBytes) or uri (from GCSURI).
func buildVeoImageInput(img *genai.Image) (map[string]any, error) {
	if img == nil {
		return nil, fmt.Errorf("qiniu: image is required for image-to-video")
	}
	mime := strings.TrimSpace(img.MIMEType)
	if mime == "" {
		mime = "image/png"
	}
	if len(img.ImageBytes) > 0 {
		return map[string]any{
			"bytesBase64Encoded": base64.StdEncoding.EncodeToString(img.ImageBytes),
			"mimeType":           mime,
		}, nil
	}
	if img.GCSURI != "" {
		return map[string]any{
			"uri":      img.GCSURI,
			"mimeType": mime,
		}, nil
	}
	return nil, fmt.Errorf("qiniu: image must have ImageBytes or GCSURI")
}

// buildVeoVideoInput converts genai.Video to Qiniu instances[0].video format (VideoInput).
// Supports bytesBase64Encoded (from VideoBytes) or uri (from URI). Video size ≤10MB.
func buildVeoVideoInput(vid *genai.Video) (map[string]any, error) {
	if vid == nil {
		return nil, fmt.Errorf("qiniu: video is required for video input")
	}
	mime := strings.TrimSpace(vid.MIMEType)
	if mime == "" {
		mime = "video/mp4"
	}
	if len(vid.VideoBytes) > 0 {
		return map[string]any{
			"bytesBase64Encoded": base64.StdEncoding.EncodeToString(vid.VideoBytes),
			"mimeType":           mime,
		}, nil
	}
	if vid.URI != "" {
		return map[string]any{
			"uri":      vid.URI,
			"mimeType": mime,
		}, nil
	}
	return nil, fmt.Errorf("qiniu: video must have VideoBytes or URI")
}

func normalizeVeoAspectRatio(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "16:9 (landscape)":
		return "16:9"
	case "9:16 (portrait)":
		return "9:16"
	default:
		return strings.TrimSpace(v)
	}
}

func classifyVideoTaskStatus(status string, data *videoTaskData) (done, failed bool) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "completed", "succeeded", "success", "done":
		return true, false
	case "failed", "error", "cancelled", "canceled", "aborted":
		return true, true
	case "pending", "queued", "processing", "running", "in_progress":
		return false, false
	default:
		// Some gateways may not return canonical status; infer completion from outputs.
		if data != nil && len(data.Videos) > 0 {
			return true, false
		}
		return false, false
	}
}

func videoTaskID(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if u, err := url.Parse(name); err == nil && u.Path != "" && u.Host != "" {
		name = u.Path
	}
	name = strings.TrimSpace(strings.TrimSuffix(name, "/"))
	if i := strings.LastIndexByte(name, '/'); i >= 0 {
		name = name[i+1:]
	}
	return strings.TrimSpace(name)
}

func videoTaskError(status, message string) map[string]any {
	msg := strings.TrimSpace(message)
	if msg == "" {
		msg = "video generation failed"
	}
	ret := map[string]any{"message": msg}
	if v := strings.TrimSpace(status); v != "" {
		ret["status"] = v
	}
	return ret
}

func guessVideoMIME(rawURL string) string {
	if strings.HasPrefix(rawURL, "data:video/") {
		if idx := strings.Index(rawURL, ";"); idx > len("data:") {
			return rawURL[len("data:"):idx]
		}
	}
	if parsed, err := url.Parse(rawURL); err == nil {
		switch strings.ToLower(path.Ext(parsed.Path)) {
		case ".webm":
			return "video/webm"
		case ".mp4", ".m4v":
			return "video/mp4"
		}
	}
	return "video/mp4"
}

func coalesceString(items ...string) string {
	for _, item := range items {
		if v := strings.TrimSpace(item); v != "" {
			return v
		}
	}
	return ""
}

type videoTaskSubmitResponse struct {
	ID        string         `json:"id"`
	RequestID string         `json:"request_id"`
	Status    string         `json:"status"`
	Message   string         `json:"message"`
	Data      *videoTaskData `json:"data"`
}

func (p *videoTaskSubmitResponse) taskID() string {
	return coalesceString(videoTaskID(p.ID), videoTaskID(p.RequestID))
}

func (p *videoTaskSubmitResponse) metadata() map[string]any {
	ret := make(map[string]any)
	if v := strings.TrimSpace(p.Status); v != "" {
		ret["status"] = v
	}
	if v := strings.TrimSpace(p.Message); v != "" {
		ret["message"] = v
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

func (p *videoTaskSubmitResponse) toGenerateVideosResponse() *genai.GenerateVideosResponse {
	if p == nil || p.Data == nil {
		return &genai.GenerateVideosResponse{}
	}
	return p.Data.toGenerateVideosResponse()
}

type videoTaskStatusResponse struct {
	ID        string         `json:"id"`
	RequestID string         `json:"request_id"`
	Model     string         `json:"model"`
	Status    string         `json:"status"`
	Message   string         `json:"message"`
	Data      *videoTaskData `json:"data"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

func (p *videoTaskStatusResponse) taskID() string {
	return coalesceString(videoTaskID(p.ID), videoTaskID(p.RequestID))
}

func (p *videoTaskStatusResponse) metadata() map[string]any {
	ret := make(map[string]any)
	if v := strings.TrimSpace(p.Model); v != "" {
		ret["model"] = v
	}
	if v := strings.TrimSpace(p.Status); v != "" {
		ret["status"] = v
	}
	if v := strings.TrimSpace(p.Message); v != "" {
		ret["message"] = v
	}
	if v := strings.TrimSpace(p.CreatedAt); v != "" {
		ret["created_at"] = v
	}
	if v := strings.TrimSpace(p.UpdatedAt); v != "" {
		ret["updated_at"] = v
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

func (p *videoTaskStatusResponse) toGenerateVideosResponse() *genai.GenerateVideosResponse {
	if p == nil || p.Data == nil {
		return &genai.GenerateVideosResponse{}
	}
	return p.Data.toGenerateVideosResponse()
}

type videoTaskData struct {
	Videos                  []videoOutput `json:"videos"`
	RAIMediaFilteredCount   int32         `json:"raiMediaFilteredCount"`
	RAIMediaFilteredReasons []string      `json:"raiMediaFilteredReasons"`
}

func (p *videoTaskData) toGenerateVideosResponse() *genai.GenerateVideosResponse {
	if p == nil {
		return &genai.GenerateVideosResponse{}
	}
	items := make([]*genai.GeneratedVideo, 0, len(p.Videos))
	for _, item := range p.Videos {
		videoURL := strings.TrimSpace(item.URL)
		if videoURL == "" {
			continue
		}
		mime := strings.TrimSpace(item.MIMEType)
		if mime == "" {
			mime = guessVideoMIME(videoURL)
		}
		items = append(items, &genai.GeneratedVideo{
			Video: &genai.Video{
				URI:      videoURL,
				MIMEType: mime,
			},
		})
	}
	return &genai.GenerateVideosResponse{
		GeneratedVideos:         items,
		RAIMediaFilteredCount:   p.RAIMediaFilteredCount,
		RAIMediaFilteredReasons: p.RAIMediaFilteredReasons,
	}
}

type videoOutput struct {
	URL      string `json:"url"`
	MIMEType string `json:"mimeType"`
}

