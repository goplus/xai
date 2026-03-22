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
	"os"

	xai "github.com/goplus/xai/spec"
)

// -----------------------------------------------------------------------------

type Image struct {
	URI  string
	Data xai.BlobData
	MIME xai.ImageType
}

func (p *Image) Type() xai.ImageType {
	return p.MIME
}

func (p *Image) Blob() xai.BlobData {
	return p.Data
}

func (p *Image) StgUri() string {
	return p.URI
}

func (p *ServiceBase) ImageFrom(mime xai.ImageType, src io.Reader) (xai.Image, error) {
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return p.ImageFromBytes(mime, data), nil
}

func (p *ServiceBase) ImageFromLocal(mime xai.ImageType, fileName string) (xai.Image, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return p.ImageFromBytes(mime, data), nil
}

func (p *ServiceBase) ImageFromBase64(mime xai.ImageType, base64 string) (xai.Image, error) {
	return &Image{
		Data: xai.BlobFromBase64(base64),
		MIME: mime,
	}, nil
}

func (p *ServiceBase) ImageFromBytes(mime xai.ImageType, data []byte) xai.Image {
	return &Image{
		Data: xai.BlobFromRaw(data),
		MIME: mime,
	}
}

func (p *ServiceBase) ImageFromStgUri(mime xai.ImageType, stgUri string) xai.Image {
	return &Image{
		URI:  stgUri,
		MIME: mime,
	}
}

// -----------------------------------------------------------------------------

type Video struct {
	URI  string
	Data xai.BlobData
	MIME xai.VideoType
}

func (p *Video) Type() xai.VideoType {
	return p.MIME
}

func (p *Video) Blob() xai.BlobData {
	return p.Data
}

func (p *Video) StgUri() string {
	return p.URI
}

func (p *ServiceBase) VideoFrom(mime xai.VideoType, src io.Reader) (xai.Video, error) {
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return p.VideoFromBytes(mime, data), nil
}

func (p *ServiceBase) VideoFromLocal(mime xai.VideoType, fileName string) (xai.Video, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return p.VideoFromBytes(mime, data), nil
}

func (p *ServiceBase) VideoFromBase64(mime xai.VideoType, base64 string) (xai.Video, error) {
	return &Video{
		Data: xai.BlobFromBase64(base64),
		MIME: mime,
	}, nil
}

func (p *ServiceBase) VideoFromBytes(mime xai.VideoType, data []byte) xai.Video {
	return &Video{
		Data: xai.BlobFromRaw(data),
		MIME: mime,
	}
}

func (p *ServiceBase) VideoFromStgUri(mime xai.VideoType, stgUri string) xai.Video {
	return &Video{
		URI:  stgUri,
		MIME: mime,
	}
}

// -----------------------------------------------------------------------------

func (p *ServiceBase) ReferenceImage(img xai.Image, id int32, typ xai.ReferenceImageType) (xai.ReferenceImage, xai.Configurable) {
	panic("todo")
}

func (p *ServiceBase) GenVideoReferenceImages(imgs ...xai.GenVideoReferenceImage) xai.GenVideoReferenceImages {
	panic("todo")
}

func (p *ServiceBase) GenVideoMask(img xai.Image, maskMode string) xai.GenVideoMask {
	panic("todo")
}

// -----------------------------------------------------------------------------
