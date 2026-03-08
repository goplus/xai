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

// Task status constants.
const (
	StatusProcessing   = "processing"
	StatusCompleted    = "completed"
	StatusSucceeded    = "succeeded"
	StatusSucceed      = "succeed" // API returns "succeed" (not "succeeded")
	StatusFailed       = "failed"
	StatusInitializing = "initializing"
	StatusQueued       = "queued"
	StatusInProgress   = "in_progress"
	StatusDownloading  = "downloading"
	StatusUploading    = "uploading"
	StatusCancelled    = "cancelled"
	StatusInQueue      = "IN_QUEUE"
	StatusO1Completed  = "COMPLETED"
)

// -----------------------------------------------------------------------------
// Image API Responses
// -----------------------------------------------------------------------------

// ImageTaskCreateResponse is the response from creating an image generation task.
// Used by kling-v1, v1-5, v2, v2-new, v2-1.
type ImageTaskCreateResponse struct {
	TaskID string `json:"task_id"`
}

// ImageData represents a single generated image.
type ImageData struct {
	URL string `json:"url"`
}

// ImageTaskStatusResponse is the response from querying an image task status.
// Used by kling-v1, v1-5, v2, v2-new, v2-1.
type ImageTaskStatusResponse struct {
	TaskID        string      `json:"task_id"`
	Created       int64       `json:"created,omitempty"`
	Status        string      `json:"status"`
	StatusMessage string      `json:"status_message,omitempty"`
	Data          []ImageData `json:"data,omitempty"`
}

// IsCompleted returns true if the task is completed.
func (r *ImageTaskStatusResponse) IsCompleted() bool {
	return r.Status == StatusCompleted || r.Status == StatusSucceeded || r.Status == StatusSucceed
}

// IsFailed returns true if the task has failed.
func (r *ImageTaskStatusResponse) IsFailed() bool {
	return r.Status == StatusFailed
}

// IsProcessing returns true if the task is still processing.
func (r *ImageTaskStatusResponse) IsProcessing() bool {
	return r.Status == StatusProcessing
}

// GetImageURLs returns the URLs of generated images.
func (r *ImageTaskStatusResponse) GetImageURLs() []string {
	urls := make([]string, len(r.Data))
	for i, d := range r.Data {
		urls[i] = d.URL
	}
	return urls
}

// -----------------------------------------------------------------------------
// O1 Image API Responses
// -----------------------------------------------------------------------------

// O1ImageCreateResponse is the response from creating an O1 image task.
type O1ImageCreateResponse struct {
	Status      string `json:"status"`
	RequestID   string `json:"request_id"`
	ResponseURL string `json:"response_url"`
	StatusURL   string `json:"status_url"`
	CancelURL   string `json:"cancel_url,omitempty"`
}

// O1ImageItem represents a single generated image in O1 response.
type O1ImageItem struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type,omitempty"`
}

// O1ImageResult is the result field in O1 status response.
type O1ImageResult struct {
	Images []O1ImageItem `json:"images"`
}

// O1ImageMetrics contains timing information.
type O1ImageMetrics struct {
	InferenceTime float64 `json:"inference_time,omitempty"`
}

// O1ImageStatusResponse is the response from querying an O1 image task status.
type O1ImageStatusResponse struct {
	Status    string          `json:"status"`
	RequestID string          `json:"request_id"`
	Metrics   *O1ImageMetrics `json:"metrics,omitempty"`
	Result    *O1ImageResult  `json:"result,omitempty"`
}

// IsCompleted returns true if the O1 task is completed.
func (r *O1ImageStatusResponse) IsCompleted() bool {
	return r.Status == StatusO1Completed
}

// IsFailed returns true if the O1 task has failed.
func (r *O1ImageStatusResponse) IsFailed() bool {
	return r.Status == StatusFailed
}

// IsProcessing returns true if the O1 task is still processing.
func (r *O1ImageStatusResponse) IsProcessing() bool {
	return r.Status == StatusInQueue || r.Status == StatusInProgress
}

// GetImageURLs returns the URLs of generated images.
func (r *O1ImageStatusResponse) GetImageURLs() []string {
	if r.Result == nil {
		return nil
	}
	urls := make([]string, len(r.Result.Images))
	for i, img := range r.Result.Images {
		urls[i] = img.URL
	}
	return urls
}

// -----------------------------------------------------------------------------
// Video API Responses
// -----------------------------------------------------------------------------

// VideoItem represents a single generated video.
type VideoItem struct {
	ID       string `json:"id,omitempty"`
	URL      string `json:"url"`
	Duration string `json:"duration,omitempty"`
}

// VideoTaskResult contains the generated videos.
type VideoTaskResult struct {
	Videos []VideoItem `json:"videos"`
}

// VideoError represents an error in video generation.
type VideoError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// VideoTaskResponse is the response from creating or querying a video task.
type VideoTaskResponse struct {
	ID          string           `json:"id"`
	Object      string           `json:"object,omitempty"`
	Model       string           `json:"model,omitempty"`
	Mode        string           `json:"mode,omitempty"`
	Status      string           `json:"status"`
	CreatedAt   int64            `json:"created_at,omitempty"`
	UpdatedAt   int64            `json:"updated_at,omitempty"`
	CompletedAt int64            `json:"completed_at,omitempty"`
	Seconds     string           `json:"seconds,omitempty"`
	Size        string           `json:"size,omitempty"`
	TaskResult  *VideoTaskResult `json:"task_result,omitempty"`
	Error       *VideoError      `json:"error,omitempty"`
}

// IsCompleted returns true if the video task is completed.
func (r *VideoTaskResponse) IsCompleted() bool {
	return r.Status == StatusCompleted
}

// IsFailed returns true if the video task has failed.
func (r *VideoTaskResponse) IsFailed() bool {
	return r.Status == StatusFailed
}

// IsProcessing returns true if the video task is still processing.
func (r *VideoTaskResponse) IsProcessing() bool {
	switch r.Status {
	case StatusInitializing, StatusQueued, StatusInProgress, StatusDownloading, StatusUploading:
		return true
	default:
		return false
	}
}

// GetVideoURLs returns the URLs of generated videos.
func (r *VideoTaskResponse) GetVideoURLs() []string {
	if r.TaskResult == nil {
		return nil
	}
	urls := make([]string, len(r.TaskResult.Videos))
	for i, v := range r.TaskResult.Videos {
		urls[i] = v.URL
	}
	return urls
}

// GetErrorMessage returns the error message if the task failed.
func (r *VideoTaskResponse) GetErrorMessage() string {
	if r.Error != nil && r.Error.Message != "" {
		return r.Error.Message
	}
	return ""
}
