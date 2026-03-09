# OpenAI Chat Examples (Qiniu, provider_v1)

This directory contains runnable OpenAI-compatible chat completion examples backed by Qiniu endpoints.

The examples are wired to `provider_v1` (Chat Completions API) through:

- `spec/openai/provider/qiniu.NewService`
- `examples/openai/shared.NewService`

## What Is Included

- Text-only chat
- Image + text chat
- Image detail levels (`low`, `ultra_high`)
- Video input (`URL` or `qfile-...` style ID)
- Multi-video chat (with text between videos)
- Function calling round-trip (`tool_use` -> local execution -> `tool_result`)
- Thinking mode comparison with `openai.WithThinking(...)`
- Streaming and non-streaming modes
- Unified block-structured output for all demos

## Prerequisites

- Go `1.25.5+`
- A valid `QINIU_API_KEY`
- Network access to Qiniu OpenAI-compatible endpoints

## Quick Start

```bash
# 1) Set API key
export QINIU_API_KEY=your-key

# 2) List available demos
go run ./examples/openai

# 3) Run one demo
go run ./examples/openai text

# 4) Run multiple demos
go run ./examples/openai text image video
```

## Stream Control

You can control stream mode in two ways.

### CLI flags

```bash
go run ./examples/openai --stream text
go run ./examples/openai --no-stream text
```

### Environment variable

`STREAM=1`, `true`, `yes`, or `on` enables stream mode.

```bash
STREAM=1 go run ./examples/openai text
```

## Demo Matrix

- `text`: text-only prompt (`chat_text.go`)
- `image`: image URL + text (`chat_image.go`)
- `image-detail`: image URL with `detail=low` (`chat_image_detail.go`)
- `image-ultra`: image URL with `detail=ultra_high` (`chat_image_detail.go`)
- `video`: video URL + text (`chat_video.go`)
- `video-fileid`: `qfile-...` video ID + text (`chat_video.go`)
- `multi-video`: two videos with text between (`chat_multi_video.go`)
- `function-call`: function calling full loop (`chat_function_call.go`)
- `thinking`: thinking enabled vs disabled (`chat_thinking.go`)

## Output Format (Block Structure)

All demos print model outputs in a consistent block-oriented structure using:

- `shared.PrintResponseBlocks(...)`
- `shared.PrintResponseBlocksWithTitle(...)`

Structure:

- `response { candidates: N }`
- `candidate[i] { stop_reason: ..., blocks: M }`
- `block[j] { type: ... }`

Supported block types in printer:

- `text`
- `tool_use`
- `tool_result`
- `thinking`
- `compaction`
- `blob`
- `unknown` (fallback)

Example:

```text
first_response
response { candidates: 1 }
  candidate[0] { stop_reason: "pause_turn", blocks: 1 }
    block[0] {
      type: "tool_use"
      id: "call_xxx"
      name: "get_weather"
      input_json:
        {
          "city": "shanghai",
          "unit": "celsius"
        }
    }
```

## Streaming Behavior

In stream mode, each incremental chunk is printed as:

- `stream_chunk[0]`
- `stream_chunk[1]`
- ...

Each chunk is still printed using the same `response -> candidate -> block` structure.

## Function Calling Demo Details

The `function-call` demo shows a complete two-step tool loop:

1. Define tool with `svc.ToolDef("get_weather")`
2. Send first request with `Tools(...)`
3. Read `tool_use` blocks from model output
4. Execute a local mock tool
5. Send `tool_result` back to model
6. Send second request and print final blocks

Printed sections:

- `first_response`
- `final_response`

Note: this demo intentionally uses non-stream requests to keep the full round-trip flow explicit.

## Thinking Demo Details

The `thinking` demo toggles thinking mode with:

```go
opts := openai.WithThinking(svc.Options(), true)
```

and compares output against:

```go
opts := openai.WithThinking(svc.Options(), false)
```

Model used in this demo: `deepseek/deepseek-v3.2-251201`.

## Defaults

Defined in `examples/openai/shared/service.go`:

- Default model: `gemini-3.0-pro-preview`
- Thinking demo model: `deepseek/deepseek-v3.2-251201`
- Default base URL: `https://api.qnaigc.com/v1/`
- Optional overseas base URL: `https://openai.sufy.com/v1/`

## Use a Custom Base URL

```go
package main

import (
	"os"

	"github.com/goplus/xai/spec/openai/provider/qiniu"
)

func main() {
	svc := qiniu.NewService(
		os.Getenv("QINIU_API_KEY"),
		qiniu.WithBaseURL(qiniu.OverseasBaseURL),
	)
	_ = svc
}
```

## Directory Layout

```text
examples/openai/
├── README.md
├── main.go
├── urls.go
├── chat_text.go
├── chat_image.go
├── chat_image_detail.go
├── chat_video.go
├── chat_multi_video.go
├── chat_function_call.go
├── chat_thinking.go
└── shared/
    ├── blocks.go
    └── service.go
```

## Troubleshooting

- `401` / auth errors:
  - Ensure `QINIU_API_KEY` is set and valid.
- `dial tcp ... no such host`:
  - Check DNS/network access to `api.qnaigc.com`.
- Tool schema error around `parameters.type`:
  - Ensure you are using the latest code where function tool schema includes:
    - `parameters.type = "object"`
    - `parameters.properties = {}`
- Stream output looks fragmented:
  - Expected behavior. Stream mode prints incremental chunks.

## Verify Locally

```bash
go test ./examples/openai/...
```
