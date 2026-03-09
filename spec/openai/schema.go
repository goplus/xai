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
	"encoding/base64"
	"io"
	"os"

	xai "github.com/goplus/xai/spec"
)

// -----------------------------------------------------------------------------

type image struct {
	mime xai.ImageType
	blob xai.BlobData
	uri  string
}

func (p *image) Type() xai.ImageType { return p.mime }
func (p *image) Blob() xai.BlobData  { return p.blob }
func (p *image) StgUri() string      { return p.uri }

type video struct {
	mime xai.VideoType
	blob xai.BlobData
	uri  string
}

func (p *video) Type() xai.VideoType { return p.mime }
func (p *video) Blob() xai.BlobData  { return p.blob }
func (p *video) StgUri() string      { return p.uri }

func (p *Service) ImageFrom(mime xai.ImageType, src io.Reader) (xai.Image, error) {
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return p.ImageFromBytes(mime, data), nil
}

func (p *Service) ImageFromLocal(mime xai.ImageType, fileName string) (xai.Image, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return p.ImageFromBytes(mime, data), nil
}

func (p *Service) ImageFromStgUri(mime xai.ImageType, stgUri string) xai.Image {
	return &image{
		mime: mime,
		uri:  stgUri,
	}
}

func (p *Service) ImageFromBytes(mime xai.ImageType, data []byte) xai.Image {
	return &image{
		mime: mime,
		blob: xai.BlobFromRaw(data),
	}
}

func (p *Service) ImageFromBase64(mime xai.ImageType, data string) (xai.Image, error) {
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return p.ImageFromBytes(mime, raw), nil
}

// -----------------------------------------------------------------------------

func (p *Service) VideoFrom(mime xai.VideoType, src io.Reader) (xai.Video, error) {
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return p.VideoFromBytes(mime, data), nil
}

func (p *Service) VideoFromLocal(mime xai.VideoType, fileName string) (xai.Video, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return p.VideoFromBytes(mime, data), nil
}

func (p *Service) VideoFromStgUri(mime xai.VideoType, stgUri string) xai.Video {
	return &video{
		mime: mime,
		uri:  stgUri,
	}
}

func (p *Service) VideoFromBytes(mime xai.VideoType, data []byte) xai.Video {
	return &video{
		mime: mime,
		blob: xai.BlobFromRaw(data),
	}
}

func (p *Service) VideoFromBase64(mime xai.VideoType, data string) (xai.Video, error) {
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return p.VideoFromBytes(mime, raw), nil
}

// -----------------------------------------------------------------------------

func (p *Service) ReferenceImage(img xai.Image, id int32, typ xai.ReferenceImageType) (xai.ReferenceImage, xai.Configurable) {
	panic("unsupported")
}

func (p *Service) GenVideoReferenceImages(imgs ...xai.GenVideoReferenceImage) xai.GenVideoReferenceImages {
	panic("unsupported")
}

func (p *Service) GenVideoMask(img xai.Image, maskMode string) xai.GenVideoMask {
	panic("unsupported")
}

// -----------------------------------------------------------------------------
