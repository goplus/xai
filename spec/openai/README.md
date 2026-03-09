# OpenAI Spec for XAI

OpenAI-compatible implementation of [xai.Service](https://github.com/goplus/xai/spec), supporting both **v3 Responses API** and **v1 Chat Completions API**. Works with OpenAI, Qiniu (api.qnaigc.com), and any OpenAI-compatible endpoint.

## Features

- **Dual API support**: v3 Responses API (default) and v1 Chat Completions API
- **Text, image, document, video** input via `MsgBuilder` / `MsgBuilderExt`
- **Streaming** and non-streaming generation
- **Tools** (function calling) and **Web Search** tool
- **Thinking / reasoning** and **compaction** blocks
- **Provider extensions**: Qiniu provider with `ImageURLWithDetail`, `VideoFile`

## Installation

```bash
go get github.com/goplus/xai/spec/openai
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    xai "github.com/goplus/xai/spec"
    "github.com/goplus/xai/spec/openai"
)

func main() {
    ctx := context.Background()

    // Create service (v3 Responses API)
    svc, err := openai.New(ctx, "openai:base=https://api.openai.com/v1/&key=your_api_key")
    if err != nil {
        panic(err)
    }

    // Generate
    params := svc.Params().
        Model("gpt-4o").
        Messages(svc.UserMsg().Text("什么是太阳"))

    resp, err := svc.Gen(ctx, params, nil)
    if err != nil {
        panic(err)
    }

    if resp.Len() > 0 && resp.At(0).Parts() > 0 {
        fmt.Println(resp.At(0).Part(0).Text())
    }
}
```

## Service Creation

### URI Format

Both `New` and `NewV1` accept a URI in the form:

```
scheme:base=url&key=api_key&org=org_id&project=project_id&webhook_secret=secret
```

| Parameter       | Description                          |
|----------------|--------------------------------------|
| `base`         | API base URL (e.g. `https://api.openai.com/v1/`) |
| `key`          | API key for authentication           |
| `org`          | Organization ID (optional)           |
| `project`      | Project ID (optional)                 |
| `webhook_secret` | Webhook validation secret (optional) |

### v3 Responses API (default)

```go
svc, err := openai.New(ctx, "openai:base=https://api.openai.com/v1/&key=sk-xxx")
```

Uses `openai` scheme. Registers automatically with `xai.Register("openai", ...)`.

### v1 Chat Completions API

```go
svc, err := openai.NewV1(ctx, "openai-v1:base=https://api.openai.com/v1/&key=sk-xxx")
```

Uses `openai-v1` scheme. Registers automatically with `xai.Register("openai-v1", ...)`.

### Via xai.New (after Register)

```go
svc, err := xai.New(ctx, "openai:base=...&key=...")
svc, err := xai.New(ctx, "openai-v1:base=...&key=...")
```

## API Reference

### Gen / GenStream

```go
// Non-streaming
resp, err := svc.Gen(ctx, params, opts)

// Streaming (iter.Seq2[xai.GenResponse, error])
for resp, err := range svc.GenStream(ctx, params, opts) {
    if err != nil { /* handle */ }
    if resp != nil && resp.Len() > 0 {
        cand := resp.At(0)
        for i := 0; i < cand.Parts(); i++ {
            fmt.Print(cand.Part(i).Text())
        }
    }
}
```

### Params

| Method | Description |
|--------|-------------|
| `Model(model)` | Model ID (e.g. `gpt-4o`, `o1`) |
| `System(texts)` | System prompt via `svc.Texts("...")` |
| `Messages(msgs...)` | Conversation messages |
| `Tools(tools...)` | Function tools |
| `MaxOutputTokens(n)` | Max tokens to generate |
| `Temperature(v)` | Sampling temperature (0–2) |
| `TopP(v)` | Nucleus sampling |

### Messages (MsgBuilder)

Standard `xai.MsgBuilder` methods:

| Method | Description |
|--------|-------------|
| `Text(text)` | Plain text |
| `ImageURL(mime, url)` | Image by URL |
| `ImageFile(mime, fileID)` | Image by file ID |
| `DocURL(mime, url)` | Document by URL |
| `DocFile(mime, fileID)` | Document by file ID |
| `Part(part)` | Add a Part (from prior response) |
| `Thinking(v)` | Thinking/reasoning block |
| `ToolUse(v)` | Tool call |
| `ToolResult(v)` | Tool result |
| `Compaction(data)` | Compaction block |

```go
msg := svc.UserMsg().
    Text("Describe this image").
    ImageURL(xai.ImageJPEG, "https://example.com/image.jpg")
```

### MsgBuilderExt (provider-specific)

Use `UserMsgExt()` when you need image detail or video:

| Method | Description |
|--------|-------------|
| `TextExt(text)` | Text (returns MsgBuilderExt for chaining) |
| `ImageURLExt(mime, url)` | Image by URL |
| `ImageURLWithDetail(url, detail)` | Image with detail level (`low`, `medium`, `high`, `auto`, `ultra_high`) |
| `VideoFile(fileID, format)` | Video by file ID (e.g. `qfile-xxx`, MIME `video/mp4`) |

```go
msg := svc.UserMsgExt().
    TextExt("这是什么").
    ImageURLWithDetail("https://example.com/img.jpg", "ultra_high")

msg := svc.UserMsgExt().
    TextExt("视频里是什么").
    VideoFile("qfile-xxx", "video/mp4")
```

**Important**: Chain with `TextExt`/`ImageURLExt` (not `Text`/`ImageURL`) to keep `MsgBuilderExt` for `ImageURLWithDetail`/`VideoFile`.

### Tools

```go
// Define a tool
svc.ToolDef("get_weather").Description("Get weather for a city")

// Use in params
params := svc.Params().
    Model("gpt-4o").
    Messages(msgs...).
    Tools(svc.Tool("get_weather"))
```

### Web Search Tool

```go
params := svc.Params().
    Model("gpt-4o").
    Messages(msgs...).
    Tools(svc.WebSearchTool())
```

### Images / Docs Builders

```go
// From file
img, err := svc.Images().FromLocal(xai.ImagePNG, "photo.png")

// From bytes
img := svc.Images().FromBytes(xai.ImageJPEG, "photo.jpg", data)

// From base64
img, err := svc.Images().FromBase64(xai.ImageJPEG, "photo.jpg", base64Str)

// Document
doc, err := svc.Docs().FromLocal(xai.DocPDF, "doc.pdf")
doc := svc.Docs().PlainText("raw text")
```

### Response Handling

```go
resp, err := svc.Gen(ctx, params, nil)

// Iterate candidates (usually 1)
for i := 0; i < resp.Len(); i++ {
    cand := resp.At(i)
    fmt.Println("StopReason:", cand.StopReason())

    // Iterate parts (text, tool_use, reasoning, etc.)
    for j := 0; j < cand.Parts(); j++ {
        part := cand.Part(j)
        if text := part.Text(); text != "" {
            fmt.Print(text)
        }
        if thinking, ok := part.AsThinking(); ok {
            fmt.Println("Thinking:", thinking.Text)
        }
        if toolUse, ok := part.AsToolUse(); ok {
            fmt.Println("Tool:", toolUse.Name, toolUse.Input)
        }
    }
}
```

### Options

```go
opts := svc.Options().WithBaseURL("https://custom.endpoint/v1/")
resp, err := svc.Gen(ctx, params, opts)
```

## v3 vs v1

| Aspect | v3 (Responses API) | v1 (Chat Completions) |
|--------|--------------------|------------------------|
| Scheme | `openai` | `openai-v1` |
| Endpoint | `/v1/responses` | `/v1/chat/completions` |
| Features | Reasoning, compaction, richer output | Simpler, widely compatible |
| Streaming | SSE `response.output_text.delta` | SSE `choices[0].delta.content` |

Use v3 for models like o1 (reasoning) and full Responses API features. Use v1 for maximum compatibility with OpenAI-compatible proxies.

## Qiniu Provider

The [provider/qiniu](provider/qiniu/) package provides a Qiniu-backed service (api.qnaigc.com):

```go
import (
    "github.com/goplus/xai/spec/openai/provider/qiniu"
)

// Create service (uses QINIU_API_KEY env when token is empty)
svc := qiniu.NewService("your_token")

// Or with options
svc := qiniu.NewService("token", qiniu.WithBaseURL(qiniu.OverseasBaseURL))

// Register for xai.New("qiniu:")
qiniu.Register("your_token")
svc, _ := xai.New(ctx, "qiniu:")
```

Qiniu supports `ImageURLWithDetail` (including `ultra_high`) and `VideoFile` (qfile-xxx).

## Package Structure

```
spec/openai/
├── README.md           # This file
├── openai.go           # Service, New, NewV1, URI parsing
├── provider.go         # Internal provider interface, genRequest
├── provider_v3.go      # v3 Responses API implementation
├── provider_v1.go     # v1 Chat Completions implementation
├── params.go           # ParamBuilder
├── options.go          # OptionBuilder
├── message.go          # MsgBuilder, MsgBuilderExt
├── response.go         # Response types, contentBlock, buildOutputToInput
├── tool.go             # Tool, ToolDef, WebSearchTool
├── data.go             # ImageBuilder, DocumentBuilder
├── schema.go           # Image/Video builders (unsupported stubs)
├── operation.go        # Actions, Operation (todo)
└── provider/
    └── qiniu/
        └── service.go  # Qiniu-backed Service
```

## Examples

See [examples/openai](../../examples/openai/) in the parent repo:

```bash
# Text-only
go run ./examples/openai text

# Image + text
go run ./examples/openai image

# Image with detail (low / ultra_high)
go run ./examples/openai image-detail
go run ./examples/openai image-ultra

# Video
go run ./examples/openai video
go run ./examples/openai video-fileid
go run ./examples/openai multi-video

# Streaming
STREAM=1 go run ./examples/openai text
```

Set `QINIU_API_KEY` for real API calls when using the Qiniu-backed examples.

## License

Apache-2.0. See [LICENSE](../../LICENSE) in the parent repo.
