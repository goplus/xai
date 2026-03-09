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

package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	mu      sync.Mutex
	results []string
)

func Append(model, demo string, urls []string, curl string) {
	mu.Lock()
	defer mu.Unlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== %s / %s ===\n", model, demo))
	for _, url := range urls {
		sb.WriteString(fmt.Sprintf("  URL: %s\n", url))
	}
	if curl != "" {
		sb.WriteString(fmt.Sprintf("  Curl:\n%s\n", curl))
	}
	sb.WriteString("\n")
	results = append(results, sb.String())
}

func Flush() error {
	mu.Lock()
	defer mu.Unlock()

	if len(results) == 0 {
		return nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	outputPath := filepath.Join(dir, "examples", "kling", "output", "results.txt")
	if _, err := os.Stat(filepath.Dir(outputPath)); os.IsNotExist(err) {
		outputPath = filepath.Join(dir, "output", "results.txt")
	}

	f, err := os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, r := range results {
		if _, err := f.WriteString(r); err != nil {
			return err
		}
	}
	results = nil
	return nil
}
