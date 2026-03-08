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
	"fmt"

	"github.com/goplus/xai/spec/kling"
	"github.com/goplus/xai/spec/kling/video"
)

// Video API endpoints.
const (
	EndpointVideos          = "/v1/videos"
	EndpointVideoTaskStatus = "/v1/videos/"
)

// VideoRequest holds the body for a video API request.
type VideoRequest struct {
	Body map[string]any
}

// BuildVideoRequest builds a VideoRequest from typed VideoParams.
func BuildVideoRequest(model string, params kling.VideoParams) (*VideoRequest, error) {
	switch p := params.(type) {
	case *video.BaseVideoParams:
		return buildBaseVideoRequest(p), nil
	case *video.O1VideoParams:
		return buildO1VideoRequest(p), nil
	case *video.V26VideoParams:
		return buildV26VideoRequest(p), nil
	case *video.V3VideoParams:
		return buildV3VideoRequest(p), nil
	case *video.V3OmniVideoParams:
		return buildV3OmniVideoRequest(p), nil
	default:
		return nil, fmt.Errorf("qiniu: unsupported video params type: %T", params)
	}
}

func buildBaseVideoRequest(p *video.BaseVideoParams) *VideoRequest {
	body := map[string]any{
		"model":  p.Model(),
		"prompt": p.Prompt,
	}
	setOptionalString(body, "input_reference", p.InputReference)
	setOptionalString(body, "image_tail", p.ImageTail)
	setOptionalString(body, "negative_prompt", p.NegativePrompt)
	setOptionalString(body, "mode", p.Mode)
	setOptionalString(body, "seconds", p.Seconds)
	setOptionalString(body, "size", p.Size)

	return &VideoRequest{Body: body}
}

func buildO1VideoRequest(p *video.O1VideoParams) *VideoRequest {
	body := map[string]any{
		"model":  p.Model(),
		"prompt": p.Prompt,
	}
	if len(p.ImageList) > 0 {
		imgList := make([]map[string]any, len(p.ImageList))
		for i, img := range p.ImageList {
			item := map[string]any{"image": img.Image}
			if img.Type != "" {
				item["type"] = img.Type
			}
			imgList[i] = item
		}
		body["image_list"] = imgList
	}
	if len(p.VideoList) > 0 {
		vidList := make([]map[string]any, len(p.VideoList))
		for i, vid := range p.VideoList {
			item := map[string]any{"video_url": vid.VideoURL}
			if vid.ReferType != "" {
				item["refer_type"] = vid.ReferType
			}
			if vid.KeepOriginalSound != "" {
				item["keep_original_sound"] = vid.KeepOriginalSound
			}
			vidList[i] = item
		}
		body["video_list"] = vidList
	}
	setOptionalString(body, "negative_prompt", p.NegativePrompt)
	setOptionalString(body, "mode", p.Mode)
	setOptionalString(body, "seconds", p.Seconds)
	setOptionalString(body, "size", p.Size)
	setOptionalString(body, "video_mode", p.VideoMode)

	return &VideoRequest{Body: body}
}

func buildV26VideoRequest(p *video.V26VideoParams) *VideoRequest {
	body := map[string]any{
		"model":  p.Model(),
		"prompt": p.Prompt,
	}
	setOptionalString(body, "input_reference", p.InputReference)
	setOptionalString(body, "image_tail", p.ImageTail)
	setOptionalString(body, "negative_prompt", p.NegativePrompt)
	setOptionalString(body, "mode", p.Mode)
	setOptionalString(body, "seconds", p.Seconds)
	setOptionalString(body, "size", p.Size)
	setOptionalString(body, "sound", p.Sound)
	setOptionalString(body, "image_url", p.ImageURL)
	setOptionalString(body, "video_url", p.VideoURL)
	setOptionalString(body, "character_orientation", p.CharacterOrientation)
	setOptionalString(body, "keep_original_sound", string(p.KeepOriginalSound))

	return &VideoRequest{Body: body}
}

func buildV3VideoRequest(p *video.V3VideoParams) *VideoRequest {
	body := map[string]any{
		"model":  p.Model(),
		"prompt": p.Prompt,
	}
	setOptionalBool(body, "multi_shot", p.MultiShot)
	setOptionalString(body, "shot_type", p.ShotType)
	if len(p.MultiPrompt) > 0 {
		mpList := make([]map[string]any, len(p.MultiPrompt))
		for i, mp := range p.MultiPrompt {
			mpList[i] = map[string]any{
				"index":    mp.Index,
				"prompt":   mp.Prompt,
				"duration": mp.Duration,
			}
		}
		body["multi_prompt"] = mpList
	}
	setOptionalString(body, "input_reference", p.InputReference)
	setOptionalString(body, "sound", p.Sound)
	setOptionalString(body, "mode", p.Mode)
	setOptionalString(body, "seconds", p.Seconds)
	setOptionalString(body, "size", p.Size)

	return &VideoRequest{Body: body}
}

func buildV3OmniVideoRequest(p *video.V3OmniVideoParams) *VideoRequest {
	body := map[string]any{
		"model": p.Model(),
	}
	setOptionalString(body, "prompt", p.Prompt)
	setOptionalBool(body, "multi_shot", p.MultiShot)
	setOptionalString(body, "shot_type", p.ShotType)

	if len(p.MultiPrompt) > 0 {
		mpList := make([]map[string]any, len(p.MultiPrompt))
		for i, mp := range p.MultiPrompt {
			mpList[i] = map[string]any{
				"index":    mp.Index,
				"prompt":   mp.Prompt,
				"duration": mp.Duration,
			}
		}
		body["multi_prompt"] = mpList
	}

	if len(p.ImageList) > 0 {
		imgList := make([]map[string]any, len(p.ImageList))
		for i, img := range p.ImageList {
			item := map[string]any{"image": img.Image}
			if img.Type != "" {
				item["type"] = img.Type
			}
			imgList[i] = item
		}
		body["image_list"] = imgList
	}

	if len(p.VideoList) > 0 {
		vidList := make([]map[string]any, len(p.VideoList))
		for i, vid := range p.VideoList {
			item := map[string]any{"video_url": vid.VideoURL}
			if vid.ReferType != "" {
				item["refer_type"] = vid.ReferType
			}
			if vid.KeepOriginalSound != "" {
				item["keep_original_sound"] = vid.KeepOriginalSound
			}
			vidList[i] = item
		}
		body["video_list"] = vidList
	}

	setOptionalString(body, "sound", p.Sound)
	setOptionalString(body, "mode", p.Mode)
	setOptionalString(body, "seconds", p.Seconds)
	setOptionalString(body, "size", p.Size)

	return &VideoRequest{Body: body}
}

// GetVideoTaskStatusEndpoint returns the endpoint for querying video task status.
func GetVideoTaskStatusEndpoint(taskID string) string {
	return EndpointVideoTaskStatus + taskID
}
