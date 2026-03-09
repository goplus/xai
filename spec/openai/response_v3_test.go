package openai

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	xai "github.com/goplus/xai/spec"
	"github.com/openai/openai-go/v3/responses"
)

func mustUnmarshalOutputItem(t *testing.T, s string) responses.ResponseOutputItemUnion {
	t.Helper()
	var item responses.ResponseOutputItemUnion
	if err := json.Unmarshal([]byte(s), &item); err != nil {
		t.Fatalf("unmarshal output item: %v", err)
	}
	return item
}

func mustUnmarshalResponse(t *testing.T, s string) responses.Response {
	t.Helper()
	var resp responses.Response
	if err := json.Unmarshal([]byte(s), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	return resp
}

func rawBlobBytes(t *testing.T, b xai.Blob) []byte {
	t.Helper()
	if b.BlobData == nil {
		t.Fatalf("blob data is nil")
	}
	raw, err := b.BlobData.Raw()
	if err != nil {
		t.Fatalf("blob raw failed: %v", err)
	}
	return raw
}

func TestRawJSONOrString(t *testing.T) {
	if got := rawJSONOrString(`{"a":1}`); got == nil {
		t.Fatalf("expected non-nil")
	} else if _, ok := got.(json.RawMessage); !ok {
		t.Fatalf("expected json.RawMessage, got %T", got)
	}

	if got := rawJSONOrString("not-json"); got == nil {
		t.Fatalf("expected wrapped map")
	} else if m, ok := got.(map[string]any); !ok || m["raw"] != "not-json" {
		t.Fatalf("unexpected wrapped value: %#v", got)
	}

	if got := rawJSONOrString("   "); got == nil {
		t.Fatalf("expected default empty json")
	} else if raw, ok := got.(json.RawMessage); !ok || string(raw) != "{}" {
		t.Fatalf("unexpected empty fallback: %#v", got)
	}
}

func TestMarshalToolInput(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want string
	}{
		{name: "raw-valid", in: json.RawMessage(`{"k":1}`), want: `{"k":1}`},
		{name: "bytes-valid", in: []byte(`{"k":2}`), want: `{"k":2}`},
		{name: "string-valid", in: `{"k":3}`, want: `{"k":3}`},
		{name: "string-invalid", in: "oops", want: `{"raw":"oops"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := marshalToolInput(tc.in)
			if strings.TrimSpace(string(got)) != tc.want {
				t.Fatalf("marshalToolInput got=%s want=%s", string(got), tc.want)
			}
		})
	}
}

func TestV3OutputAndReasoningText(t *testing.T) {
	msg := v3OutputMessageText([]responses.ResponseOutputMessageContentUnion{
		{Type: "output_text", Text: "hello"},
		{Type: "refusal", Refusal: "blocked"},
		{Type: "output_text", Text: "world"},
	})
	if got, want := msg, "hello\nblocked\nworld"; got != want {
		t.Fatalf("v3OutputMessageText got=%q want=%q", got, want)
	}

	r1 := responses.ResponseReasoningItem{Content: []responses.ResponseReasoningItemContent{{Text: "step1"}, {Text: "step2"}}}
	if got, want := v3ReasoningText(r1), "step1\nstep2"; got != want {
		t.Fatalf("v3ReasoningText(content) got=%q want=%q", got, want)
	}

	r2 := responses.ResponseReasoningItem{Summary: []responses.ResponseReasoningItemSummary{{Text: "sum1"}, {Text: "sum2"}}}
	if got, want := v3ReasoningText(r2), "sum1\nsum2"; got != want {
		t.Fatalf("v3ReasoningText(summary) got=%q want=%q", got, want)
	}
}

func TestV3ContentBlockThinkingToolUseToolResult(t *testing.T) {
	reasoningItem := mustUnmarshalOutputItem(t, `{
		"type":"reasoning",
		"id":"reasoning-1",
		"summary":[{"type":"summary_text","text":"think"}],
		"encrypted_content":"enc-signature"
	}`)
	thinking, ok := (v3ContentBlock{content: &reasoningItem}).AsThinking()
	if !ok || thinking.Text != "think" || thinking.Signature != "enc-signature" {
		t.Fatalf("unexpected thinking: %+v ok=%v", thinking, ok)
	}

	functionItem := mustUnmarshalOutputItem(t, `{
		"type":"function_call",
		"id":"item-id-1",
		"call_id":"call-id-1",
		"name":"sum",
		"arguments":"{\"a\":1}"
	}`)
	toolUse, ok := (v3ContentBlock{content: &functionItem}).AsToolUse()
	if !ok || toolUse.ID != "call-id-1" || toolUse.Name != "sum" {
		t.Fatalf("unexpected function tool use: %+v ok=%v", toolUse, ok)
	}
	if _, ok := toolUse.Input.(json.RawMessage); !ok {
		t.Fatalf("expected raw json input")
	}

	mcpErrItem := mustUnmarshalOutputItem(t, `{
		"type":"mcp_call",
		"id":"mcp-1",
		"name":"tool-a",
		"error":"permission denied"
	}`)
	toolRes, ok := (v3ContentBlock{content: &mcpErrItem}).AsToolResult()
	if !ok || !toolRes.IsError || toolRes.Name != "tool-a" {
		t.Fatalf("unexpected mcp tool result: %+v ok=%v", toolRes, ok)
	}

	mcpOutItem := mustUnmarshalOutputItem(t, `{
		"type":"mcp_call",
		"id":"mcp-2",
		"name":"tool-b",
		"output":"{\"ok\":true}"
	}`)
	toolRes2, ok := (v3ContentBlock{content: &mcpOutItem}).AsToolResult()
	if !ok || toolRes2.IsError {
		t.Fatalf("unexpected mcp output result: %+v ok=%v", toolRes2, ok)
	}
	if _, ok := toolRes2.Result.(json.RawMessage); !ok {
		t.Fatalf("expected json.RawMessage in mcp output")
	}
}

func TestV3ContentBlockBlobCompactionText(t *testing.T) {
	pngData := []byte("PNG")
	imageItem := mustUnmarshalOutputItem(t, `{
		"type":"image_generation_call",
		"id":"img-1",
		"result":"`+base64.StdEncoding.EncodeToString(pngData)+`"
	}`)
	blob, ok := (v3ContentBlock{content: &imageItem}).AsBlob()
	if !ok || blob.MIME != "image/png" {
		t.Fatalf("unexpected blob: %+v ok=%v", blob, ok)
	}
	if got := rawBlobBytes(t, blob); string(got) != string(pngData) {
		t.Fatalf("blob bytes got=%q want=%q", string(got), string(pngData))
	}

	compactionItem := mustUnmarshalOutputItem(t, `{
		"type":"compaction",
		"id":"cmp-1",
		"encrypted_content":"cipher"
	}`)
	compaction, ok := (v3ContentBlock{content: &compactionItem}).AsCompaction()
	if !ok || compaction.Data != "cipher" {
		t.Fatalf("unexpected compaction: %+v ok=%v", compaction, ok)
	}

	messageItem := mustUnmarshalOutputItem(t, `{
		"type":"message",
		"id":"msg-1",
		"role":"assistant",
		"content":[
			{"type":"output_text","text":"hello"},
			{"type":"refusal","refusal":"blocked"}
		]
	}`)
	if got, want := (v3ContentBlock{content: &messageItem}).Text(), "hello\nblocked"; got != want {
		t.Fatalf("unexpected text extraction got=%q want=%q", got, want)
	}
}

func TestV3ResponseStopReason(t *testing.T) {
	cases := []struct {
		name string
		resp *responses.Response
		want xai.StopReason
	}{
		{
			name: "compaction-precedence",
			resp: &responses.Response{Status: responses.ResponseStatusCompleted, Output: []responses.ResponseOutputItemUnion{{Type: "compaction"}}},
			want: xai.StopCompaction,
		},
		{
			name: "host-tool-precedence",
			resp: &responses.Response{Status: responses.ResponseStatusCompleted, Output: []responses.ResponseOutputItemUnion{{Type: "function_call"}}},
			want: xai.PauseTurn,
		},
		{
			name: "completed",
			resp: &responses.Response{Status: responses.ResponseStatusCompleted},
			want: xai.EndTurn,
		},
		{
			name: "incomplete-max-tokens",
			resp: &responses.Response{Status: responses.ResponseStatusIncomplete, IncompleteDetails: responses.ResponseIncompleteDetails{Reason: "max_output_tokens"}},
			want: xai.StopMaxTokens,
		},
		{
			name: "incomplete-content-filter",
			resp: &responses.Response{Status: responses.ResponseStatusIncomplete, IncompleteDetails: responses.ResponseIncompleteDetails{Reason: "content_filter"}},
			want: xai.Refusal,
		},
		{
			name: "failed-with-error",
			resp: &responses.Response{Status: responses.ResponseStatusFailed, Error: responses.ResponseError{Message: "bad request"}},
			want: xai.Refusal,
		},
		{
			name: "unspecified",
			resp: &responses.Response{Status: responses.ResponseStatusInProgress},
			want: xai.Unspecified,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := (&v3Response{msg: tc.resp}).StopReason()
			if got != tc.want {
				t.Fatalf("StopReason got=%s want=%s", got, tc.want)
			}
		})
	}
}

func TestV3ResponsePartAndAt(t *testing.T) {
	resp := &v3Response{msg: &responses.Response{Status: responses.ResponseStatusCompleted, Output: []responses.ResponseOutputItemUnion{{Type: "message", Content: []responses.ResponseOutputMessageContentUnion{{Type: "output_text", Text: "x"}}}}}}
	if got, want := resp.Len(), 1; got != want {
		t.Fatalf("Len got=%d want=%d", got, want)
	}
	if got, want := resp.Parts(), 1; got != want {
		t.Fatalf("Parts got=%d want=%d", got, want)
	}
	if got, want := resp.Part(0).Text(), "x"; got != want {
		t.Fatalf("Part(0).Text got=%q want=%q", got, want)
	}
	if cand := resp.At(0); cand != resp {
		t.Fatalf("At(0) should return response itself")
	}
	expectPanic(t, func() { _ = resp.Part(1) })
	expectPanic(t, func() { _ = resp.At(1) })
}

func TestV3ResponseToMsg(t *testing.T) {
	respWire := mustUnmarshalResponse(t, `{
		"status":"completed",
		"output":[
			{
				"type":"message",
				"id":"msg-1",
				"role":"assistant",
				"content":[{"type":"output_text","text":"hello"}]
			},
			{
				"type":"reasoning",
				"id":"reason-1",
				"summary":[{"type":"summary_text","text":"think"}],
				"encrypted_content":"sig"
			},
			{
				"type":"function_call",
				"id":"fn-item",
				"call_id":"f1",
				"name":"tool_f",
				"arguments":"{\"x\":1}"
			},
			{
				"type":"custom_tool_call",
				"id":"custom-item",
				"call_id":"c1",
				"name":"tool_c",
				"input":"{\"y\":2}"
			},
			{
				"type":"compaction",
				"id":"cmp-1",
				"encrypted_content":"compact-data"
			},
			{
				"type":"image_generation_call",
				"id":"img-1",
				"result":"`+base64.StdEncoding.EncodeToString([]byte("img"))+`"
			}
		]
	}`)
	resp := &v3Response{msg: &respWire}

	mb, ok := resp.ToMsg().(*msgBuilder)
	if !ok {
		t.Fatalf("expected *msgBuilder")
	}
	if mb.msg.Role != "assistant" {
		t.Fatalf("unexpected role: %s", mb.msg.Role)
	}
	if got, want := len(mb.msg.Contents), 6; got != want {
		t.Fatalf("contents len got=%d want=%d", got, want)
	}

	if mb.msg.Contents[0].Type != contentText || mb.msg.Contents[0].Text != "hello" {
		t.Fatalf("unexpected text content: %+v", mb.msg.Contents[0])
	}
	if mb.msg.Contents[1].Type != contentThinking || mb.msg.Contents[1].Thinking == nil || mb.msg.Contents[1].Thinking.Text != "think" {
		t.Fatalf("unexpected thinking content: %+v", mb.msg.Contents[1])
	}
	if mb.msg.Contents[2].Type != contentToolUse || mb.msg.Contents[2].ToolUse == nil || mb.msg.Contents[2].ToolUse.Name != "tool_f" {
		t.Fatalf("unexpected tool use content #1: %+v", mb.msg.Contents[2])
	}
	if mb.msg.Contents[3].Type != contentToolUse || mb.msg.Contents[3].ToolUse == nil || mb.msg.Contents[3].ToolUse.Name != "tool_c" {
		t.Fatalf("unexpected tool use content #2: %+v", mb.msg.Contents[3])
	}
	if mb.msg.Contents[4].Type != contentCompaction || mb.msg.Contents[4].Compaction != "compact-data" {
		t.Fatalf("unexpected compaction content: %+v", mb.msg.Contents[4])
	}
	if mb.msg.Contents[5].Type != contentImageURL || !strings.HasPrefix(mb.msg.Contents[5].ImageURL, "data:image/png;base64,") {
		t.Fatalf("unexpected image content: %+v", mb.msg.Contents[5])
	}
}

func TestV3StreamChunk(t *testing.T) {
	chunk := &v3StreamChunk{text: "delta"}
	if got, want := chunk.Len(), 1; got != want {
		t.Fatalf("Len got=%d want=%d", got, want)
	}
	if got, want := chunk.Parts(), 1; got != want {
		t.Fatalf("Parts got=%d want=%d", got, want)
	}
	if got, want := chunk.Part(0).Text(), "delta"; got != want {
		t.Fatalf("Part(0).Text got=%q want=%q", got, want)
	}
	if cand := chunk.At(0); cand != chunk {
		t.Fatalf("At(0) should return chunk itself")
	}
	if mb := chunk.ToMsg().(*msgBuilder); mb.msg.Role != "assistant" || mb.msg.Contents[0].Text != "delta" {
		t.Fatalf("unexpected ToMsg result: %+v", mb.msg)
	}
	expectPanic(t, func() { _ = chunk.Part(1) })
	expectPanic(t, func() { _ = chunk.At(1) })
}
