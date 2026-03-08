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
 * See the License for the specific language governing permissions and limitations under the License.
 */

package internal

// Video param value constants (single source of truth for spec/kling).
const (
	ModeStd       = "std"
	ModePro       = "pro"
	Seconds5      = "5"
	Seconds10     = "10"
	Size1920x1080 = "1920x1080"
	Size1080x1920 = "1080x1920"
	Size1280x720  = "1280x720"
	Size720x1280  = "720x1280"
	Size1080x1080 = "1080x1080"
	Size720x720   = "720x720"
	SoundOn       = "on"
	SoundOff      = "off"
)

// Image param value constants.
const (
	Resolution1K = "1K"
	Resolution2K = "2K"
	Resolution4K = "4K"
	AspectAuto   = "auto"
	Aspect16x9   = "16:9"
	Aspect9x16   = "9:16"
	Aspect1x1    = "1:1"
	Aspect4x3    = "4:3"
	Aspect3x4    = "3:4"
	Aspect3x2    = "3:2"
	Aspect2x3    = "2:3"
	Aspect21x9   = "21:9"
	// Image reference (kling-v1-5)
	ImageRefSubject = "subject"
	ImageRefFace    = "face"
)
