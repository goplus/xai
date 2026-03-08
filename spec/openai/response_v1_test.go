package openai

import (
	"encoding/json"
	"testing"

	xai "github.com/goplus/xai/spec"
	openai "github.com/openai/openai-go/v3"
)

func expectPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic, got nil")
		}
	}()
	fn()
}

func mustUnmarshalChoice(t *testing.T, s string) openai.ChatCompletionChoice {
	t.Helper()
	var c openai.ChatCompletionChoice
	if err := json.Unmarshal([]byte(s), &c); err != nil {
		t.Fatalf("unmarshal choice: %v", err)
	}
	return c
}

func mustUnmarshalToolCall(t *testing.T, s string) openai.ChatCompletionMessageToolCallUnion {
	t.Helper()
	var c openai.ChatCompletionMessageToolCallUnion
	if err := json.Unmarshal([]byte(s), &c); err != nil {
		t.Fatalf("unmarshal tool call: %v", err)
	}
	return c
}

func testToolCallFunction(t *testing.T) openai.ChatCompletionMessageToolCallUnion {
	t.Helper()
	return mustUnmarshalToolCall(t, `{
		"id":"call_fn_1",
		"type":"function",
		"function":{
			"name":"get_weather",
			"arguments":"{\"city\":\"SF\"}"
		}
	}`)
}

func testToolCallCustom(t *testing.T) openai.ChatCompletionMessageToolCallUnion {
	t.Helper()
	return mustUnmarshalToolCall(t, `{
		"id":"call_custom_1",
		"type":"custom",
		"custom":{
			"name":"custom_echo",
			"input":"{\"text\":\"hello\"}"
		}
	}`)
}

func TestParseV1LegacyFunctionCall(t *testing.T) {
	if got := parseV1LegacyFunctionCall(""); got != nil {
		t.Fatalf("expected nil for empty input")
	}
	if got := parseV1LegacyFunctionCall("not-json"); got != nil {
		t.Fatalf("expected nil for invalid json")
	}
	if got := parseV1LegacyFunctionCall(`{"function_call":{"arguments":"{}"}}`); got != nil {
		t.Fatalf("expected nil for missing function name")
	}
	got := parseV1LegacyFunctionCall(`{"function_call":{"name":"legacy_tool","arguments":"{\"x\":1}"}}`)
	if got == nil {
		t.Fatalf("expected parsed legacy function call")
	}
	if got.Name != "legacy_tool" || got.Arguments != `{"x":1}` {
		t.Fatalf("unexpected parsed value: %+v", got)
	}
}

func TestV1ContentBlockAsToolUse(t *testing.T) {
	fn := v1ContentBlock{toolCall: &([]openai.ChatCompletionMessageToolCallUnion{testToolCallFunction(t)})[0]}
	fnTool, ok := fn.AsToolUse()
	if !ok {
		t.Fatalf("expected function tool use")
	}
	if fnTool.ID != "call_fn_1" || fnTool.Name != "get_weather" {
		t.Fatalf("unexpected function tool use: %+v", fnTool)
	}
	if _, ok := fnTool.Input.(json.RawMessage); !ok {
		t.Fatalf("expected raw json input for function tool")
	}

	custom := v1ContentBlock{toolCall: &([]openai.ChatCompletionMessageToolCallUnion{testToolCallCustom(t)})[0]}
	customTool, ok := custom.AsToolUse()
	if !ok {
		t.Fatalf("expected custom tool use")
	}
	if customTool.ID != "call_custom_1" || customTool.Name != "custom_echo" {
		t.Fatalf("unexpected custom tool use: %+v", customTool)
	}
	if _, ok := customTool.Input.(json.RawMessage); !ok {
		t.Fatalf("expected raw json input for custom tool")
	}

	legacy := v1ContentBlock{legacyFunctionCall: &v1LegacyFunctionCall{Name: "legacy_tool", Arguments: "not-json"}}
	legacyTool, ok := legacy.AsToolUse()
	if !ok {
		t.Fatalf("expected legacy tool use")
	}
	if legacyTool.Name != "legacy_tool" {
		t.Fatalf("unexpected legacy tool name: %s", legacyTool.Name)
	}
	wrapped, ok := legacyTool.Input.(map[string]any)
	if !ok || wrapped["raw"] != "not-json" {
		t.Fatalf("unexpected legacy input: %#v", legacyTool.Input)
	}
}

