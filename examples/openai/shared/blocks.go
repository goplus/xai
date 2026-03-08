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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	xai "github.com/goplus/xai/spec"
)

// PrintResponseBlocks prints GenResponse as block-structured output.
func PrintResponseBlocks(resp xai.GenResponse) {
	PrintResponseBlocksWithTitle("", resp)
}

// PrintResponseBlocksWithTitle prints GenResponse with an optional title.
func PrintResponseBlocksWithTitle(title string, resp xai.GenResponse) {
	if title != "" {
		fmt.Printf("%s\n", title)
	}
	if resp == nil {
		fmt.Println("response: <nil>")
		return
	}
	fmt.Printf("response { candidates: %d }\n", resp.Len())
	for i := 0; i < resp.Len(); i++ {
		printCandidate("  ", i, resp.At(i))
	}
}

func printCandidate(prefix string, idx int, cand xai.Candidate) {
	fmt.Printf("%scandidate[%d] { stop_reason: %q, blocks: %d }\n", prefix, idx, string(cand.StopReason()), cand.Parts())
	for i := 0; i < cand.Parts(); i++ {
		printPart(prefix+"  ", i, cand.Part(i))
	}
}

func printPart(prefix string, idx int, part xai.Part) {
	fmt.Printf("%sblock[%d] {\n", prefix, idx)
	if toolUse, ok := safeAsToolUse(part); ok {
		fmt.Printf("%s  type: %q\n", prefix, "tool_use")
		fmt.Printf("%s  id: %q\n", prefix, toolUse.ID)
		fmt.Printf("%s  name: %q\n", prefix, toolUse.Name)
		printMultiline(prefix, "input_json", prettyJSON(toolUse.Input))
		fmt.Printf("%s}\n", prefix)
		return
	}
	if toolRet, ok := safeAsToolResult(part); ok {
		fmt.Printf("%s  type: %q\n", prefix, "tool_result")
		fmt.Printf("%s  id: %q\n", prefix, toolRet.ID)
		fmt.Printf("%s  name: %q\n", prefix, toolRet.Name)
		fmt.Printf("%s  is_error: %t\n", prefix, toolRet.IsError)
		printMultiline(prefix, "result_json", prettyJSON(toolRet.Result))
		fmt.Printf("%s}\n", prefix)
		return
	}
	if thinking, ok := safeAsThinking(part); ok {
		fmt.Printf("%s  type: %q\n", prefix, "thinking")
		fmt.Printf("%s  redacted: %t\n", prefix, thinking.Redacted)
		if thinking.Text != "" {
			fmt.Printf("%s  text: %q\n", prefix, thinking.Text)
		}
		if thinking.Signature != "" {
			fmt.Printf("%s  signature: %q\n", prefix, thinking.Signature)
		}
		fmt.Printf("%s}\n", prefix)
		return
	}
	if compaction, ok := safeAsCompaction(part); ok {
		fmt.Printf("%s  type: %q\n", prefix, "compaction")
		fmt.Printf("%s  data: %q\n", prefix, compaction.Data)
		fmt.Printf("%s}\n", prefix)
		return
	}
	if blob, ok := safeAsBlob(part); ok {
		fmt.Printf("%s  type: %q\n", prefix, "blob")
		fmt.Printf("%s  display_name: %q\n", prefix, blob.DisplayName)
		fmt.Printf("%s  mime: %q\n", prefix, blob.MIME)
		fmt.Printf("%s}\n", prefix)
		return
	}
	if text := strings.TrimSpace(part.Text()); text != "" {
		fmt.Printf("%s  type: %q\n", prefix, "text")
		fmt.Printf("%s  text: %q\n", prefix, text)
		fmt.Printf("%s}\n", prefix)
		return
	}
	fmt.Printf("%s  type: %q\n", prefix, "unknown")
	if u := part.Underlying(); u != nil {
		fmt.Printf("%s  underlying: %q\n", prefix, reflect.TypeOf(u).String())
	}
	fmt.Printf("%s}\n", prefix)
}

func printMultiline(prefix, key, value string) {
	if !strings.Contains(value, "\n") {
		fmt.Printf("%s  %s: %s\n", prefix, key, value)
		return
	}
	fmt.Printf("%s  %s:\n", prefix, key)
	for _, line := range strings.Split(value, "\n") {
		fmt.Printf("%s    %s\n", prefix, line)
	}
}

func prettyJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%q", fmt.Sprint(v))
	}
	return string(b)
}

func safeAsToolUse(part xai.Part) (ret xai.ToolUse, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return part.AsToolUse()
}

func safeAsToolResult(part xai.Part) (ret xai.ToolResult, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return part.AsToolResult()
}

func safeAsThinking(part xai.Part) (ret xai.Thinking, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return part.AsThinking()
}

func safeAsCompaction(part xai.Part) (ret xai.Compaction, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return part.AsCompaction()
}

func safeAsBlob(part xai.Part) (ret xai.Blob, ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	return part.AsBlob()
}
