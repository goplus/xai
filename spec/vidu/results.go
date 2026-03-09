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

package vidu

import (
	"context"
	"time"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/vidu/video"
)

// NewOutputVideos creates xai.Results from video URLs. Used by Executor
// implementations when converting aiprovider response to xai types.
func NewOutputVideos(urls []string) xai.Results {
	return video.NewOutputVideos(urls)
}

// SyncOperationResponse is a synchronous xai.OperationResponse.
type SyncOperationResponse struct {
	R xai.Results
}

func (p *SyncOperationResponse) Done() bool { return true }
func (p *SyncOperationResponse) Sleep()     {}
func (p *SyncOperationResponse) Retry(ctx context.Context, svc xai.Service) (xai.OperationResponse, error) {
	return p, nil
}
func (p *SyncOperationResponse) Results() xai.Results { return p.R }
func (p *SyncOperationResponse) TaskID() string       { return "" }

// AsyncOperationResponse is an xai.OperationResponse for async task polling.
type AsyncOperationResponse struct {
	RetryFunc func(ctx context.Context) (xai.OperationResponse, error)
	SleepDur  time.Duration
	taskID    string
}

// NewAsyncOperationResponse creates an async response.
func NewAsyncOperationResponse(retryFunc func(ctx context.Context) (xai.OperationResponse, error), taskID string) *AsyncOperationResponse {
	return &AsyncOperationResponse{
		RetryFunc: retryFunc,
		SleepDur:  2 * time.Second,
		taskID:    taskID,
	}
}

func (p *AsyncOperationResponse) Done() bool { return false }

func (p *AsyncOperationResponse) Sleep() {
	if p.SleepDur > 0 {
		time.Sleep(p.SleepDur)
	}
}

func (p *AsyncOperationResponse) Retry(ctx context.Context, svc xai.Service) (xai.OperationResponse, error) {
	if p.RetryFunc == nil {
		return p, nil
	}
	return p.RetryFunc(ctx)
}

func (p *AsyncOperationResponse) Results() xai.Results { return nil }
func (p *AsyncOperationResponse) TaskID() string       { return p.taskID }
