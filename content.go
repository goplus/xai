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

// -----------------------------------------------------------------------------

type TextBuilder interface {
	Text(string) TextBuilder
}

// -----------------------------------------------------------------------------

type ImageType string

const (
	ImageJPEG ImageType = "image/jpeg"
	ImagePNG  ImageType = "image/png"
	ImageGIF  ImageType = "image/gif"
	ImageWebP ImageType = "image/webp"
)

type ContentBuilder interface {
	Text(string) ContentBuilder
	ImageURL(string) ContentBuilder
	ImageBase64(mime ImageType, base64 []byte) ContentBuilder
}

// -----------------------------------------------------------------------------

type MessageBuilder interface {
	User(content ContentBuilder) MessageBuilder
	Assistant(content ContentBuilder) MessageBuilder
}

// -----------------------------------------------------------------------------

type Message interface {
}

// -----------------------------------------------------------------------------

type StreamMessage interface {
}

// -----------------------------------------------------------------------------

type ToolBuilder interface {
}

// -----------------------------------------------------------------------------
