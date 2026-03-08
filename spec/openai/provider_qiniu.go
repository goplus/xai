/*
 * Copyright (c) 2026 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use it except in compliance with the License.
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
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"strings"

	xai "github.com/goplus/xai/spec"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// qiniuV1Provider implements provider with extended response parsing for
// APIs that return an "images" field (e.g. Qiniu gemini-2.5-flash-image).
type qiniuV1Provider struct {
	base    *v1Provider
	baseURL string
	apiKey  string
}

func newQiniuV1Provider(opts []option.RequestOption, baseURL, apiKey string) *qiniuV1Provider {
	return &qiniuV1Provider{
		base:    newV1Provider(opts),
		baseURL: strings.TrimSuffix(baseURL, "/") + "/",
		apiKey:  apiKey,
	}
}

func (p *qiniuV1Provider) Features() xai.Feature {
	return xai.FeatureGen | xai.FeatureGenStream
}

func (p *qiniuV1Provider) Gen(ctx context.Context, req *genRequest, opts []option.RequestOption) (genResponse, error) {
	params := p.base.buildParams(req)
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chat completions: %s: %s", resp.Status, string(respBody))
	}

	return parseChatCompletionResponseExtended(respBody)
}

func (p *qiniuV1Provider) GenStream(ctx context.Context, req *genRequest, opts []option.RequestOption) iter.Seq2[genResponse, error] {
	// Fall back to standard provider for streaming (images typically not in stream)
	return p.base.GenStream(ctx, req, opts)
}

// chatCompletionResponseRaw parses the extended response format with images.
type chatCompletionResponseRaw struct {
	ID      string `json:"id"`
	Choices []struct {
		Index        int32  `json:"index"`
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Role      string `json:"role"`
			Content   string `json:"content"`
			Images    []struct {
				Type     string `json:"type"`
				ImageURL struct {
					URL string `json:"url"`
				} `json:"image_url"`
			} `json:"images"`
			ToolCalls []openai.ChatCompletionMessageToolCallUnion `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
}

func parseChatCompletionResponseExtended(body []byte) (genResponse, error) {
	var raw chatCompletionResponseRaw
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if len(raw.Choices) == 0 {
		return nil, fmt.Errorf("chat completions: empty choices")
	}

	// Build openai.ChatCompletion for compatibility
	choices := make([]openai.ChatCompletionChoice, len(raw.Choices))
	imagesPerChoice := make([][]xai.Blob, len(raw.Choices))
	rawPerChoice := make([]*chatCompletionMessageRaw, len(raw.Choices))

	for i, c := range raw.Choices {
		choices[i] = openai.ChatCompletionChoice{
			Index:        int64(c.Index),
			FinishReason: c.FinishReason,
			Message: openai.ChatCompletionMessage{
				Content:   c.Message.Content,
				ToolCalls: c.Message.ToolCalls,
			},
		}
		// Parse images (data: URLs become blobs; plain URLs are skipped)
		var blobs []xai.Blob
		for _, img := range c.Message.Images {
			if url := img.ImageURL.URL; url != "" {
				b := blobFromImageURL(url)
				if b.BlobData != nil {
					blobs = append(blobs, b)
				}
			}
		}
		imagesPerChoice[i] = blobs
		if len(blobs) > 0 {
			rawPerChoice[i] = &chatCompletionMessageRaw{
				Content:   c.Message.Content,
				ToolCalls: c.Message.ToolCalls,
			}
			for j := range c.Message.Images {
				rawPerChoice[i].Images = append(rawPerChoice[i].Images, struct {
					Type     string `json:"type"`
					ImageURL struct {
						URL string `json:"url"`
					} `json:"image_url"`
				}{
					Type: c.Message.Images[j].Type,
					ImageURL: struct {
						URL string `json:"url"`
					}{URL: c.Message.Images[j].ImageURL.URL},
				})
			}
		}
	}

	return &v1ResponseWithImages{
		msg: &openai.ChatCompletion{
			Choices: choices,
		},
		images: imagesPerChoice,
		raw:    rawPerChoice,
	}, nil
}

func blobFromImageURL(rawURL string) xai.Blob {
	if !strings.HasPrefix(rawURL, "data:") {
		return xai.Blob{} // URL-only, skip (no blob data to display)
	}
	payload := strings.TrimPrefix(rawURL, "data:")
	pos := strings.Index(payload, ",")
	if pos < 0 {
		return xai.Blob{}
	}
	header := payload[:pos]
	mime := strings.TrimSpace(strings.Split(header, ";")[0])
	if mime == "" {
		mime = "image/png"
	}
	b64 := payload[pos+1:]
	buf, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return xai.Blob{}
	}
	return xai.Blob{
		MIME:     mime,
		BlobData: xai.BlobFromRaw(buf),
	}
}