func TestNewV1CandidateWithOverridesAndPartOrder(t *testing.T) {
	choice := &openai.ChatCompletionChoice{
		FinishReason: "tool_calls",
		Message: openai.ChatCompletionMessage{
			Content:   "original",
			ToolCalls: []openai.ChatCompletionMessageToolCallUnion{testToolCallFunction(t)},
		},
	}
	raw := &chatCompletionMessageRaw{
		Content:   "override-content",
		ToolCalls: []openai.ChatCompletionMessageToolCallUnion{testToolCallCustom(t)},
	}
	img := xai.Blob{MIME: "image/png", BlobData: xai.BlobFromRaw([]byte{0x1, 0x2})}
	cand := newV1Candidate(choice, []xai.Blob{img}, raw)

	if got, want := cand.StopReason(), xai.PauseTurn; got != want {
		t.Fatalf("StopReason got=%s want=%s", got, want)
	}
	if got, want := cand.Parts(), 3; got != want {
		t.Fatalf("Parts got=%d want=%d", got, want)
	}

	if got := cand.Part(0).Text(); got != "override-content" {
		t.Fatalf("unexpected text part: %q", got)
	}
	if blob, ok := cand.Part(1).AsBlob(); !ok || blob.MIME != "image/png" {
		t.Fatalf("expected blob at part 1, got=%+v ok=%v", blob, ok)
	}
	if tool, ok := cand.Part(2).AsToolUse(); !ok || tool.Name != "custom_echo" {
		t.Fatalf("expected custom tool at part 2, got=%+v ok=%v", tool, ok)
	}

	expectPanic(t, func() { _ = cand.Part(3) })
}

func TestV1CandidateLegacyFunctionCallFromRawJSON(t *testing.T) {
	choice := mustUnmarshalChoice(t, `{
		"finish_reason": "function_call",
		"index": 0,
		"message": {
			"role": "assistant",
			"content": "",
			"function_call": {
				"name": "legacy_lookup",
				"arguments": "{\"query\":\"abc\"}"
			}
		}
	}`)

	cand := newV1Candidate(&choice, nil, nil)
	if got, want := cand.Parts(), 1; got != want {
		t.Fatalf("Parts got=%d want=%d", got, want)
	}
	tool, ok := cand.Part(0).AsToolUse()
	if !ok {
		t.Fatalf("expected tool use for legacy function_call")
	}
	if tool.Name != "legacy_lookup" {
		t.Fatalf("unexpected tool name: %s", tool.Name)
	}
	if got, want := cand.StopReason(), xai.PauseTurn; got != want {
		t.Fatalf("StopReason got=%s want=%s", got, want)
	}
}

func TestV1CandidateToMsg(t *testing.T) {
	cand := &v1Candidate{
		text:      "hello",
		toolCalls: []openai.ChatCompletionMessageToolCallUnion{testToolCallFunction(t)},
	}
	mb, ok := cand.ToMsg().(*msgBuilder)
	if !ok {
		t.Fatalf("expected *msgBuilder")
	}
	if mb.msg.Role != "assistant" {
		t.Fatalf("unexpected role: %s", mb.msg.Role)
	}
	if got, want := len(mb.msg.Contents), 2; got != want {
		t.Fatalf("contents len got=%d want=%d", got, want)
	}
	if mb.msg.Contents[0].Type != contentText || mb.msg.Contents[0].Text != "hello" {
		t.Fatalf("unexpected first content: %+v", mb.msg.Contents[0])
	}
	if mb.msg.Contents[1].Type != contentToolUse || mb.msg.Contents[1].ToolUse == nil {
		t.Fatalf("unexpected second content: %+v", mb.msg.Contents[1])
	}
	if mb.msg.Contents[1].ToolUse.Name != "get_weather" {
		t.Fatalf("unexpected tool name: %s", mb.msg.Contents[1].ToolUse.Name)
	}
}

