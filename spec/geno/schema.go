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

package geno

import (
	"io"

	xai "github.com/goplus/xai/spec"
)

// -----------------------------------------------------------------------------

func (p *Service[T]) ImageFrom(mime xai.ImageType, src io.Reader) (xai.Image, error) {
	panic("todo")
}

func (p *Service[T]) ImageFromLocal(mime xai.ImageType, fileName string) (xai.Image, error) {
	panic("todo")
}

func (p *Service[T]) ImageFromBase64(mime xai.ImageType, base64 string) (xai.Image, error) {
	panic("todo")
}

func (p *Service[T]) ImageFromBytes(mime xai.ImageType, data []byte) xai.Image {
	panic("todo")
}

func (p *Service[T]) ImageFromStgUri(mime xai.ImageType, stgUri string) xai.Image {
	panic("todo")
}

// -----------------------------------------------------------------------------

func (p *Service[T]) VideoFrom(mime xai.VideoType, src io.Reader) (xai.Video, error) {
	panic("todo")
}
func (p *Service[T]) VideoFromLocal(mime xai.VideoType, fileName string) (xai.Video, error) {
	panic("todo")
}
func (p *Service[T]) VideoFromBase64(mime xai.VideoType, base64 string) (xai.Video, error) {
	panic("todo")
}
func (p *Service[T]) VideoFromBytes(mime xai.VideoType, data []byte) xai.Video {
	panic("todo")
}
func (p *Service[T]) VideoFromStgUri(mime xai.VideoType, stgUri string) xai.Video {
	panic("todo")
}

// -----------------------------------------------------------------------------

func (p *Service[T]) ReferenceImage(img xai.Image, id int32, typ xai.ReferenceImageType) (xai.ReferenceImage, xai.Configurable) {
	panic("todo")
}

func (p *Service[T]) GenVideoReferenceImages(imgs ...xai.GenVideoReferenceImage) xai.GenVideoReferenceImages {
	panic("todo")
}

func (p *Service[T]) GenVideoMask(img xai.Image, maskMode string) xai.GenVideoMask {
	panic("todo")
}

// -----------------------------------------------------------------------------
