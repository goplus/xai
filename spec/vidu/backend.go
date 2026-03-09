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

	xai "github.com/goplus/xai/spec"
)

// Backend defines the transport/backend capabilities needed by spec/vidu.
// Implementations can be based on Qiniu, Fal.ai, or any vendor-specific protocol.
type Backend interface {
	Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error)
	GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error)
}
