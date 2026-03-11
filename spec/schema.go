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

type ValueLimit interface {
	valueLimit()
}

type StringEnum struct {
	Values []string
}

func (*StringEnum) valueLimit() {}

// -----------------------------------------------------------------------------

type Field struct {
	Name string
	Kind types.Kind
}

// Restriction represents the restrictions on a parameter. It defines the limit of the
// parameter value, as well as the parameters that should or should not exist together
// with the parameter.
type Restriction struct {
	// Limit defines the limit of the parameter value. It can be nil if there is
	// no limit for the parameter.
	Limit ValueLimit

	// If NotAllowedIf is not nil, it indicates that the parameter is not allowed if
	// the parameters in NotAllowedIf exist together.
	NotAllowedIf []string

	// If OptionalIf is not nil, it indicates that the parameter is optional only if the
	// parameters in OptionalIf exist together. If OptionalIf is nil, the parameter is
	// either required or optional based on the Required field.
	OptionalIf []string

	// Required indicates whether the parameter is required.
	// If a parameter is required, it must be provided by the user.
	Required bool
}

// InputSchema represents the schema of `Params`.
type InputSchema interface {
	// Fields returns the list of fields defined in the schema.
	Fields() []Field

	// Restriction returns the `Restriction` for the parameter with the given name. It
	// returns nil if there is no restriction for the parameter.
	Restriction(name string) *Restriction
}

// Params represents the parameters that can be set.
type Params interface {
	// Set sets a parameter for the operation. You can call this method multiple
	// times to set multiple parameters.
	Set(name string, val any) Params
}

// Configurable represents an object that can be configured with parameters defined in
// an InputSchema.
type Configurable interface {
	// Schema returns the schema of the configurable object, which defines the
	// parameters that can be set for the object.
	Schema() InputSchema

	// Params returns a `Params` that can be used to set parameters for the object.
	Params() Params
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

// ReferenceImageType represents the type of a reference image, which defines how the
// reference image will be used.
type ReferenceImageType int

const (
	RawReferenceImage ReferenceImageType = iota
	MaskReferenceImage
	ControlReferenceImage
	StyleReferenceImage
	SubjectReferenceImage
	ContentReferenceImage
)

// ReferenceImage is an interface that represents a generic reference image.
type ReferenceImage any

// -----------------------------------------------------------------------------

// A reference image for video generation.
type GenVideoReferenceImage struct {
	// The reference image.
	Image Image

	// The type of the reference image, which defines how the reference
	// image will be used to generate the video.
	//
	// ReferenceType = "ASSET":
	// A reference image that provides assets to the generated video,
	// such as the scene, an object, a character, etc.
	//
	// ReferenceType = "STYLE":
	// A reference image that provides aesthetics including colors,
	// lighting, texture, etc., to be used as the style of the generated video,
	// such as 'anime', 'photography', 'origami', etc.
	ReferenceType string
}

type GenVideoReferenceImages any

// GenVideoMask is a reference image with a mask mode for video generation.
type GenVideoMask any

// -----------------------------------------------------------------------------

// Generated represents a generated image or video. It can be one of the following
// types:
//   - OutputVideo: represents a generated video, which is returned by GenVideo action.
//   - OutputImage: represents a generated image, which is returned by GenImage, EditImage,
//     RecontextImage, UpscaleImage actions.
//   - OutputImageMask: represents a generated image mask with detected entity labels,
//     which is returned by SegmentImage action.
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

	// Optional. Watermarked image, anti-hotlinking format.
	Watermarked Image

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

func (*OutputImage) generated() {}

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

func (*OutputImageMask) generated() {}

type OutputVideo struct {
	// The output video data.
	Video

	// Optional. Watermarked video, anti-hotlinking format.
	Watermarked Video
}

func (*OutputVideo) generated() {}

// -----------------------------------------------------------------------------

type objectFactory interface {
	ImageFrom(mime ImageType, src io.Reader) (Image, error)
	ImageFromLocal(mime ImageType, fileName string) (Image, error)
	ImageFromBase64(mime ImageType, base64 string) (Image, error)
	ImageFromBytes(mime ImageType, data []byte) Image
	ImageFromStgUri(mime ImageType, stgUri string) Image

	VideoFrom(mime VideoType, src io.Reader) (Video, error)
	VideoFromLocal(mime VideoType, fileName string) (Video, error)
	VideoFromBase64(mime VideoType, base64 string) (Video, error)
	VideoFromBytes(mime VideoType, data []byte) Video
	VideoFromStgUri(mime VideoType, stgUri string) Video

	ReferenceImage(img Image, id int32, typ ReferenceImageType) (ReferenceImage, Configurable)

	GenVideoReferenceImages(imgs ...GenVideoReferenceImage) GenVideoReferenceImages

	// GenVideoMask is a reference image with a mask mode for video generation.
	//
	// `img` is the image mask to use for generating videos.
	//
	// `maskMode` describes how the mask will be used. Inpainting masks must
	// match the aspect ratio of the input video. Outpainting masks can be
	// either 9:16 or 16:9.
	//
	// maskMode = "INSERT":
	// The image mask contains a masked rectangular region which is
	// applied on the first frame of the input video. The object described in
	// the prompt is inserted into this region and will appear in subsequent
	// frames.
	//
	// maskMode = "REMOVE":
	// The image mask is used to determine an object in the
	// first video frame to track. This object is removed from the video.
	//
	// maskMode = "REMOVE_STATIC":
	// The image mask is used to determine a region in the
	// video. Objects in this region will be removed.
	//
	// maskMode = "OUTPAINT":
	// The image mask contains a masked rectangular region where
	// the input video will go. The remaining area will be generated. Video
	// masks are not supported.
	GenVideoMask(img Image, maskMode string) GenVideoMask
}

// -----------------------------------------------------------------------------
