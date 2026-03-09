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

// Audio examples: ASR (speech-to-text) and TTS (text-to-speech) via spec/audio.
// Run: go run ./examples/audio [demo]
// Set QINIU_API_KEY for real API calls; otherwise uses mock responses.
package main

import (
	"fmt"
	"os"
)

var demos = map[string]func(){
	"list-voices": runListVoices,
	"asr":         runASR,
	"tts":         runTTS,
	"tts-voice":   runTTSWithVoice,
}

var demoOrder = []string{
	"list-voices",
	"asr",
	"tts",
	"tts-voice",
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Audio examples (ASR / TTS):")
		fmt.Println("  Set QINIU_API_KEY for real API calls")
		fmt.Println()
		for _, name := range demoOrder {
			fmt.Printf("  %-12s %s\n", name, demoDesc(name))
		}
		fmt.Printf("  %-12s %s\n", "all", "run all demos")
		fmt.Println()
		fmt.Println("Usage: go run ./examples/audio [demo|all]")
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
	case "list-voices":
		return "List available TTS voices"
	case "asr":
		return "ASR: speech-to-text (Transcribe)"
	case "tts":
		return "TTS: text-to-speech (Synthesize)"
	case "tts-voice":
		return "TTS with specific voice"
	default:
		return ""
	}
}
