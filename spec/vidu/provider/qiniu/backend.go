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
	"github.com/goplus/xai/spec/vidu"
)

// ErrTaskFailed is returned when a task fails.
var ErrTaskFailed = errors.New("qiniu: task failed")

// -----------------------------------------------------------------------------
// Vidu API endpoints
// -----------------------------------------------------------------------------

const (
	EndpointQ1TextToVideo      = "/queue/fal-ai/vidu/q1/text-to-video"
	EndpointQ1ReferenceToVideo = "/queue/fal-ai/vidu/q1/reference-to-video"

	EndpointQ2TextToVideo      = "/queue/fal-ai/vidu/q2/text-to-video"
	EndpointQ2ReferenceToVideo = "/queue/fal-ai/vidu/q2/reference-to-video"
	EndpointQ2ImageToVideoPro  = "/queue/fal-ai/vidu/q2/image-to-video/pro"
	EndpointQ2StartEndToVideo  = "/queue/fal-ai/vidu/q2/start-end-to-video/pro"

	EndpointTaskStatus = "/queue/fal-ai/vidu/requests/"
)

// Task status constants.
const (
	StatusInQueue    = "IN_QUEUE"
	StatusInProgress = "IN_PROGRESS"
	StatusProcessing = "PROCESSING"
	StatusRunning    = "RUNNING"
	StatusQueued     = "QUEUED"
	StatusCompleted  = "COMPLETED"
	StatusFailed     = "FAILED"
	StatusCancelled  = "CANCELLED"
	StatusCanceled   = "CANCELED"
	StatusError      = "ERROR"
)

// -----------------------------------------------------------------------------
// Backend implementation
// -----------------------------------------------------------------------------

type backend struct {
	client *Client
}

func newBackend(client *Client) *backend {
	return &backend{client: client}
}

// NewBackend creates a vidu.Backend for the Qiniu API. Useful for testing or custom wiring.
func NewBackend(client *Client) vidu.Backend {
	return newBackend(client)
}

func (b *backend) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	viduParams, ok := params.(*vidu.Params)
	if !ok {
		return nil, fmt.Errorf("qiniu: expected *vidu.Params, got %T", params)
	}

	typedParams, err := vidu.BuildVideoParams(string(model), viduParams)
	if err != nil {
		return nil, err
	}

	req, err := BuildVideoRequest(string(model), typedParams)
	if err != nil {
		return nil, err
	}

	respBody, err := b.client.Post(ctx, req.Endpoint, req.Body)
	if err != nil {
		return nil, err
	}

	var resp VideoCreateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("qiniu: failed to parse video create response: %w", err)
	}
	if resp.RequestID == "" {
		return nil, fmt.Errorf("qiniu: empty request_id in response")
	}

	return b.newPollingResponse(resp.RequestID), nil
}

func (b *backend) GetTaskStatus(ctx context.Context, requestID string) (xai.OperationResponse, error) {
	endpoint := GetVideoTaskStatusEndpoint(requestID)
	respBody, err := b.client.Get(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var resp VideoStatusResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("qiniu: failed to parse video status response: %w", err)
	}

	if resp.IsCompleted() {
		urls := resp.GetVideoURLs()
		if len(urls) == 0 {
			return nil, fmt.Errorf("%w: completed but no video url", ErrTaskFailed)
		}
		return &vidu.SyncOperationResponse{R: vidu.NewOutputVideos(urls)}, nil
	}

	if resp.IsFailed() {
		msg := resp.GetErrorMessage()
		if msg == "" {
			msg = "unknown error"
		}
		return nil, fmt.Errorf("%w: %s", ErrTaskFailed, msg)
	}

	return b.newPollingResponse(requestID), nil
}

func (b *backend) newPollingResponse(requestID string) xai.OperationResponse {
	return vidu.NewAsyncOperationResponse(func(ctx context.Context) (xai.OperationResponse, error) {
		return b.GetTaskStatus(ctx, requestID)
	}, requestID)
}

// -----------------------------------------------------------------------------
// Video request building
// -----------------------------------------------------------------------------

type videoRequest struct {
	Endpoint string
	Body     map[string]any
}

