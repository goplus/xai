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

package vidu

import (
	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/vidu/video"
)

// imageByBytes implements xai.Image with raw bytes.
type imageByBytes struct {
	mime xai.ImageType
	data []byte
}

func (p *imageByBytes) Type() xai.ImageType { return p.mime }
func (p *imageByBytes) Blob() xai.BlobData  { return xai.BlobFromRaw(p.data) }
func (p *imageByBytes) StgUri() string      { return "" }

// imageByURI implements xai.Image with only StgUri.
type imageByURI struct {
	mime   xai.ImageType
	stgURI string
}

func (p *imageByURI) Type() xai.ImageType { return p.mime }
func (p *imageByURI) Blob() xai.BlobData  { return nil }
func (p *imageByURI) StgUri() string      { return p.stgURI }

func newImageFromBytes(mime xai.ImageType, data []byte) xai.Image {
	return &imageByBytes{mime: mime, data: data}
}

func newImageFromURI(mime xai.ImageType, stgURI string) xai.Image {
	return &imageByURI{mime: mime, stgURI: stgURI}
}

func newVideoFromBytes(mime xai.VideoType, data []byte) xai.Video {
	return video.NewVideoFromBytes(mime, data)
}

func newVideoFromURI(mime xai.VideoType, stgURI string) xai.Video {
	return video.NewVideoFromURI(mime, stgURI)
}
