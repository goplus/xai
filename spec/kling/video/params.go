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

import "github.com/goplus/xai/spec/kling/internal"

// ImageRefType for image_list: 普通参考图 | 首帧 | 尾帧
type ImageRefType string

const (
	ImageTypeRef        ImageRefType = ""            // 普通参考图（主体、场景、风格等）
	ImageTypeFirstFrame ImageRefType = "first_frame" // 首帧
	ImageTypeEndFrame   ImageRefType = "end_frame"   // 尾帧
)

// VideoReferType 参考视频类型
type VideoReferType string

const (
	VideoReferTypeFeature VideoReferType = "feature" // 特征参考视频
	VideoReferTypeBase    VideoReferType = "base"    // 待编辑视频
)

// KeepOriginalSound 是否保留视频原声
type KeepOriginalSound string

const (
	KeepOriginalSoundYes KeepOriginalSound = "yes" // 保留
	KeepOriginalSoundNo  KeepOriginalSound = "no"  // 不保留
)

// ImageInput is a single image reference for multi-ref video (O1).
type ImageInput struct {
	Image string       `json:"image"`
	Type  ImageRefType `json:"type,omitempty"`
}

// VideoRef is a video reference for multi-ref video (O1).
type VideoRef struct {
	VideoURL          string            `json:"video_url"`
	ReferType         VideoReferType    `json:"refer_type,omitempty"`
	KeepOriginalSound KeepOriginalSound `json:"keep_original_sound,omitempty"`
}

// VideoParams is the unified interface for all video model parameters.
type VideoParams interface {
	Model() string
	videoParams()
}

type videoParamsMarker struct{}

func (videoParamsMarker) videoParams() {}

// BaseVideoParams for kling-v2-1 and kling-v2-5-turbo.
// V2-1 is img2video only; V2-5-turbo supports text2video + img2video.
type BaseVideoParams struct {
	videoParamsMarker
	ModelName      string `json:"model"`
	Prompt         string `json:"prompt"`
	InputReference string `json:"input_reference,omitempty"`
	ImageTail      string `json:"image_tail,omitempty"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Mode           string `json:"mode,omitempty"`
	Seconds        string `json:"seconds,omitempty"`
	Size           string `json:"size,omitempty"`
}

func (p *BaseVideoParams) Model() string { return p.ModelName }

// V21VideoParams is an alias for backward compatibility.
type V21VideoParams = BaseVideoParams

// V25VideoParams is an alias for backward compatibility.
type V25VideoParams = BaseVideoParams

// O1VideoParams for kling-video-o1 (multi-ref).
// Uses image_list and video_list as primary; input_reference/image_tail are converted to image_list in build.
type O1VideoParams struct {
	videoParamsMarker
	Prompt         string       `json:"prompt"`
	ImageList      []ImageInput `json:"image_list,omitempty"`
	VideoList      []VideoRef   `json:"video_list,omitempty"`
	NegativePrompt string       `json:"negative_prompt,omitempty"`
	Mode           string       `json:"mode,omitempty"`
	Seconds        string       `json:"seconds,omitempty"`
	Size           string       `json:"size,omitempty"`
	VideoMode      string       `json:"video_mode,omitempty"`
}

func (p *O1VideoParams) Model() string { return internal.ModelKlingVideoO1 }

// V26VideoParams for kling-v2-6 through v2-9.
type V26VideoParams struct {
	videoParamsMarker
	ModelName            string            `json:"model"` // kling-v2-6, kling-v2-7, etc.
	Prompt               string            `json:"prompt"`
	InputReference       string            `json:"input_reference,omitempty"`
	ImageTail            string            `json:"image_tail,omitempty"`
	NegativePrompt       string            `json:"negative_prompt,omitempty"`
	Mode                 string            `json:"mode,omitempty"`
	Seconds              string            `json:"seconds,omitempty"`
	Size                 string            `json:"size,omitempty"`
	Sound                string            `json:"sound,omitempty"`
	ImageURL             string            `json:"image_url,omitempty"`
	VideoURL             string            `json:"video_url,omitempty"`
	CharacterOrientation string            `json:"character_orientation,omitempty"`
	KeepOriginalSound    KeepOriginalSound `json:"keep_original_sound,omitempty"`
}

func (p *V26VideoParams) Model() string { return p.ModelName }

// V3VideoParams for kling-v3.
type V3VideoParams struct {
	videoParamsMarker
	ModelName      string            `json:"model"`
	Prompt         string            `json:"prompt"`
	InputReference string            `json:"input_reference,omitempty"`
	MultiShot      bool              `json:"multi_shot,omitempty"`
	ShotType       string            `json:"shot_type,omitempty"`
	MultiPrompt    []MultiPromptItem `json:"multi_prompt,omitempty"`
	Sound          string            `json:"sound,omitempty"`
	Mode           string            `json:"mode,omitempty"`
	Seconds        string            `json:"seconds,omitempty"`
	Size           string            `json:"size,omitempty"`
}

func (p *V3VideoParams) Model() string { return p.ModelName }

// MultiPromptItem for multi_shot (V3/V3-omni).
type MultiPromptItem struct {
	Index    int    `json:"index"`
	Prompt   string `json:"prompt"`
	Duration string `json:"duration"`
}

// V3OmniVideoParams for kling-v3-omni.
type V3OmniVideoParams struct {
	videoParamsMarker
	ModelName   string            `json:"model"`
	Prompt      string            `json:"prompt"`
	MultiShot   bool              `json:"multi_shot,omitempty"`
	ShotType    string            `json:"shot_type,omitempty"`
	MultiPrompt []MultiPromptItem `json:"multi_prompt,omitempty"`
	ImageList   []ImageInput      `json:"image_list,omitempty"`
	VideoList   []VideoRef        `json:"video_list,omitempty"`
	Sound       string            `json:"sound,omitempty"`
	Mode        string            `json:"mode,omitempty"`
	Seconds     string            `json:"seconds,omitempty"`
	Size        string            `json:"size,omitempty"`
}

func (p *V3OmniVideoParams) Model() string { return p.ModelName }
