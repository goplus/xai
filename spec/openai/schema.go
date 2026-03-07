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

package openai

import (
	"io"

	xai "github.com/goplus/xai/spec"
)

// -----------------------------------------------------------------------------

func (p *Service) ImageFrom(mime xai.ImageType, src io.Reader) (xai.Image, error) {
	panic("unsupported")
}

func (p *Service) ImageFromLocal(mime xai.ImageType, fileName string) (xai.Image, error) {
	panic("unsupported")
}

func (p *Service) ImageFromStgUri(mime xai.ImageType, stgUri string) xai.Image {
	panic("unsupported")
}

func (p *Service) ImageFromBytes(mime xai.ImageType, data []byte) xai.Image {
	panic("unsupported")
}

func (p *Service) ImageFromBase64(mime xai.ImageType, data string) (xai.Image, error) {
	panic("unsupported")
}

// -----------------------------------------------------------------------------

func (p *Service) VideoFrom(mime xai.VideoType, src io.Reader) (xai.Video, error) {
	panic("unsupported")
}

func (p *Service) VideoFromLocal(mime xai.VideoType, fileName string) (xai.Video, error) {
	panic("unsupported")
}

func (p *Service) VideoFromStgUri(mime xai.VideoType, stgUri string) xai.Video {
	panic("unsupported")
}

func (p *Service) VideoFromBytes(mime xai.VideoType, data []byte) xai.Video {
	panic("unsupported")
}

func (p *Service) VideoFromBase64(mime xai.VideoType, data string) (xai.Video, error) {
	panic("unsupported")
}

// -----------------------------------------------------------------------------

func (p *Service) ReferenceImage(img xai.Image, id int32, typ xai.ReferenceImageType) (xai.ReferenceImage, xai.Configurable) {
	panic("unsupported")
}

func (p *Service) GenVideoReferenceImages(imgs ...xai.GenVideoReferenceImage) xai.GenVideoReferenceImages {
	panic("unsupported")
}

// -----------------------------------------------------------------------------
