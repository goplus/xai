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

package claude

import (
	"encoding/base64"
	"io"
	"os"
	"unsafe"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/goplus/xai"
)

// -----------------------------------------------------------------------------

type imageData anthropic.BetaBase64ImageSourceParam

func (p *imageData) ImageType() xai.ImageType {
	return xai.ImageType(p.MediaType)
}

type imageBuilder struct {
}

func (p imageBuilder) From(mime xai.ImageType, displayName string, src io.Reader) (xai.ImageData, error) {
	data, err := io.ReadAll(src) // TODO(xsw): optimize for large files
	if err != nil {
		return nil, err
	}
	return p.FromBytes(mime, displayName, data), nil
}

func (p imageBuilder) FromLocal(mime xai.ImageType, fileName string) (xai.ImageData, error) {
	data, err := os.ReadFile(fileName) // TODO(xsw): optimize for large files
	if err != nil {
		return nil, err
	}
	return p.FromBytes(mime, "", data), nil
}

func (p imageBuilder) FromBytes(mime xai.ImageType, displayName string, data []byte) xai.ImageData {
	return &imageData{
		Data:      base64.StdEncoding.EncodeToString(data),
		MediaType: anthropic.BetaBase64ImageSourceMediaType(mime),
	}
}

func (p imageBuilder) FromBase64(mime xai.ImageType, _ string, base64 string) (xai.ImageData, error) {
	return &imageData{
		Data:      base64,
		MediaType: anthropic.BetaBase64ImageSourceMediaType(mime),
	}, nil
}

func (p *Provider) Images() xai.ImageBuilder {
	return imageBuilder{}
}

// -----------------------------------------------------------------------------

type docData struct {
	data anthropic.BetaContentBlockParamUnion
	mime xai.DocumentType
}

func (p *docData) DocumentType() xai.DocumentType {
	return p.mime
}

type docBuilder struct {
}

func (p docBuilder) From(mime xai.DocumentType, displayName string, src io.Reader) (xai.DocumentData, error) {
	data, err := io.ReadAll(src) // TODO(xsw): optimize for large files
	if err != nil {
		return nil, err
	}
	return fromBytes(mime, data), nil
}

func (p docBuilder) FromLocal(mime xai.DocumentType, fileName string) (xai.DocumentData, error) {
	data, err := os.ReadFile(fileName) // TODO(xsw): optimize for large files
	if err != nil {
		return nil, err
	}
	return fromBytes(mime, data), nil
}

func (p docBuilder) FromBase64(mime xai.DocumentType, _ string, base64 string) (xai.DocumentData, error) {
	return fromBase64(mime, base64), nil
}

func (p docBuilder) FromBytes(mime xai.DocumentType, displayName string, data []byte) xai.DocumentData {
	return fromBytes(mime, data)
}

func (p docBuilder) PlainText(text string) xai.DocumentData {
	return fromText(text)
}

func fromBytes(mime xai.DocumentType, data []byte) xai.DocumentData {
	if mime == xai.DocPlainText {
		return fromText(unsafe.String(unsafe.SliceData(data), len(data)))
	}
	return fromBase64(mime, base64.StdEncoding.EncodeToString(data))
}

func fromBase64(mime xai.DocumentType, base64 string) xai.DocumentData {
	switch mime {
	case xai.DocPDF:
		return &docData{
			data: anthropic.NewBetaDocumentBlock(anthropic.BetaBase64PDFSourceParam{
				Data: base64,
			}),
			mime: xai.DocPDF,
		}
	default:
		panic("todo")
	}
}

func fromText(text string) xai.DocumentData {
	return &docData{
		data: anthropic.NewBetaDocumentBlock(anthropic.BetaPlainTextSourceParam{
			Data: text,
		}),
		mime: xai.DocPlainText,
	}
}

func (p *Provider) Docs() xai.DocumentBuilder {
	return docBuilder{}
}

// -----------------------------------------------------------------------------
