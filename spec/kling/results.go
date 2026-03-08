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

package kling

import (
	"context"
	"time"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling/image"
	"github.com/goplus/xai/spec/kling/video"
)

// NewOutputImages creates xai.Results from image URLs. Used by Executor
// implementations when converting aiprovider response to xai types.
func NewOutputImages(urls []string) xai.Results {
	return image.NewOutputImages(urls)
}

// NewOutputVideos creates xai.Results from video URLs. Used by Executor
// implementations when converting aiprovider response to xai types.
func NewOutputVideos(urls []string) xai.Results {
	return video.NewOutputVideos(urls)
}

// SyncOperationResponse is a synchronous xai.OperationResponse (Done()==true).
// Used by Executor implementations for sync responses.
type SyncOperationResponse struct {
	R xai.Results
}

func (p *SyncOperationResponse) Done() bool   { return true }
func (p *SyncOperationResponse) Sleep()       {}
func (p *SyncOperationResponse) Retry(ctx context.Context, svc xai.Service) (xai.OperationResponse, error) {
	return p, nil
}
func (p *SyncOperationResponse) Results() xai.Results { return p.R }
func (p *SyncOperationResponse) TaskID() string        { return "" }

// -----------------------------------------------------------------------------
// AsyncOperationResponse for async task polling
// -----------------------------------------------------------------------------

// AsyncOperationResponse is an xai.OperationResponse for async tasks. Done() is false
// until the task completes. Retry() calls the retryFunc to poll status and returns
// a new OperationResponse (Sync when done, or another Async for further retries).
type AsyncOperationResponse struct {
	RetryFunc func(ctx context.Context) (xai.OperationResponse, error)
	SleepDur  time.Duration
	taskID    string
}

// NewAsyncOperationResponse creates an async response. retryFunc should call
// the backend's GetTaskStatus and return SyncOperationResponse when done, or
// a new AsyncOperationResponse when still processing.
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
func (p *AsyncOperationResponse) TaskID() string      { return p.taskID }
