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

package qiniu

import (
	"context"
	"net/url"
	"os"
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/openai"
)

// NewVideoService creates a video-only Service backed by Qiniu Sora API (api.qnaigc.com).
// It supports only xai.GenVideo operations for sora-* models; Gen and GenStream
// return xai.ErrNotSupported.
//
// Use this when you need only video generation without chat completions.
// The token can be passed directly or read from QINIU_API_KEY env when empty.
func NewVideoService(token string, opts ...ClientOption) *openai.Service {
	if token == "" {
		token = os.Getenv("QINIU_API_KEY")
	}
	cfg := &clientConfig{baseURL: DefaultBaseURL}
	for _, opt := range opts {
		opt(cfg)
	}
	uri := openai.SchemeV1 + ":base=" + cfg.baseURL + "&key=" + url.QueryEscape(token)
	svc, err := openai.NewVideoOnly(context.Background(), uri)
	if err != nil {
		panic("qiniu: " + err.Error())
	}
	return svc.(*openai.Service)
}

// RegisterVideo registers a video-only Qiniu service with xai under scheme "qiniu-video".
// After calling RegisterVideo(token), xai.New(ctx, "qiniu-video:") returns the video-only service.
func RegisterVideo(token string, opts ...ClientOption) {
	svc := NewVideoService(token, opts...)
	xai.Register("qiniu-video", func(ctx context.Context, uri string) (xai.Service, error) {
		if uri == "" {
			return svc, nil
		}
		params, err := url.ParseQuery(uri)
		if err != nil {
			return nil, err
		}
		tok := token
		if k := params["key"]; len(k) > 0 {
			tok = k[0]
		}
		if tok == "" {
			tok = os.Getenv("QINIU_API_KEY")
		}
		base := DefaultBaseURL
		if b := params["base"]; len(b) > 0 {
			base = strings.TrimSuffix(b[0], "/") + "/"
		}
		clientOpts := []ClientOption{}
		if base != DefaultBaseURL {
			clientOpts = append(clientOpts, WithBaseURL(base))
		}
		return NewVideoService(tok, clientOpts...), nil
	})
}
