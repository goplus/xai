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

package gemini

import (
	"encoding/base64"
	"io"
	"os"
	"path/filepath"

	"github.com/goplus/xai"
	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

type imageData genai.Blob

func (p *imageData) ImageType() xai.ImageType {
	return xai.ImageType(p.MIMEType)
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
	return p.FromBytes(mime, filepath.Base(fileName), data), nil
}

func (p imageBuilder) FromBytes(mime xai.ImageType, displayName string, data []byte) xai.ImageData {
	return &imageData{
		Data:        data,
		DisplayName: displayName,
		MIMEType:    string(mime),
	}
}

func (p imageBuilder) FromBase64(mime xai.ImageType, displayName string, data string) (xai.ImageData, error) {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return &imageData{
		Data:        b,
		DisplayName: displayName,
		MIMEType:    string(mime),
	}, nil
}

func (p *Provider) Images() xai.ImageBuilder {
	return imageBuilder{}
}

// -----------------------------------------------------------------------------

type docData genai.Part

func (p *docData) DocumentType() xai.DocumentType {
	if p.InlineData != nil {
		return xai.DocumentType(p.InlineData.MIMEType)
	}
	return xai.DocPlainText
}

type docBuilder struct {
}

func (p docBuilder) From(mime xai.DocumentType, displayName string, src io.Reader) (xai.DocumentData, error) {
	data, err := io.ReadAll(src) // TODO(xsw): optimize for large files
	if err != nil {
		return nil, err
	}
	return p.FromBytes(mime, displayName, data), nil
}

func (p docBuilder) FromLocal(mime xai.DocumentType, fileName string) (xai.DocumentData, error) {
	data, err := os.ReadFile(fileName) // TODO(xsw): optimize for large files
	if err != nil {
		return nil, err
	}
	return p.FromBytes(mime, filepath.Base(fileName), data), nil
}

func (p docBuilder) FromBase64(mime xai.DocumentType, displayName string, data string) (xai.DocumentData, error) {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return p.FromBytes(mime, displayName, b), nil
}

func (p docBuilder) FromBytes(mime xai.DocumentType, displayName string, data []byte) xai.DocumentData {
	return &docData{
		InlineData: &genai.Blob{
			Data:        data,
			DisplayName: displayName,
			MIMEType:    string(mime),
		},
	}
}

func (p docBuilder) PlainText(text string) xai.DocumentData {
	return (*docData)(genai.NewPartFromText(text))
}

func (p *Provider) Docs() xai.DocumentBuilder {
	return docBuilder{}
}

// -----------------------------------------------------------------------------
