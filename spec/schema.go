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
	"encoding/base64"
	"io"

	"github.com/goplus/xai/types"
)

// -----------------------------------------------------------------------------

type ImageType string

const (
	ImageJPEG ImageType = "image/jpeg"
	ImagePNG  ImageType = "image/png"
	ImageGIF  ImageType = "image/gif"
	ImageWebP ImageType = "image/webp"
)

type VideoType string

const (
	VideoMP4  VideoType = "video/mp4"
	VideoWebM VideoType = "video/webm"
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

// -----------------------------------------------------------------------------

// BlobData represents the raw data of a blob, which can be an image or a document.
// It provides methods to retrieve the raw bytes or the base64-encoded string of the
// data.
type BlobData interface {
	Raw() ([]byte, error)
	Base64() string
}

type blobRaw struct {
	raw []byte
}

func (b blobRaw) Raw() ([]byte, error) {
	return b.raw, nil
}

func (b blobRaw) Base64() string {
	return base64.StdEncoding.EncodeToString(b.raw)
}

// BlobFromRaw creates a BlobData from raw bytes.
func BlobFromRaw(raw []byte) BlobData {
	return blobRaw{raw: raw}
}

type blobBase64 struct {
	base64 string
}

func (b blobBase64) Raw() ([]byte, error) {
	return base64.StdEncoding.DecodeString(b.base64)
}

func (b blobBase64) Base64() string {
	return b.base64
}

// BlobFromBase64 creates a BlobData from a base64-encoded string.
func BlobFromBase64(base64 string) BlobData {
	return blobBase64{base64: base64}
}

// -----------------------------------------------------------------------------

type Image interface {
	Type() ImageType // MIME type of the image, e.g. "image/jpeg"
	Blob() BlobData  // may return nil if the image is represented by a storage URI
	StgUri() string  // may return empty string if the image is represented by raw data
}

type Video interface {
	Type() VideoType // MIME type of the video, e.g. "video/mp4"
	Blob() BlobData  // may return nil if the video is represented by a storage URI
	StgUri() string  // may return empty string if the video is represented by raw data
}

// -----------------------------------------------------------------------------

type Generated interface {
	generated()
}

type SafetyAttributes struct {
	// List of RAI categories.
	Categories []string

	// List of scores of each categories.
	Scores []float32
}

type OutputImage struct {
	// The output image data.
	Image

	// Optional. Responsible AI filter reason if the image is filtered out of the
	// response.
	RAIFilteredReason string

	// Optional. Safety attributes of the image. Lists of RAI categories and their
	// scores of each content.
	SafetyAttributes *SafetyAttributes

	// Optional. The rewritten prompt used for the image generation if the prompt
	// enhancer is enabled.
	EnhancedPrompt string
}

// An entity representing the segmented area.
type EntityLabel struct {
	// Optional. The label of the segmented entity.
	Label string
	// Optional. The confidence score of the detected label.
	Score float32
}

type EntityLabels interface {
	// Len returns the number of detected entities.
	Len() int

	// At retrieves a detected entity by index.
	At(i int) EntityLabel
}

type OutputImageMask struct {
	// The generated image mask.
	Mask Image

	// The detected entities on the segmented area.
	Labels EntityLabels
}

type OutputVideo struct {
	// The output video data.
	Video
}

func (*OutputImage) generated()     {}
func (*OutputImageMask) generated() {}
func (*OutputVideo) generated()     {}

// -----------------------------------------------------------------------------

type objectFactory interface {
	ImageFrom(mime ImageType, src io.Reader) (Image, error)
	ImageFromLocal(mime ImageType, fileName string) (Image, error)
	ImageFromBase64(mime ImageType, base64 string) (Image, error)
	ImageFromBytes(mime ImageType, data []byte) Image
	ImageFromStgUri(mime ImageType, stgUri string) Image
}

// -----------------------------------------------------------------------------