type endpointKey struct {
	model string
	route vidu.GenerationRoute
}

var endpointByRoute = map[endpointKey]string{
	{model: vidu.ModelViduQ1, route: vidu.RouteTextToVideo}:      EndpointQ1TextToVideo,
	{model: vidu.ModelViduQ1, route: vidu.RouteReferenceToVideo}: EndpointQ1ReferenceToVideo,
	{model: vidu.ModelViduQ2, route: vidu.RouteTextToVideo}:      EndpointQ2TextToVideo,
	{model: vidu.ModelViduQ2, route: vidu.RouteReferenceToVideo}: EndpointQ2ReferenceToVideo,
	{model: vidu.ModelViduQ2, route: vidu.RouteImageToVideo}:     EndpointQ2ImageToVideoPro,
	{model: vidu.ModelViduQ2, route: vidu.RouteStartEndToVideo}:  EndpointQ2StartEndToVideo,
}

type routeBodyPatcher func(dst map[string]any, params *vidu.VideoParams)

var routeBodyPatchers = map[vidu.GenerationRoute]routeBodyPatcher{
	vidu.RouteReferenceToVideo: patchReferenceRouteBody,
	vidu.RouteImageToVideo:     patchImageRouteBody,
	vidu.RouteStartEndToVideo:  patchStartEndRouteBody,
}

// BuildVideoRequest builds an API request from typed Vidu params. Exported for tests.
func BuildVideoRequest(model string, params *vidu.VideoParams) (*VideoRequest, error) {
	if params == nil {
		return nil, fmt.Errorf("qiniu: video params is nil")
	}
	model = strings.ToLower(strings.TrimSpace(model))
	if model == "" {
		model = params.ModelName
	}

	route := params.Route()
	endpoint, err := SelectEndpoint(model, route)
	if err != nil {
		return nil, err
	}

	body := buildCommonVideoBody(params)
	if patcher, ok := routeBodyPatchers[route]; ok {
		patcher(body, params)
	}

	return &VideoRequest{Endpoint: endpoint, Body: body}, nil
}

// VideoRequest holds endpoint and body for a video API request.
type VideoRequest struct {
	Endpoint string
	Body     map[string]any
}

func buildCommonVideoBody(params *vidu.VideoParams) map[string]any {
	body := map[string]any{"prompt": params.Prompt}
	setOptionalInt(body, "seed", params.Seed)
	setOptionalInt(body, "duration", params.Duration)
	setOptionalString(body, "resolution", params.Resolution)
	setOptionalString(body, "movement_amplitude", params.MovementAmplitude)
	setOptionalBool(body, "watermark", params.Watermark)
	return body
}

func patchReferenceRouteBody(dst map[string]any, params *vidu.VideoParams) {
	if len(params.ReferenceImageURLs) > 0 {
		dst["reference_image_urls"] = params.ReferenceImageURLs
	}
	subjects := buildSubjectsPayload(params.Subjects)
	if len(subjects) > 0 {
		dst["subjects"] = subjects
	}
}

func patchImageRouteBody(dst map[string]any, params *vidu.VideoParams) {
	dst["image_url"] = params.ImageURL
}

func patchStartEndRouteBody(dst map[string]any, params *vidu.VideoParams) {
	dst["start_image_url"] = params.StartImageURL
	dst["end_image_url"] = params.EndImageURL
}

func buildSubjectsPayload(subjects []vidu.Subject) []map[string]any {
	items := make([]map[string]any, 0, len(subjects))
	for _, sb := range subjects {
		item := map[string]any{
			"id":     sb.ID,
			"images": sb.Images,
		}
		if sb.VoiceID != "" {
			item["voice_id"] = sb.VoiceID
		}
		items = append(items, item)
	}
	return items
}

// SelectEndpoint returns the API endpoint for the given model and route. Exported for tests.
func SelectEndpoint(model string, route vidu.GenerationRoute) (string, error) {
	model = strings.ToLower(strings.TrimSpace(model))
	if endpoint, ok := endpointByRoute[endpointKey{model: model, route: route}]; ok {
		return endpoint, nil
	}
	if !vidu.IsVideoModel(model) {
		return "", fmt.Errorf("qiniu: unsupported model %q", model)
	}
	return "", fmt.Errorf("qiniu: %s does not support route %s", model, route)
}

