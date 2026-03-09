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
	"os"

	"github.com/goplus/xai/spec/vidu"
)

// NewService creates a vidu.Service with Qiniu backend.
func NewService(token string, opts ...ClientOption) *vidu.Service {
	if token == "" {
		token = os.Getenv("QINIU_API_KEY")
	}
	client := NewClient(token, opts...)
	return vidu.NewWithBackend(newBackend(client))
}

// Register creates a vidu.Service with Qiniu backend and registers it with xai.
func Register(token string, opts ...ClientOption) {
	svc := NewService(token, opts...)
	vidu.Register(svc)
}
