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
	xai "github.com/goplus/xai/spec"
)

// outputVideos implements xai.Results for video generation.
type outputVideos struct {
	urls []string
}

func (p *outputVideos) XGo_Attr(name string) any { return nil }
func (p *outputVideos) Len() int                 { return len(p.urls) }
func (p *outputVideos) At(i int) xai.Generated {
	return &xai.OutputVideo{
		Video: &videoByURI{mime: xai.VideoMP4, stgURI: p.urls[i]},
	}
}

// NewOutputVideos creates xai.Results from video URLs. Used by Executor
// implementations when converting aiprovider response to xai types.
func NewOutputVideos(urls []string) xai.Results {
	return &outputVideos{urls: urls}
}
