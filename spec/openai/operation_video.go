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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/types"
)

const (
	videoEndpoint     = "videos"
	videoPollInterval = 5 * time.Second
	maxPromptLen      = 2500
)

var (
	enumVideoSeconds = &xai.StringEnum{
		Values: []string{"4", "8", "12"},
	}
	enumVideoSizeSora2 = &xai.StringEnum{
		Values: []string{"720x1280", "1280x720"},
	}
	enumVideoSizeSora2Pro = &xai.StringEnum{
		Values: []string{"1280x720", "1024x1792", "1792x1024"},
	}
	enumVideoSize = &xai.StringEnum{
		Values: []string{"720x1280", "1280x720", "1024x1792", "1792x1024"},
	}
	videoRestrictions = map[string]*xai.Restriction{
		"Prompt":  {Required: true},
		"Seconds": {Limit: enumVideoSeconds},
		"Size":    {Limit: enumVideoSize},
	}
	videoFields = []xai.Field{
		{Name: "Prompt", Kind: types.String},
		{Name: "InputReference", Kind: types.String},
		{Name: "Seconds", Kind: types.String},
		{Name: "Size", Kind: types.String},
		{Name: "RemixFromVideoID", Kind: types.String},
	}
)

type genVideo struct {
	model  string
	params *videoParams
}

type videoInputSchema struct{}

func (videoInputSchema) Fields() []xai.Field {
	return append([]xai.Field(nil), videoFields...)
}

func (videoInputSchema) Restrict(name string) *xai.Restriction {
	return videoRestrictions[name]
}

func (p *genVideo) InputSchema() xai.InputSchema {
	return videoInputSchema{}
}

func (p *genVideo) Params() xai.Params {
	if p.params == nil {
		p.params = &videoParams{}
	}
	return p.params
}

func (p *genVideo) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (xai.OperationResponse, error) {
	s, ok := svc.(*Service)
	if !ok {
		return nil, xai.ErrNotSupported
	}
	params := p.Params().(*videoParams)
	if err := params.validate(p.model); err != nil {
		return nil, err
	}
	if strings.TrimSpace(s.apiKey) == "" {
		return nil, fmt.Errorf("openai: api key is required for video operations")
	}

	baseURL := s.operationBaseURL(opts)
	var (
		task *videoTask
		err  error
	)
	if params.RemixFromVideoID != "" {
		if videoTaskID(params.RemixFromVideoID) == "" {
			return nil, fmt.Errorf("openai: RemixFromVideoID is required and must be a valid video task ID in remix mode")
		}
		body := map[string]any{"prompt": params.Prompt}
		task, err = s.postVideoTask(ctx, baseURL, remixVideoEndpoint(params.RemixFromVideoID), body)
	} else {
		body := map[string]any{
			"model":  p.model,
			"prompt": params.Prompt,
		}
		if params.InputReference != "" {
			body["input_reference"] = params.InputReference
		}
		if params.Seconds != "" {
			body["seconds"] = params.Seconds
		}
		if params.Size != "" {
			body["size"] = params.Size
		}
		task, err = s.postVideoTask(ctx, baseURL, videoEndpoint, body)
	}
	if err != nil {
		return nil, err
	}
	return newVideoResp(task, baseURL), nil
}

type videoParams struct {
	Prompt           string
	InputReference   string
	Seconds          string
	Size             string
	RemixFromVideoID string
}

func (p *videoParams) Set(name string, val any) xai.Params {
	switch name {
	case "Prompt":
		p.Prompt = valueToString(val)
	case "InputReference":
		p.InputReference = valueToInputReference(val)
	case "Seconds":
		p.Seconds = valueToString(val)
	case "Size":
		p.Size = valueToString(val)
	case "RemixFromVideoID":
		p.RemixFromVideoID = valueToString(val)
	}
	return p
}

func (p *videoParams) validate(model string) error {
	if p.Prompt == "" {
		return fmt.Errorf("openai: Prompt is required")
	}
	if len(p.Prompt) > maxPromptLen {
		return fmt.Errorf("openai: Prompt must not exceed %d characters", maxPromptLen)
	}
	if p.Seconds != "" {
		if err := videoRestrictions["Seconds"].ValidateString("Seconds", p.Seconds); err != nil {
			return err
		}
	}
	if p.Size != "" {
		enum := sizeEnumForModel(model)
		if err := (&xai.Restriction{Limit: enum}).ValidateString("Size", p.Size); err != nil {
			return err
		}
	}
	if p.RemixFromVideoID != "" {
		if strings.TrimSpace(p.RemixFromVideoID) == "" {
			return fmt.Errorf("openai: RemixFromVideoID is required in remix mode")
		}
		if p.InputReference != "" {
			return fmt.Errorf("openai: InputReference is not supported in remix mode")
		}
		if p.Seconds != "" {
			return fmt.Errorf("openai: Seconds is not supported in remix mode")
		}
		if p.Size != "" {
			return fmt.Errorf("openai: Size is not supported in remix mode")
		}
	}
	return nil
}

