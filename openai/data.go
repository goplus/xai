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
	"strings"
	"unsafe"

	"github.com/goplus/xai"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
)

// -----------------------------------------------------------------------------

type inputData struct {
	data responses.ResponseInputFileParam
	mime string
}

type base64Data string

func makeInputData(mime string, src any) (*inputData, error) {
	var b strings.Builder
	b.WriteString("data:")
	b.WriteString(mime)
	b.WriteString(";base64,")
	if data, ok := src.(base64Data); ok {
		b.WriteString(string(data))
	} else {
		encoder := base64.NewEncoder(base64.StdEncoding, &b)
		if data, ok := src.([]byte); ok {
			encoder.Write(data)
		} else {
			_, err := io.Copy(encoder, src.(io.Reader))
			if err != nil {
				return nil, err
			}
		}
		encoder.Close()
	}
	return &inputData{
		data: responses.ResponseInputFileParam{
			FileData: param.NewOpt(b.String()),
		},
		mime: mime,
	}, nil
}

func (p *inputData) ImageType() xai.ImageType {
	return xai.ImageType(p.mime)
}

func (p *inputData) DocumentType() xai.DocumentType {
	return xai.DocumentType(p.mime)
}

// -----------------------------------------------------------------------------

type imageBuilder struct {
}

func (p imageBuilder) From(mime xai.ImageType, displayName string, src io.Reader) (xai.ImageData, error) {
	return makeInputData(string(mime), src)
}

func (p imageBuilder) FromLocal(mime xai.ImageType, fileName string) (xai.ImageData, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return makeInputData(string(mime), f)
}

func (p imageBuilder) FromBytes(mime xai.ImageType, displayName string, data []byte) xai.ImageData {
	ret, _ := makeInputData(string(mime), data)
	return ret
}

func (p imageBuilder) FromBase64(mime xai.ImageType, displayName string, data string) (xai.ImageData, error) {
	return makeInputData(string(mime), base64Data(data))
}

func (p *Provider) Images() xai.ImageBuilder {
	return imageBuilder{}
}

// -----------------------------------------------------------------------------

type docBuilder struct {
}

func (p docBuilder) From(mime xai.DocumentType, displayName string, src io.Reader) (xai.DocumentData, error) {
	return makeInputData(string(mime), src)
}

func (p docBuilder) FromLocal(mime xai.DocumentType, fileName string) (xai.DocumentData, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return makeInputData(string(mime), f)
}

func (p docBuilder) FromBase64(mime xai.DocumentType, displayName string, data string) (xai.DocumentData, error) {
	return makeInputData(string(mime), base64Data(data))
}

func (p docBuilder) FromBytes(mime xai.DocumentType, displayName string, data []byte) xai.DocumentData {
	ret, _ := makeInputData(string(mime), data)
	return ret
}

func (p docBuilder) PlainText(text string) xai.DocumentData {
	data := unsafe.Slice(unsafe.StringData(text), len(text))
	ret, _ := makeInputData(string(xai.DocPlainText), data)
	return ret
}

func (p *Provider) Docs() xai.DocumentBuilder {
	return docBuilder{}
}

// -----------------------------------------------------------------------------
