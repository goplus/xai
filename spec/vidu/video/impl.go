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

// videoByBytes implements xai.Video with raw bytes.
type videoByBytes struct {
	mime xai.VideoType
	data []byte
}

func (p *videoByBytes) Type() xai.VideoType { return p.mime }
func (p *videoByBytes) Blob() xai.BlobData  { return xai.BlobFromRaw(p.data) }
func (p *videoByBytes) StgUri() string      { return "" }

// videoByURI implements xai.Video with only StgUri.
type videoByURI struct {
	mime   xai.VideoType
	stgURI string
}

func (p *videoByURI) Type() xai.VideoType { return p.mime }
func (p *videoByURI) Blob() xai.BlobData  { return nil }
func (p *videoByURI) StgUri() string      { return p.stgURI }

// NewVideoFromBytes creates a Video from raw bytes.
func NewVideoFromBytes(mime xai.VideoType, data []byte) xai.Video {
	return &videoByBytes{mime: mime, data: data}
}

// NewVideoFromURI creates a Video from a storage URI.
func NewVideoFromURI(mime xai.VideoType, stgURI string) xai.Video {
	return &videoByURI{mime: mime, stgURI: stgURI}
}