// GetVideoTaskStatusEndpoint returns the endpoint for querying task status.
func GetVideoTaskStatusEndpoint(requestID string) string {
	return EndpointTaskStatus + requestID + "/status"
}

func setOptionalString(dst map[string]any, key, value string) {
	if value != "" {
		dst[key] = value
	}
}

func setOptionalInt(dst map[string]any, key string, value *int) {
	if value != nil {
		dst[key] = *value
	}
}

func setOptionalBool(dst map[string]any, key string, value *bool) {
	if value != nil {
		dst[key] = *value
	}
}

// -----------------------------------------------------------------------------
// Response types
// -----------------------------------------------------------------------------

type statusSet map[string]struct{}

func newStatusSet(statuses ...string) statusSet {
	set := make(statusSet, len(statuses))
	for _, s := range statuses {
		set[normalizeStatus(s)] = struct{}{}
	}
	return set
}

func (s statusSet) Has(status string) bool {
	_, ok := s[normalizeStatus(status)]
	return ok
}

func normalizeStatus(status string) string {
	return strings.ToUpper(strings.TrimSpace(status))
}

var (
	completedStatuses = newStatusSet(StatusCompleted)
	failedStatuses   = newStatusSet(StatusFailed, StatusCancelled, StatusCanceled, StatusError)
	processingStatus = newStatusSet(StatusInQueue, StatusInProgress, StatusProcessing, StatusRunning, StatusQueued)
)

// VideoCreateResponse is the response from creating a video task.
type VideoCreateResponse struct {
	Status      string `json:"status"`
	RequestID   string `json:"request_id"`
	ResponseURL string `json:"response_url"`
	StatusURL   string `json:"status_url"`
	CancelURL   string `json:"cancel_url,omitempty"`
}

type videoMetrics struct {
	InferenceTime float64 `json:"inference_time,omitempty"`
}

// VideoItem represents a generated video.
type VideoItem struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type,omitempty"`
}

// VideoResult is the result field in status response.
type VideoResult struct {
	Video  *VideoItem  `json:"video,omitempty"`
	Videos []VideoItem `json:"videos,omitempty"`
}

// VideoError represents a structured error in status response.
type VideoError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// VideoStatusResponse is the response from querying a video task status.
type VideoStatusResponse struct {
	Status      string        `json:"status"`
	RequestID   string        `json:"request_id"`
	ResponseURL string        `json:"response_url,omitempty"`
	StatusURL   string        `json:"status_url,omitempty"`
	CancelURL   string        `json:"cancel_url,omitempty"`
	Metrics     *videoMetrics `json:"metrics,omitempty"`
	Result      *VideoResult  `json:"result,omitempty"`

	Message string      `json:"message,omitempty"`
	Error   *VideoError `json:"error,omitempty"`
}

func (r *VideoStatusResponse) IsCompleted() bool {
	return completedStatuses.Has(r.Status)
}

func (r *VideoStatusResponse) IsFailed() bool {
	return failedStatuses.Has(r.Status)
}

func (r *VideoStatusResponse) IsProcessing() bool {
	return processingStatus.Has(r.Status)
}

func (r *VideoStatusResponse) GetVideoURLs() []string {
	if r.Result == nil {
		return nil
	}
	urls := make([]string, 0, 1+len(r.Result.Videos))
	if r.Result.Video != nil {
		urls = appendURL(urls, r.Result.Video.URL)
	}
	for _, v := range r.Result.Videos {
		urls = appendURL(urls, v.URL)
	}
	return urls
}

func appendURL(urls []string, raw string) []string {
	url := strings.TrimSpace(raw)
	if url == "" {
		return urls
	}
	return append(urls, url)
}

// GetErrorMessage returns an error message for failed task.
func (r *VideoStatusResponse) GetErrorMessage() string {
	if r.Error != nil {
		msg := strings.TrimSpace(r.Error.Message)
		if msg != "" {
			return msg
		}
	}
	msg := strings.TrimSpace(r.Message)
	if msg != "" {
		return msg
	}
	return ""
}
