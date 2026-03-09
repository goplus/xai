# Sora Examples

Runnable Sora video generation demos using `spec/openai` with Qiniu OpenAI-compatible endpoints.

## Prerequisites

- Go `1.24+`
- `QINIU_API_KEY`

## Quick Start

```bash
# 1) set key
export QINIU_API_KEY=your-key

# 2) list demos
go run ./examples/sora

# 3) run a demo
go run ./examples/sora text-to-video

# 4) run all demos
go run ./examples/sora all
```

## Demos

### Text-to-Video

| Demo | Model | Description |
|------|-------|-------------|
| `text-to-video` | sora-2 | Basic text-to-video (4s, 1280x720) |
| `text-to-video-sora2-pro` | sora-2-pro | Pro model (4s, 1280x720) |
| `text-to-video-portrait` | sora-2 | Portrait (4s, 720x1280) |
| `text-to-video-8sec` | sora-2 | 8 second duration |
| `text-to-video-12sec` | sora-2 | 12 second duration |
| `text-to-video-sora2-pro-portrait` | sora-2-pro | Pro portrait (4s, 1024x1792) |

### Image-to-Video

| Demo | Model | Description |
|------|-------|-------------|
| `image-to-video` | sora-2 | Image-to-video with input_reference |
| `image-to-video-sora2-pro` | sora-2-pro | Pro model image-to-video |

### Remix

| Demo | Model | Description |
|------|-------|-------------|
| `remix` | sora-2 | Remix from existing video (requires SORA_SOURCE_VIDEO_ID) |
| `remix-sora2-pro` | sora-2-pro | Pro model remix |

### Remix Setup

```bash
export SORA_SOURCE_VIDEO_ID=qvideo-user123-1766453713089395279
go run ./examples/sora remix
```

## API Mapping

| Demo Type | Endpoint |
|-----------|----------|
| text-to-video | `POST /v1/videos` |
| image-to-video | `POST /v1/videos` with `input_reference` |
| remix | `POST /v1/videos/{id}/remix` |

All demos use:

- `Operation(xai.Model("sora-2")|"sora-2-pro", xai.GenVideo)`
- `xai.CallSync(...)`
- `xai.Wait(...)` polling

## Reference

- [spec/openai/provider/qiniu/sora.md](../../spec/openai/provider/qiniu/sora.md)