func TestV1ResponseAndV1ResponseWithImages(t *testing.T) {
	resp := &v1Response{msg: &openai.ChatCompletion{Choices: []openai.ChatCompletionChoice{
		{FinishReason: "stop", Message: openai.ChatCompletionMessage{Content: "c1"}},
		{FinishReason: "length", Message: openai.ChatCompletionMessage{Content: "c2"}},
	}}}
	if got, want := resp.Len(), 2; got != want {
		t.Fatalf("Len got=%d want=%d", got, want)
	}
	if got, want := resp.At(1).Part(0).Text(), "c2"; got != want {
		t.Fatalf("At(1) text got=%q want=%q", got, want)
	}
	if got, want := resp.StopReason(), xai.EndTurn; got != want {
		t.Fatalf("StopReason got=%s want=%s", got, want)
	}
	expectPanic(t, func() { _ = resp.At(2) })

	img := xai.Blob{MIME: "image/png", BlobData: xai.BlobFromRaw([]byte{0x3})}
	raw := &chatCompletionMessageRaw{Content: "img-override"}
	ext := &v1ResponseWithImages{
		msg: &openai.ChatCompletion{Choices: []openai.ChatCompletionChoice{
			{FinishReason: "stop", Message: openai.ChatCompletionMessage{Content: "plain"}},
		}},
		images: [][]xai.Blob{{img}},
		raw:    []*chatCompletionMessageRaw{raw},
	}
	cand := ext.At(0)
	if got, want := cand.Parts(), 2; got != want {
		t.Fatalf("parts got=%d want=%d", got, want)
	}
	if got, want := cand.Part(0).Text(), "img-override"; got != want {
		t.Fatalf("text got=%q want=%q", got, want)
	}
	if blob, ok := cand.Part(1).AsBlob(); !ok || blob.MIME != "image/png" {
		t.Fatalf("unexpected blob part: %+v ok=%v", blob, ok)
	}

	msg := ext.ToMsg().(*msgBuilder)
	if msg.msg.Role != "assistant" || len(msg.msg.Contents) == 0 {
		t.Fatalf("unexpected ToMsg result: %+v", msg.msg)
	}
	expectPanic(t, func() { _ = ext.At(1) })
}

func TestV1StreamChunk(t *testing.T) {
	chunk := &v1StreamChunk{text: "delta"}
	if got, want := chunk.Len(), 1; got != want {
		t.Fatalf("Len got=%d want=%d", got, want)
	}
	if got, want := chunk.Parts(), 1; got != want {
		t.Fatalf("Parts got=%d want=%d", got, want)
	}
	if got, want := chunk.Part(0).Text(), "delta"; got != want {
		t.Fatalf("Part(0).Text got=%q want=%q", got, want)
	}
	if got, want := chunk.StopReason(), xai.Unspecified; got != want {
		t.Fatalf("StopReason got=%s want=%s", got, want)
	}
	if cand := chunk.At(0); cand != chunk {
		t.Fatalf("At(0) should return chunk itself")
	}
	if mb := chunk.ToMsg().(*msgBuilder); mb.msg.Role != "assistant" || mb.msg.Contents[0].Text != "delta" {
		t.Fatalf("unexpected ToMsg content: %+v", mb.msg)
	}
	expectPanic(t, func() { _ = chunk.Part(1) })
	expectPanic(t, func() { _ = chunk.At(1) })
}
