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

package internal

// Param name constants (shared by kling, image, video packages).
const (
	ParamPrompt           = "prompt"
	ParamAspectRatio      = "aspect_ratio"
	ParamReferenceImages  = "reference_images"
	ParamImage            = "image"
	ParamImageReference   = "image_reference"
	ParamNegativePrompt   = "negative_prompt"
	ParamImageFidelity    = "image_fidelity"
	ParamHumanFidelity    = "human_fidelity"
	ParamN                = "n"
	ParamResolution       = "resolution"
	ParamInputReference   = "input_reference"
	ParamImageTail        = "image_tail"
	ParamMode             = "mode"
	ParamSeconds          = "seconds"
	ParamSize             = "size"
	ParamImageList        = "image_list"
	ParamVideoList        = "video_list"
	ParamSound            = "sound"
	ParamSubjectImageList = "subject_image_list"
	ParamSubjectImage     = "subject_image" // key in each subject_image_list item
	ParamSceneImage       = "scene_image"
	ParamStyleImage       = "style_image"
	// Motion control (V2.6)
	ParamImageUrl             = "image_url"
	ParamVideoUrl             = "video_url"
	ParamCharacterOrientation = "character_orientation"
	ParamKeepOriginalSound    = "keep_original_sound"
	// Multi-shot (V3-omni)
	ParamMultiShot   = "multi_shot"
	ParamShotType    = "shot_type"
	ParamMultiPrompt = "multi_prompt"
)
