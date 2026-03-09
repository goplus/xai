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

// imageByBytes implements xai.Image with raw bytes.
type imageByBytes struct {
	mime xai.ImageType
	data []byte
}

func (p *imageByBytes) Type() xai.ImageType    { return p.mime }
func (p *imageByBytes) Blob() xai.BlobData     { return xai.BlobFromRaw(p.data) }
func (p *imageByBytes) StgUri() string         { return "" }

// NewImageFromBytes creates xai.Image from raw bytes.
func NewImageFromBytes(mime xai.ImageType, data []byte) xai.Image {
	return &imageByBytes{mime: mime, data: data}
}

// NewImageFromURI creates xai.Image with only StgUri (URL).
func NewImageFromURI(mime xai.ImageType, stgUri string) xai.Image {
	return &imageByURI{mime: mime, stgUri: stgUri}
}
