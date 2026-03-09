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
 * WITHOUT WARRANTIES OR CONDITIONS OF KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package shared

import (
	"context"
	"os"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"
	"github.com/goplus/xai/spec/kling/provider/qiniu"
)

// NewService creates a Kling Service. If QINIU_API_KEY is set, uses real Qnagic API;
// otherwise returns a mock service with placeholder URLs.
func NewService() (*kling.Service, error) {
	token := os.Getenv("QINIU_API_KEY")
	if token != "" {
		return qiniu.NewService(token), nil
	}
	imgExec := &mockImageExecutor{urls: []string{"https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png"}}
	vidExec := &mockVideoExecutor{urls: []string{"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"}}
	return kling.NewService(imgExec, vidExec), nil
}

type mockImageExecutor struct {
	urls []string
}

func (m *mockImageExecutor) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	_ = params
	return &kling.SyncOperationResponse{R: kling.NewOutputImages(m.urls)}, nil
}

func (m *mockImageExecutor) GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	_ = taskID
	return nil, xai.ErrNotSupported
}

type mockVideoExecutor struct {
	urls []string
}

func (m *mockVideoExecutor) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	_ = params
	return &kling.SyncOperationResponse{R: kling.NewOutputVideos(m.urls)}, nil
}

func (m *mockVideoExecutor) GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	_ = taskID
	return nil, xai.ErrNotSupported
}

// NewServiceForModels creates a minimal Service for listing models/actions (no API calls).
// Use when QINIU_API_KEY is not set or for models-only demos.
func NewServiceForModels() *kling.Service {
	return kling.NewService(nil, nil)
}
