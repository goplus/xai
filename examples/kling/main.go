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
 * See the License for the specific language governing permissions and limitations under the License.
 */

// Kling image and video generation examples. Dispatches to images/ and video/ subdirs by model.
// Run: go run ./examples/kling [model]
// With QINIU_API_KEY set, uses real Qnagic API; otherwise mock.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var modelOrder = []string{
	"models",
	"kling-v1", "kling-v1-5", "kling-v2", "kling-v2-new", "kling-v2-1", "kling-image-o1",
	"kling-v2-5-turbo", "kling-v2-6", "kling-video-o1", "kling-v3", "kling-v3-omni",
}

var imageModels = map[string]bool{
	"kling-v1": true, "kling-v1-5": true, "kling-v2": true, "kling-v2-new": true,
	"kling-v2-1": true, "kling-image-o1": true,
}
var videoModels = map[string]bool{
	"kling-v2-1": true, "kling-v2-5-turbo": true, "kling-v2-6": true,
	"kling-video-o1": true, "kling-v3": true, "kling-v3-omni": true,
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Running all Kling examples:")
		RunModels()
		runSubprogram("images", []string{})
		runSubprogram("video", []string{})
		return
	}

	for _, arg := range args {
		if arg == "models" {
			RunModels()
			continue
		}
		if imageModels[arg] {
			runSubprogram("images", []string{arg})
		}
		if videoModels[arg] {
			runSubprogram("video", []string{arg})
		}
		if !imageModels[arg] && !videoModels[arg] {
			fmt.Printf("Unknown model: %s\nAvailable: %v\n", arg, modelOrder)
		}
	}
}

func runSubprogram(subdir string, modelArgs []string) {
	cwd, _ := os.Getwd()
	subPath := filepath.Join(cwd, subdir)
	if _, err := os.Stat(subPath); err != nil {
		subPath = filepath.Join(cwd, "examples", "kling", subdir)
	}
	if _, err := os.Stat(subPath); err != nil {
		fmt.Printf("Error: cannot find %s directory\n", subdir)
		return
	}
	runPath := filepath.Join("examples", "kling", subdir)
	if filepath.Base(cwd) == "kling" {
		runPath = subdir
	}
	cmd := exec.Command("go", append([]string{"run", "./" + runPath}, modelArgs...)...)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running %s: %v\n", subdir, err)
	}
}
