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
	"context"
	"fmt"
	"os"
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/openai"
	"github.com/goplus/xai/spec/openai/provider/qiniu"
)

func newService() *openai.Service {
	token := strings.TrimSpace(os.Getenv("QINIU_API_KEY"))
	return qiniu.NewService(token)
}

func runOperation(ctx context.Context, svc *openai.Service, op xai.Operation, name string) {
	opts := openai.WithDebugCurl(svc.Options(), true)
	resp, err := xai.CallSync(ctx, svc, op, opts)
	if err != nil {
		fmt.Println("Call error:", err)
		return
	}
	if taskID := resp.TaskID(); taskID != "" {
		fmt.Println("task_id:", taskID)
	}

	results, err := xai.Wait(ctx, svc, resp, func(resp xai.OperationResponse) {
		if taskID := resp.TaskID(); taskID != "" {
			fmt.Printf("  [%s] polling task: %s\n", name, taskID)
		}
	})
	if err != nil {
		fmt.Println("Wait error:", err)
		return
	}

	fmt.Printf("%s: videos=%d\n", name, results.Len())
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputVideo)
		fmt.Printf("  video[%d]: %s (%s)\n", i, out.URL(), out.Video.Type())
	}
}
