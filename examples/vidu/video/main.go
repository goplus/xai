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

// Vidu video generation examples.
// Run: go run ./examples/vidu/video [demo]
// With QINIU_API_KEY set, uses real Qnagic API; otherwise mock.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/goplus/xai/examples/vidu/output"
	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/vidu"
)

var demos = map[string]func(){
	"q1-text":               runDemoQ1Text,
	"q1-ref-urls":           runDemoQ1RefURLs,
	"q1-ref-subjects":       runDemoQ1RefSubjects,
	"q1-ref-subjects-audio": runDemoQ1RefSubjectsAudio,
	"q2-text":               runDemoQ2Text,
	"q2-ref-urls":           runDemoQ2RefURLs,
	"q2-ref-subjects":       runDemoQ2RefSubjects,
	"q2-image-pro":          runDemoQ2ImagePro,
	"q2-image-pro-audio":    runDemoQ2ImageProAudio,
	"q2-image-turbo":        runDemoQ2ImageTurbo,
	"q2-start-end-pro":      runDemoQ2StartEndPro,
	"call-sync":             RunCallSyncExample,
}

var demoOrder = []string{
	"q1-text", "q1-ref-urls", "q1-ref-subjects", "q1-ref-subjects-audio",
	"q2-text", "q2-ref-urls", "q2-ref-subjects", "q2-image-pro", "q2-image-pro-audio", "q2-image-turbo", "q2-start-end-pro",
	"call-sync",
}

func main() {
	defer func() {
		if err := output.Flush(); err != nil {
			fmt.Println("flush output error:", err)
		}
	}()

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Vidu video examples:")
		fmt.Println("  Set QINIU_API_KEY for real API calls")
		fmt.Println()
		for _, name := range demoOrder {
			fmt.Printf("  %-22s %s\n", name, demoDesc(name))
		}
		fmt.Printf("  %-22s %s\n", "all", "run all demos")
		fmt.Println()
		fmt.Println("Usage: go run ./examples/vidu/video [demo|all]")
		return
	}

	for _, arg := range args {
		if arg == "all" {
			for _, name := range demoOrder {
				fmt.Println("---", name, "---")
				demos[name]()
			}
			continue
		}
		if fn, ok := demos[arg]; ok {
			fmt.Println("---", arg, "---")
			fn()
		} else {
			fmt.Printf("Unknown demo: %s\nAvailable: %v + all\n", arg, demoOrder)
		}
	}
}

func demoDesc(name string) string {
	switch name {
	case "q1-text":
		return "Q1 text-to-video (style/bgm/aspect)"
	case "q1-ref-urls":
		return "Q1 reference-to-video (URLs)"
	case "q1-ref-subjects":
		return "Q1 reference-to-video (subjects)"
	case "q1-ref-subjects-audio":
		return "Q1 reference-to-video (subjects + audio)"
	case "q2-text":
		return "Q2 text-to-video (aspect/bgm)"
	case "q2-ref-urls":
		return "Q2 reference-to-video (URLs)"
	case "q2-ref-subjects":
		return "Q2 reference-to-video (subjects)"
	case "q2-image-pro":
		return "Q2 image-to-video-pro"
	case "q2-image-pro-audio":
		return "Q2 Pro image-to-video (audio + voice_id)"
	case "q2-image-turbo":
		return "Q2 Turbo image-to-video"
	case "q2-start-end-pro":
		return "Q2 start-end-to-video-pro"
	case "call-sync":
		return "CallSync + TaskID + GetTask resume"
	default:
		return ""
	}
}

func runDemo(ctx context.Context, model xai.Model, runFn func(context.Context, xai.Service, xai.Model) error) {
	svc, err := newService()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if err := runFn(ctx, svc, model); err != nil {
		fmt.Println("Error:", err)
	}
}

func runDemoQ1Text() {
	runDemo(context.Background(), xai.Model(vidu.ModelViduQ1), runViduQ1TextToVideo)
}

func runDemoQ1RefURLs() {
	runDemo(context.Background(), xai.Model(vidu.ModelViduQ1), runViduQ1ReferenceToVideoURLs)
}

func runDemoQ1RefSubjects() {
	runDemo(context.Background(), xai.Model(vidu.ModelViduQ1), runViduQ1ReferenceToVideoSubjects)
}

func runDemoQ2Text() {
	runDemo(context.Background(), xai.Model(vidu.ModelViduQ2), runViduQ2TextToVideo)
}

func runDemoQ2RefURLs() {
	runDemo(context.Background(), xai.Model(vidu.ModelViduQ2), runViduQ2ReferenceToVideoURLs)
}

func runDemoQ2RefSubjects() {
	runDemo(context.Background(), xai.Model(vidu.ModelViduQ2), runViduQ2ReferenceToVideoSubjects)
}

func runDemoQ2ImagePro() {
	runDemo(context.Background(), xai.Model(vidu.ModelViduQ2), runViduQ2ImageToVideoPro)
}

func runDemoQ2ImageProAudio() {
	runDemo(context.Background(), xai.Model(vidu.ModelViduQ2Pro), runViduQ2ImageToVideoProAudio)
}

func runDemoQ2ImageTurbo() {
	runDemo(context.Background(), xai.Model(vidu.ModelViduQ2Turbo), runViduQ2ImageToVideoTurbo)
}

func runDemoQ2StartEndPro() {
	runDemo(context.Background(), xai.Model(vidu.ModelViduQ2), runViduQ2StartEndToVideoPro)
}

func runDemoQ1RefSubjectsAudio() {
	runDemo(context.Background(), xai.Model(vidu.ModelViduQ1), runViduQ1ReferenceToVideoSubjectsAudio)
}
