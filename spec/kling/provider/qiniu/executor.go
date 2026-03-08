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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"
)

// ErrTaskFailed is returned when a task fails.
var ErrTaskFailed = errors.New("qiniu: task failed")

// -----------------------------------------------------------------------------
// ImageExecutor
// -----------------------------------------------------------------------------

// ImageExecutor implements kling.ImageGenExecutor for Qiniu API.
type ImageExecutor struct {
	client *Client
}

// NewImageExecutor creates a new ImageExecutor.
func NewImageExecutor(client *Client) *ImageExecutor {
	return &ImageExecutor{client: client}
}

// Submit submits an image generation task and returns an async response.
func (e *ImageExecutor) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	klingParams, ok := params.(*kling.Params)
	if !ok {
		return nil, fmt.Errorf("qiniu: expected *kling.Params, got %T", params)
	}

	typedParams, err := kling.BuildImageParams(string(model), klingParams)
	if err != nil {
		return nil, err
	}

	req, err := BuildImageRequest(string(model), typedParams)
	if err != nil {
		return nil, err
	}

	respBody, err := e.client.Post(ctx, req.Endpoint, req.Body)
	if err != nil {
		return nil, err
	}

	if req.IsO1 {
		return e.handleO1CreateResponse(ctx, respBody)
	}
	return e.handleStandardCreateResponse(ctx, respBody)
}

func (e *ImageExecutor) handleStandardCreateResponse(ctx context.Context, respBody []byte) (xai.OperationResponse, error) {
	var resp ImageTaskCreateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("qiniu: failed to parse image task response: %w", err)
	}
	if resp.TaskID == "" {
		return nil, fmt.Errorf("qiniu: empty task_id in response")
	}

	return kling.NewAsyncOperationResponse(func(ctx context.Context) (xai.OperationResponse, error) {
		return e.getStandardTaskStatus(ctx, resp.TaskID)
	}, resp.TaskID), nil
}

func (e *ImageExecutor) handleO1CreateResponse(ctx context.Context, respBody []byte) (xai.OperationResponse, error) {
	var resp O1ImageCreateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("qiniu: failed to parse O1 image response: %w", err)
	}
	if resp.RequestID == "" {
		return nil, fmt.Errorf("qiniu: empty request_id in O1 response")
	}

	return kling.NewAsyncOperationResponse(func(ctx context.Context) (xai.OperationResponse, error) {
		return e.getO1TaskStatus(ctx, resp.RequestID)
	}, resp.RequestID), nil
}

// GetTaskStatus queries the status of an image task.
func (e *ImageExecutor) GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	if isO1TaskID(taskID) {
		return e.getO1TaskStatus(ctx, taskID)
	}
	return e.getStandardTaskStatus(ctx, taskID)
}

func (e *ImageExecutor) getStandardTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	endpoint := GetImageTaskStatusEndpoint(taskID, false)
	respBody, err := e.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var resp ImageTaskStatusResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("qiniu: failed to parse image task status: %w", err)
	}

	if resp.IsCompleted() {
		urls := resp.GetImageURLs()
		return &kling.SyncOperationResponse{R: kling.NewOutputImages(urls)}, nil
	}

	if resp.IsFailed() {
		msg := resp.StatusMessage
		if msg == "" {
			msg = "unknown error"
		}
		return nil, fmt.Errorf("%w: %s", ErrTaskFailed, msg)
	}

	return kling.NewAsyncOperationResponse(func(ctx context.Context) (xai.OperationResponse, error) {
		return e.getStandardTaskStatus(ctx, taskID)
	}, taskID), nil
}

