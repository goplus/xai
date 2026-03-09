# xai

Unified Go SDK for AI chat, image generation, and video generation. Supports multiple providers (OpenAI-compatible, Gemini, Kling, Sora, Veo, Vidu) through a common API.

## Features

- **Chat Completions**: Text, image, video multimodal chat with streaming
- **Function Calling**: Tool use round-trip with `tool_use` / `tool_result`
- **Image Generation**: Text-to-image, image-to-image, image edit
- **Video Generation**: Text-to-video, image-to-video, remix, keyframe
- **Long-running Operations**: `CallSync` + `TaskID` + `GetTask` for async task persistence

## Prerequisites

- Go 1.24+
- `QINIU_API_KEY` for real API calls (omit for mock mode)

## Installation

```bash
go get github.com/goplus/xai
```

## Quick Start

```bash
# Set API key for real calls
export QINIU_API_KEY=your-key

# OpenAI-compatible chat (text, image, video, function calling)
go run ./examples/openai text
go run ./examples/openai image video function-call

# Gemini chat + image generation
go run ./examples/gemini chat-text image-generate

# Kling image & video
go run ./examples/kling/images kling-v2-1
go run ./examples/kling/video kling-v2-6

# Sora video
go run ./examples/sora text-to-video image-to-video

# Veo video
go run ./examples/veo veo-3.0-generate-preview

# Vidu video
go run ./examples/vidu/video q2-text
```

## Examples Overview

| Example | Description |
|---------|-------------|
| [examples/openai](examples/openai) | OpenAI-compatible chat: text, image, video, multi-video, function calling, thinking mode |
| [examples/gemini](examples/gemini) | Gemini chat + image generation / edit |
| [examples/kling](examples/kling) | Kling image & video: text2image, image2image, text2video, img2video, keyframe |
| [examples/sora](examples/sora) | Sora text-to-video, image-to-video, remix |
| [examples/veo](examples/veo) | Veo text-to-video, image-to-video, first+last frame, reference images |
| [examples/vidu](examples/vidu) | Vidu Q1/Q2 text-to-video, reference-to-video, image-to-video |

## Backend Mode

- **Mock** (default): No API key. Returns placeholder URLs. Works in CI.
- **Real**: Set `QINIU_API_KEY` to call Qnagic API.

## Supported Models

| Category | Models |
|----------|--------|
| Image | kling-v1, kling-v1-5, kling-v2, kling-v2-new, kling-v2-1, kling-image-o1 |
| Video | kling-v2-1, kling-v2-5-turbo, kling-v2-6, kling-video-o1, kling-v3, kling-v3-omni |
| Veo | veo-2.0-generate-001, veo-2.0-generate-exp, veo-3.0-generate-preview, veo-3.1-generate-preview, ... |
| Sora | sora-2, sora-2-pro |
| Chat | gemini-3.0-pro-preview, deepseek-v3.2, etc. |

## API Usage

```go
package main

import (
    "context"
    "os"

    xai "github.com/goplus/xai/spec"
    "github.com/goplus/xai/spec/openai/provider/qiniu"
)

func main() {
    svc := qiniu.NewService(os.Getenv("QINIU_API_KEY"))
    ctx := context.Background()

    // Chat (OpenAI-compatible)
    resp, _ := svc.Gen(ctx, svc.Params().
        Model(xai.Model("gemini-3.0-pro-preview")).
        Messages(svc.UserMsg().Text("Hello")), svc.Options())

    // Video generation (Sora): CallSync + Wait for async polling
    op, _ := svc.Operation(xai.Model("sora-2"), xai.GenVideo)
    op.Params().Set("Prompt", "A cat walking on the beach").Set("Seconds", "4")
    opResp, _ := xai.CallSync(ctx, svc, op, svc.Options())
    results, _ := xai.Wait(ctx, svc, opResp, nil)
}
```

## License

Apache-2.0
