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
	"context"
	"iter"

	xai "github.com/goplus/xai/spec"
	"github.com/openai/openai-go/v3/option"
)

// videoOnlyProvider implements provider for services that only support video
// operations (e.g. Sora). Gen and GenStream return ErrNotSupported.
type videoOnlyProvider struct{}

func newVideoOnlyProvider() *videoOnlyProvider {
	return &videoOnlyProvider{}
}

func (p *videoOnlyProvider) Features() xai.Feature {
	return xai.FeatureOperation
}

func (p *videoOnlyProvider) Gen(ctx context.Context, req *genRequest, opts []option.RequestOption) (genResponse, error) {
	return nil, xai.ErrNotSupported
}

func (p *videoOnlyProvider) GenStream(ctx context.Context, req *genRequest, opts []option.RequestOption) iter.Seq2[genResponse, error] {
	return func(yield func(genResponse, error) bool) {
		yield(nil, xai.ErrNotSupported)
	}
}
