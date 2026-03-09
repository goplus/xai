# Gemini Examples (Qiniu Provider)

This directory contains runnable Gemini demos through `spec/gemini/provider/qiniu`.

- Chat mode uses OpenAI-compatible endpoint: `/v1/chat/completions`
- Image operations use:
  - `/v1/images/generations`
  - `/v1/images/edits`
- `chat-*` demos复用 `examples/openai/shared` 的服务创建与流式输出逻辑

## Prerequisites

- Go `1.24.5+`
- `QINIU_API_KEY`
- Network access to Qiniu endpoints

## Quick Start

```bash
# 1) Set API key
export QINIU_API_KEY=your-key

# 2) List demos
go run ./examples/gemini

# 3) Run one demo
go run ./examples/gemini chat-text

# 4) Run multiple demos
go run ./examples/gemini chat-text chat-image image-generate
```

## Stream Mode

```bash
go run ./examples/gemini --stream chat-text
go run ./examples/gemini --no-stream chat-image

STREAM=1 go run ./examples/gemini chat-text
```

## Demo Matrix

- `chat-text`: text-only chat
- `chat-image`: text + image_url chat
- `chat-tool`: function calling round-trip
- `image-generate`: `Operation(xai.GenImage)`
- `image-edit`: `Operation(xai.EditImage)`

## Notes

- Recommended models:
  - `gemini-2.5-flash-image`
  - `gemini-3.0-pro-image-preview`
  - `gemini-3.1-flash-image-preview`
- `chat-tool` demo is executed in non-stream mode for clearer round-trip output.

## Directory Layout

```text
examples/gemini/
├── README.md
├── main.go
├── urls.go
├── chat_text.go
├── chat_image.go
├── chat_tool.go
├── images_common.go
├── images_generate.go
├── images_edit.go
└── shared/
    ├── blocks.go
    └── service.go
```
