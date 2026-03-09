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
 * WITHOUT WARRANTIES OR CONDITIONS OF KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package image

import (
	xai "github.com/goplus/xai/spec"
)

// imageByURI implements xai.Image with only StgUri (URL from API response).
type imageByURI struct {
	mime   xai.ImageType
	stgUri string
}

func (p *imageByURI) Type() xai.ImageType     { return p.mime }
func (p *imageByURI) Blob() xai.BlobData      { return nil }
func (p *imageByURI) StgUri() string          { return p.stgUri }

// outputImages implements xai.Results for image generation.
type outputImages struct {
	urls []string
}

func (p *outputImages) XGo_Attr(name string) any { return nil }
func (p *outputImages) Len() int                  { return len(p.urls) }
func (p *outputImages) At(i int) xai.Generated {
	return &xai.OutputImage{
		Image: &imageByURI{mime: xai.ImagePNG, stgUri: p.urls[i]},
	}
}

// NewOutputImages creates xai.Results from image URLs. Used by Executor
// implementations when converting aiprovider response to xai types.
func NewOutputImages(urls []string) xai.Results {
	return &outputImages{urls: urls}
}
