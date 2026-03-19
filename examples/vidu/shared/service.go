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

package shared

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/vidu"
	"github.com/goplus/xai/spec/vidu/provider/qiniu"
)

// Service wraps xai.Service with optional SetApiKey for runtime configuration.
type Service struct {
	xai.Service
	setApiKey func(string)
}

// SetApiKey updates the API key at runtime. Only effective when using real Qiniu API;
// no-op for mock service.
func (s *Service) SetApiKey(apiKey string) {
	if s.setApiKey != nil {
		s.setApiKey(apiKey)
	}
}

// ViduService returns the embedded vidu.Service for operation.Call.
// Required so that *Service implements the serviceProvider interface
// used by vidu.genVideo.Call when delegating to the backend.
func (s *Service) ViduService() *vidu.Service {
	if sp, ok := s.Service.(interface{ ViduService() *vidu.Service }); ok {
		return sp.ViduService()
	}
	return nil
}

// GetTask passes through task resume support for wrapped services so examples can
// persist task IDs and continue polling through xai.GetTask.
func (s *Service) GetTask(ctx context.Context, model xai.Model, action xai.Action, taskID string) (xai.OperationResponse, error) {
	if sp, ok := s.Service.(interface {
		GetTask(context.Context, xai.Model, xai.Action, string) (xai.OperationResponse, error)
	}); ok {
		return sp.GetTask(ctx, model, action, taskID)
	}
	return nil, xai.ErrNotSupported
}

// NewService creates a Vidu Service.
// If QINIU_API_KEY is set, real Qiniu backend is used.
// Otherwise, a local async mock executor is used.
// The returned *Service supports SetApiKey(apiKey) for runtime API key updates.
func NewService() (*Service, error) {
	apiKey := strings.TrimSpace(os.Getenv("QINIU_API_KEY"))
	if apiKey != "" {
		qiniuSvc := qiniu.NewService(apiKey)
		return &Service{Service: qiniuSvc, setApiKey: qiniuSvc.SetApiKey}, nil
	}
	mockSvc := vidu.NewService(newMockVideoExecutor())
	return &Service{Service: mockSvc, setApiKey: nil}, nil
}

type mockVideoExecutor struct {
	mu       sync.Mutex
	seq      int64
	donePoll int
	tasks    map[string]*mockTask
}

type mockTask struct {
	polls int
	urls  []string
}

func newMockVideoExecutor() *mockVideoExecutor {
	return &mockVideoExecutor{
		donePoll: 1,
		tasks:    make(map[string]*mockTask),
	}
}

func (m *mockVideoExecutor) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	vparams, ok := params.(*vidu.Params)
	if !ok {
		return nil, fmt.Errorf("mock vidu: expected *vidu.Params, got %T", params)
	}
	typed, err := vidu.BuildVideoParams(string(model), vparams)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.seq++
	taskID := fmt.Sprintf("mock-vidu-task-%d", m.seq)
	url := fmt.Sprintf("https://example.com/mock/vidu/%s/%s/%s.mp4", typed.Model(), typed.Route(), taskID)
	m.tasks[taskID] = &mockTask{
		polls: 0,
		urls:  []string{url},
	}

	resp := vidu.NewAsyncOperationResponse(func(ctx context.Context) (xai.OperationResponse, error) {
		return m.GetTaskStatus(ctx, taskID)
	}, taskID)
	resp.SleepDur = 80 * time.Millisecond
	return resp, nil
}

func (m *mockVideoExecutor) GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("mock vidu: task not found: %s", taskID)
	}

	if task.polls < m.donePoll {
		task.polls++
		resp := vidu.NewAsyncOperationResponse(func(ctx context.Context) (xai.OperationResponse, error) {
			return m.GetTaskStatus(ctx, taskID)
		}, taskID)
		resp.SleepDur = 80 * time.Millisecond
		return resp, nil
	}

	return &vidu.SyncOperationResponse{R: vidu.NewOutputVideos(task.urls)}, nil
}
