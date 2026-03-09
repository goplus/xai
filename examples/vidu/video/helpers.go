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

package main

import (
	"fmt"
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/vidu"

	exampleshared "github.com/goplus/xai/examples/vidu/shared"
)

func newService() (*vidu.Service, error) {
	return exampleshared.NewService()
}

func newViduOptions(svc *vidu.Service, userID string) xai.OptionBuilder {
	opts := svc.Options()
	if o, ok := opts.(*vidu.Options); ok && strings.TrimSpace(userID) != "" {
		o.WithUserID(strings.TrimSpace(userID))
	}
	return opts
}

func progressPrinter(label string) func(resp xai.OperationResponse) {
	return func(resp xai.OperationResponse) {
		if taskID := resp.TaskID(); taskID != "" {
			fmt.Printf("  [%s] polling task: %s\n", label, taskID)
		}
	}
}