func sizeEnumForModel(model string) *xai.StringEnum {
	m := strings.ToLower(strings.TrimSpace(model))
	if strings.Contains(m, "sora-2-pro") {
		return enumVideoSizeSora2Pro
	}
	return enumVideoSizeSora2
}

func valueToString(val any) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v)
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func valueToInputReference(val any) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v)
	case xai.Image:
		if u := strings.TrimSpace(v.StgUri()); u != "" {
			return u
		}
		if b := v.Blob(); b != nil {
			return "data:" + string(v.Type()) + ";base64," + b.Base64()
		}
		return ""
	case xai.Video:
		if u := strings.TrimSpace(v.StgUri()); u != "" {
			return u
		}
		if b := v.Blob(); b != nil {
			return "data:" + string(v.Type()) + ";base64," + b.Base64()
		}
		return ""
	default:
		return valueToString(v)
	}
}

type videoResp struct {
	task     *videoTask
	baseURL  string
	sleepDur time.Duration
}

func newVideoResp(task *videoTask, baseURL string) *videoResp {
	return &videoResp{
		task:     task,
		baseURL:  normalizeAPIBaseURL(baseURL),
		sleepDur: videoPollInterval,
	}
}

func (p *videoResp) Done() bool {
	done, _ := classifyVideoStatus(p.task)
	return done
}

func (p *videoResp) Sleep() {
	if p.sleepDur > 0 {
		time.Sleep(p.sleepDur)
	}
}

func (p *videoResp) Retry(ctx context.Context, svc xai.Service) (xai.OperationResponse, error) {
	s, ok := svc.(*Service)
	if !ok {
		return nil, xai.ErrNotSupported
	}
	task, err := s.getVideoTask(ctx, p.baseURL, p.TaskID())
	if err != nil {
		return nil, err
	}
	return newVideoResp(task, p.baseURL), nil
}

func (p *videoResp) Results() xai.Results {
	return newVideoResults(p.task)
}

func (p *videoResp) TaskID() string {
	if p.task == nil {
		return ""
	}
	return strings.TrimSpace(p.task.ID)
}

func (p *videoResp) GetError() error {
	_, failed := classifyVideoStatus(p.task)
	if !failed {
		return nil
	}
	if p.task != nil && p.task.Error != nil {
		msg := strings.TrimSpace(p.task.Error.Message)
		if msg == "" {
			msg = "video generation failed"
		}
		if code := strings.TrimSpace(p.task.Error.Code); code != "" {
			return fmt.Errorf("openai: [%s] %s", code, msg)
		}
		return fmt.Errorf("openai: %s", msg)
	}
	status := ""
	if p.task != nil {
		status = strings.TrimSpace(p.task.Status)
	}
	if status == "" {
		status = "failed"
	}
	return fmt.Errorf("openai: video generation failed (status=%s)", status)
}

type videoResults struct {
	task  *videoTask
	items []*xai.OutputVideo
}

func newVideoResults(task *videoTask) *videoResults {
	ret := &videoResults{task: task}
	if task == nil || task.TaskResult == nil {
		return ret
	}
	for _, item := range task.TaskResult.Videos {
		rawURL := strings.TrimSpace(item.URL)
		if rawURL == "" {
			continue
		}
		ret.items = append(ret.items, &xai.OutputVideo{
			Video: &resultVideo{
				typ: guessOutputVideoType(rawURL),
				uri: rawURL,
			},
		})
	}
	return ret
}

func (p *videoResults) XGo_Attr(name string) any {
	if p.task == nil {
		return nil
	}
	switch name {
	case "ID":
		return p.task.ID
	case "Model":
		return p.task.Model
	case "Status":
		return p.task.Status
	case "CreatedAt":
		return p.task.CreatedAt
	case "UpdatedAt":
		return p.task.UpdatedAt
	case "CompletedAt":
		return p.task.CompletedAt
	case "Seconds":
		return p.task.Seconds
	case "Size":
		return p.task.Size
	case "RemixedFromVideoID":
		return p.task.RemixedFromVideoID
	case "Error":
		return p.task.Error
	case "TaskResult":
		return p.task.TaskResult
	default:
		return nil
	}
}

func (p *videoResults) Len() int {
	return len(p.items)
}

func (p *videoResults) At(i int) xai.Generated {
	n := len(p.items)
	if i < 0 || i >= n {
		panicIndex("videoResults.At", i, n)
	}
	return p.items[i]
}

type resultVideo struct {
	typ xai.VideoType
	uri string
}

func (p *resultVideo) Type() xai.VideoType {
	return p.typ
}

func (p *resultVideo) Blob() xai.BlobData {
	return nil
}

func (p *resultVideo) StgUri() string {
	return p.uri
}

