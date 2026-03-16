# Qiniu Vidu Provider

Qiniu Cloud Vidu API backend implementation, supporting `vidu-q1`, `vidu-q2`, `viduq2-turbo` and `viduq2-pro` video generation.

## Overview

This package implements `vidu.Backend`, providing Vidu video generation capabilities via Qiniu Cloud API.

Supported routes:

- `q1/text-to-video`
- `q1/reference-to-video` (`reference_image_urls` or `subjects`)
- `q2/text-to-video`
- `q2/reference-to-video` (`reference_image_urls` or `subjects`)
- `q2/image-to-video/pro`
- `q2/image-to-video/turbo`
- `q2/start-end-to-video/pro`
- `q2/start-end-to-video/turbo`

## Quick Start

```go
import (
    "context"

    xai "github.com/goplus/xai/spec"
    "github.com/goplus/xai/spec/vidu"
    "github.com/goplus/xai/spec/vidu/provider/qiniu"
)

func run(ctx context.Context, apiKey string) error {
    svc := qiniu.NewService(apiKey)

    op, err := svc.Operation(xai.Model(vidu.ModelViduQ2), xai.GenVideo)
    if err != nil {
        return err
    }

    op.Params().
        Set(vidu.ParamPrompt, "A woman walking through a vibrant city street at night.").
        Set(vidu.ParamImageURL, "https://example.com/ref.jpg").
        Set(vidu.ParamDuration, 5).
        Set(vidu.ParamResolution, vidu.Resolution720p).
        Set(vidu.ParamAudio, true).
        Set(vidu.ParamVoiceID, "voice-city-guide")

    _, err = xai.Call(ctx, svc, op, svc.Options(), nil)
    return err
}
```

## Examples

Runnable demos (set `QINIU_API_KEY` for real API; otherwise mock):

```bash
go run ./examples/vidu/video              # list demos
go run ./examples/vidu/video q1-text       # Q1 text-to-video
go run ./examples/vidu/video q1-ref-urls  # Q1 reference-to-video (URLs)
go run ./examples/vidu/video q1-ref-subjects
go run ./examples/vidu/video q1-ref-subjects-audio
go run ./examples/vidu/video q2-text       # Q2 text-to-video
go run ./examples/vidu/video q2-ref-urls
go run ./examples/vidu/video q2-ref-subjects
go run ./examples/vidu/video q2-image-pro  # Q2 image-to-video-pro
go run ./examples/vidu/video q2-image-pro-audio
go run ./examples/vidu/video q2-image-turbo
go run ./examples/vidu/video q2-start-end-pro
go run ./examples/vidu/video call-sync     # CallSync + GetTask resume
```

## Configuration Options

Pass options to `NewService`:

```go
svc := qiniu.NewService(apiKey,
    qiniu.WithBaseURL(qiniu.OverseasBaseURL),  // Overseas endpoint
    qiniu.WithRetry(3, time.Second),           // Retry with backoff
    qiniu.WithDebugLog(true),                  // Print curl and response logs
)
```

## File Structure

```
spec/vidu/provider/qiniu/
├── backend.go     # vidu.Backend implementation (Submit, GetTaskStatus, request/response)
├── client.go      # HTTP client
├── service.go     # NewService, Register
├── qiniu_test.go  # Unit tests
├── vidu_video.md  # Video API detailed documentation
└── README.md      # This file
```

## Detailed Documentation

- [vidu_video.md](./vidu_video.md)
- [Qiniu Qnagic API](https://apidocs.qnaigc.com)
