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

// Veo optional parameter constants. Use with Operation(GenVideo).Params().Set(...).
// All params except Prompt are optional; use these constants when setting to ensure valid values.
//
// AspectRatio (Veo 2/3): 16:9 or 9:16
const (
	Aspect16x9 = "16:9 (landscape)" // 横屏
	Aspect9x16 = "9:16 (portrait)"  // 竖屏
)

// Resolution (Veo 3 only): 720p or 1080p
const (
	Res720p  = "720p"
	Res1080p = "1080p"
)

// PersonGeneration: dont_allow or allow_adult
const (
	PersonDontAllow  = "dont_allow"
	PersonAllowAdult = "allow_adult"
)

// DurationSeconds: Veo 2 supports 5–8; Veo 3 supports 4, 6, 8
const (
	Duration4 int32 = 4
	Duration5 int32 = 5
	Duration6 int32 = 6
	Duration7 int32 = 7
	Duration8 int32 = 8
)

// CompressionQuality: LOSSLESS or OPTIMIZED (genai.VideoCompressionQuality)
const (
	CompressionLossless = "LOSSLESS"
	CompressionOptimized = "OPTIMIZED"
)