func (e *ImageExecutor) getO1TaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	endpoint := GetImageTaskStatusEndpoint(taskID, true)
	respBody, err := e.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var resp O1ImageStatusResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("qiniu: failed to parse O1 image status: %w", err)
	}

	if resp.IsCompleted() {
		urls := resp.GetImageURLs()
		return &kling.SyncOperationResponse{R: kling.NewOutputImages(urls)}, nil
	}

	if resp.IsFailed() {
		return nil, fmt.Errorf("%w: O1 task failed", ErrTaskFailed)
	}

	return kling.NewAsyncOperationResponse(func(ctx context.Context) (xai.OperationResponse, error) {
		return e.getO1TaskStatus(ctx, taskID)
	}, taskID), nil
}

// isO1TaskID checks if the task ID is from O1 API (starts with "qimage-").
func isO1TaskID(taskID string) bool {
	return strings.HasPrefix(taskID, "qimage-")
}

// -----------------------------------------------------------------------------
// VideoExecutor
// -----------------------------------------------------------------------------

// VideoExecutor implements kling.VideoGenExecutor for Qiniu API.
type VideoExecutor struct {
	client *Client
}

// NewVideoExecutor creates a new VideoExecutor.
func NewVideoExecutor(client *Client) *VideoExecutor {
	return &VideoExecutor{client: client}
}

// Submit submits a video generation task and returns an async response.
func (e *VideoExecutor) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	klingParams, ok := params.(*kling.Params)
	if !ok {
		return nil, fmt.Errorf("qiniu: expected *kling.Params, got %T", params)
	}

	typedParams, err := kling.BuildVideoParams(string(model), klingParams)
	if err != nil {
		return nil, err
	}

	req, err := BuildVideoRequest(string(model), typedParams)
	if err != nil {
		return nil, err
	}

	respBody, err := e.client.Post(ctx, EndpointVideos, req.Body)
	if err != nil {
		return nil, err
	}

	var resp VideoTaskResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("qiniu: failed to parse video task response: %w", err)
	}
	if resp.ID == "" {
		return nil, fmt.Errorf("qiniu: empty id in video response")
	}

	if resp.IsCompleted() {
		urls := resp.GetVideoURLs()
		return &kling.SyncOperationResponse{R: kling.NewOutputVideos(urls)}, nil
	}

	return kling.NewAsyncOperationResponse(func(ctx context.Context) (xai.OperationResponse, error) {
		return e.GetTaskStatus(ctx, resp.ID)
	}, resp.ID), nil
}

// GetTaskStatus queries the status of a video task.
func (e *VideoExecutor) GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	endpoint := GetVideoTaskStatusEndpoint(taskID)
	respBody, err := e.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var resp VideoTaskResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("qiniu: failed to parse video task status: %w", err)
	}

	if resp.IsCompleted() {
		urls := resp.GetVideoURLs()
		return &kling.SyncOperationResponse{R: kling.NewOutputVideos(urls)}, nil
	}

	if resp.IsFailed() {
		msg := resp.GetErrorMessage()
		if msg == "" {
			msg = "unknown error"
		}
		return nil, fmt.Errorf("%w: %s", ErrTaskFailed, msg)
	}

	return kling.NewAsyncOperationResponse(func(ctx context.Context) (xai.OperationResponse, error) {
		return e.GetTaskStatus(ctx, taskID)
	}, taskID), nil
}

// -----------------------------------------------------------------------------
// Convenience functions
// -----------------------------------------------------------------------------

// NewExecutors creates both ImageExecutor and VideoExecutor from a single client.
func NewExecutors(client *Client) (*ImageExecutor, *VideoExecutor) {
	return NewImageExecutor(client), NewVideoExecutor(client)
}

// NewService creates a kling.Service with Qiniu executors.
func NewService(token string, opts ...ClientOption) *kling.Service {
	client := NewClient(token, opts...)
	imgExec, vidExec := NewExecutors(client)
	return kling.NewService(imgExec, vidExec)
}

// Register creates a kling.Service with Qiniu executors and registers it with xai.
func Register(token string, opts ...ClientOption) {
	svc := NewService(token, opts...)
	kling.Register(svc)
}