type videoTask struct {
	ID                 string            `json:"id"`
	Object             string            `json:"object"`
	Model              string            `json:"model"`
	Status             string            `json:"status"`
	CreatedAt          int64             `json:"created_at"`
	UpdatedAt          int64             `json:"updated_at"`
	CompletedAt        int64             `json:"completed_at"`
	Seconds            string            `json:"seconds"`
	Size               string            `json:"size"`
	RemixedFromVideoID string            `json:"remixed_from_video_id"`
	Error              *videoTaskError   `json:"error"`
	TaskResult         *videoTaskResults `json:"task_result"`
}

type videoTaskError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type videoTaskResults struct {
	Videos []videoOutput `json:"videos"`
}

type videoOutput struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	Duration string `json:"duration"`
}

func classifyVideoStatus(task *videoTask) (done, failed bool) {
	if task == nil {
		return false, false
	}
	switch strings.ToLower(strings.TrimSpace(task.Status)) {
	case "completed", "succeeded", "success", "done":
		return true, false
	case "failed", "error", "cancelled", "canceled", "aborted":
		return true, true
	case "queued", "pending", "running", "in_progress", "processing":
		return false, false
	default:
		if task.TaskResult != nil && len(task.TaskResult.Videos) > 0 {
			return true, false
		}
		return false, false
	}
}

func isSoraModel(model xai.Model) bool {
	m := strings.ToLower(strings.TrimSpace(string(model)))
	return strings.HasPrefix(m, "sora-")
}

func (p *Service) operationBaseURL(opts xai.OptionBuilder) string {
	if override := strings.TrimSpace(baseURLOption(opts)); override != "" {
		return normalizeAPIBaseURL(override)
	}
	return normalizeAPIBaseURL(p.baseURL)
}

func (p *Service) postVideoTask(ctx context.Context, baseURL, endpoint string, body map[string]any) (*videoTask, error) {
	var ret videoTask
	if err := p.doVideoJSON(ctx, http.MethodPost, baseURL, endpoint, body, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

func (p *Service) getVideoTask(ctx context.Context, baseURL, taskID string) (*videoTask, error) {
	taskID = videoTaskID(taskID)
	if taskID == "" {
		return nil, fmt.Errorf("openai: empty video task id")
	}
	var ret videoTask
	if err := p.doVideoJSON(ctx, http.MethodGet, baseURL, videoEndpoint+"/"+url.PathEscape(taskID), nil, &ret); err != nil {
		return nil, err
	}
	if ret.ID == "" {
		ret.ID = taskID
	}
	return &ret, nil
}

func (p *Service) doVideoJSON(ctx context.Context, method, baseURL, endpoint string, body any, out any) error {
	baseURL = normalizeAPIBaseURL(baseURL)
	u := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(endpoint, "/")

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	var rd io.Reader
	if len(payload) > 0 {
		rd = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, rd)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if len(payload) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if key := strings.TrimSpace(p.apiKey); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}

	resp, err := p.httpDoer().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseVideoAPIError(resp.StatusCode, data)
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

func (p *Service) httpDoer() httpDoer {
	p.httpClientOnce.Do(func() {
		if p.httpClient == nil {
			p.httpClient = &http.Client{}
		}
	})
	return p.httpClient
}

type videoAPIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func parseVideoAPIError(statusCode int, body []byte) error {
	var apiErr videoAPIError
	_ = json.Unmarshal(body, &apiErr)

	code := strings.TrimSpace(apiErr.Code)
	msg := strings.TrimSpace(apiErr.Message)
	if apiErr.Error != nil {
		if code == "" {
			code = strings.TrimSpace(apiErr.Error.Code)
		}
		if msg == "" {
			msg = strings.TrimSpace(apiErr.Error.Message)
		}
	}
	if msg == "" {
		msg = strings.TrimSpace(string(body))
	}
	if msg == "" {
		msg = http.StatusText(statusCode)
	}
	if code == "" {
		return fmt.Errorf("openai: %s (status: %d)", msg, statusCode)
	}
	return fmt.Errorf("openai: [%s] %s (status: %d)", code, msg, statusCode)
}

func remixVideoEndpoint(videoID string) string {
	videoID = videoTaskID(videoID)
	if videoID == "" {
		return videoEndpoint + "//remix"
	}
	return videoEndpoint + "/" + url.PathEscape(videoID) + "/remix"
}

func videoTaskID(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	if parsed, err := url.Parse(name); err == nil && parsed.Host != "" && parsed.Path != "" {
		name = parsed.Path
	}
	name = strings.TrimSuffix(strings.TrimSpace(name), "/")
	if idx := strings.LastIndexByte(name, '/'); idx >= 0 {
		name = name[idx+1:]
	}
	return strings.TrimSpace(name)
}

func guessOutputVideoType(rawURL string) xai.VideoType {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return xai.VideoMP4
	}
	switch strings.ToLower(path.Ext(parsed.Path)) {
	case ".webm":
		return xai.VideoWebM
	default:
		return xai.VideoMP4
	}
}
