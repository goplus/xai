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

package xai

import (
	"io"

	"github.com/goplus/xai/spec/types"
)

// -----------------------------------------------------------------------------

type ImageType string

const (
	ImageJPEG ImageType = "image/jpeg"
	ImagePNG  ImageType = "image/png"
	ImageGIF  ImageType = "image/gif"
	ImageWebP ImageType = "image/webp"
)

type DocumentType string

const (
	DocPlainText DocumentType = "text/plain"
	DocPDF       DocumentType = "application/pdf"
)

// -----------------------------------------------------------------------------

type Field struct {
	Name string
	Kind types.Kind
}

type InputSchema interface {
	Fields() []Field
}

type Image interface {
	MIME() ImageType
}

// -----------------------------------------------------------------------------

type objectFactory interface {
	ImageFrom(mime ImageType, src io.Reader) (Image, error)
	ImageFromLocal(mime ImageType, fileName string) (Image, error)
	ImageFromBase64(mime ImageType, base64 string) (Image, error)
	ImageFromBytes(mime ImageType, data []byte) Image
	ImageFromStgUri(mime ImageType, stgUri string) Image
}

// -----------------------------------------------------------------------------
