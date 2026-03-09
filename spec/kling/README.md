# Kling Tutorial

**English** | [中文](README_CN.md)

---

## Tutorial Overview

This tutorial walks you through using the `kling` package for Kling AI image and video generation. You will learn:

1. **Quick Start** — Run your first text-to-image in 5 lines
2. **Image Generation** — Text2Image, Image2Image, model selection
3. **Video Generation** — Image2Video, Text2Video, Keyframe
4. **Parameters & Validation** — Required/optional params, Restrict, errors
5. **Provider Integration** — Wire Kling into your application

**Design principle**: Use only xai concepts—`Actions`, `Operation`, `Params`, `Call`, `Wait`, `Results`. Executor and Backend are internal.

---

## Prerequisites

- Go 1.21+
- For real API calls: `QINIU_API_KEY` (see [provider/qiniu/example](provider/qiniu/example/))

---

## Quick Start

**Step 1.** Create a Service with Executors (from `provider/qiniu`):

```go
import (
    "context"
    xai "github.com/goplus/xai/spec"
    "github.com/goplus/xai/spec/kling"
    "github.com/goplus/xai/spec/kling/provider/qiniu"
)

// Use mock or real backend
imgExec := qiniu.NewImageGenExecutor(imgBackend)
vidExec := qiniu.NewVideoGenExecutor(vidBackend)
svc := kling.NewService(imgExec, vidExec)
```

**Step 2.** Get an Operation and set params:

```go
op, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
op.Params().Set(kling.ParamPrompt, "a sunset over the ocean")
op.Params().Set(kling.ParamAspectRatio, "16:9")
```

**Step 3.** Call and get results:

```go
ctx := context.Background()
results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
if err != nil {
    log.Fatal(err)
}
for i := 0; i < results.Len(); i++ {
    out := results.At(i).(*xai.OutputImage)
    fmt.Println(out.Image.StgUri())
}
```

**Run the example:**

```bash
go run ./spec/kling/provider/qiniu/example/ text2image
```

In the following tutorials, we assume you have `svc`, `ctx := context.Background()`, and the `xai`/`kling` imports.

---

## Tutorial 1: Image Generation

### 1.1 Text-to-Image

Use `GenImage` with `prompt` only. `prompt` is required for all image models.

```go
op, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
op.Params().Set(kling.ParamPrompt, "a cat sitting on a windowsill, soft lighting")
op.Params().Set(kling.ParamAspectRatio, "16:9")
op.Params().Set(kling.ParamN, 1)

results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
```

**Common params:** `ParamPrompt` (required), `ParamAspectRatio`, `ParamN`.

### 1.2 Image-to-Image

Add `reference_images` for style transfer. Supported by kling-v2, kling-v2-1, kling-image-o1. `reference_images` accepts a single URL (string) or a list of URLs (`[]string`).

```go
op, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
op.Params().Set(kling.ParamPrompt, "same style but with a mountain landscape")
op.Params().Set(kling.ParamReferenceImages, []string{"https://example.com/ref.png"})
op.Params().Set(kling.ParamAspectRatio, "16:9")

results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
```

### 1.3 Model Selection

| Model | Use Case |
|-------|----------|
| kling-v1 | Basic text2image, single ref image |
| kling-v1-5 | + `image_reference` (subject/face) |
| kling-v2, kling-v2-new | + `reference_images` (list) |
| kling-v2-1 | Image + video, reference_images |
| kling-image-o1 | + `resolution` (1K/2K/4K) |

```go
// List available image models
for _, m := range kling.ImageModels() {
    fmt.Println(m)
}

// Check actions for a model
actions := svc.Actions(xai.Model(kling.ModelKlingV21))
// [GenImage, GenVideo]
```

---

## Tutorial 2: Video Generation

### 2.1 Image-to-Video

kling-v2-1 requires `input_reference` (first frame). Use `mode`, `seconds`, `size` for control.

```go
op, _ := svc.Operation(xai.Model(kling.ModelKlingV21Video), xai.GenVideo)
op.Params().Set(kling.ParamPrompt, "camera pans slowly to the right")
op.Params().Set(kling.ParamInputReference, "https://example.com/first-frame.png")
op.Params().Set(kling.ParamMode, kling.ModePro)
op.Params().Set(kling.ParamSeconds, kling.Seconds5)
op.Params().Set(kling.ParamSize, kling.Size1920x1080)

results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
```

### 2.2 Text-to-Video

kling-v2-5-turbo supports text-only (no `input_reference`):

```go
op, _ := svc.Operation(xai.Model(kling.ModelKlingV25Turbo), xai.GenVideo)
op.Params().Set(kling.ParamPrompt, "a hero enters the battlefield, dramatic lighting")
op.Params().Set(kling.ParamMode, kling.ModePro)
op.Params().Set(kling.ParamSeconds, kling.Seconds5)

results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
```

### 2.3 Keyframe Video

First frame + end frame for smooth transition. **Requires** `mode="pro"` and for kling-v2-1: `seconds="10"`.

```go
op, _ := svc.Operation(xai.Model(kling.ModelKlingV21Video), xai.GenVideo)
op.Params().Set(kling.ParamPrompt, "smooth transition from day to night")
op.Params().Set(kling.ParamInputReference, "https://example.com/first.png")
op.Params().Set(kling.ParamImageTail, "https://example.com/end.png")
op.Params().Set(kling.ParamMode, kling.ModePro)   // required for keyframe
op.Params().Set(kling.ParamSeconds, kling.Seconds10) // required for kling-v2-1 keyframe

results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
```

### 2.4 Multi-Reference

kling-video-o1 uses `image_list` and `video_list`. Structure:

- **image_list**: `[]kling.ImageInput`, each item has `Image` (URL, required), `Type` (optional: `first_frame`, `end_frame`)
- **video_list**: `[]kling.VideoRef`, each item has `VideoURL` (required), `ReferType` (`feature` or `base`), `KeepOriginalSound` (optional: `yes`/`no`)

```go
op, _ := svc.Operation(xai.Model(kling.ModelKlingVideoO1), xai.GenVideo)
op.Params().Set(kling.ParamImageList, []kling.ImageInput{
    {Image: "https://example.com/ref1.png"},
    {Image: "https://example.com/ref2.png", Type: "first_frame"},
})
op.Params().Set(kling.ParamPrompt, "blend the styles")


results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
```

---

## Tutorial 3: Parameters & Validation

### 3.1 Required vs Optional

| Action | Required | Optional |
|--------|----------|----------|
| GenImage | prompt | aspect_ratio, reference_images, n, ... |
| GenVideo (kling-v2-1) | prompt, input_reference | mode, seconds, size, image_tail, ... |
| GenVideo (kling-v2-5-turbo) | prompt | input_reference, mode, seconds, ... |

Missing required params returns:
- `ErrPromptRequired` — prompt empty or not set
- `ErrInputReferenceRequired` — kling-v2-1 without input_reference
- `ErrKeyframeModeRequired` — image_tail set but mode != "pro"
- `ErrKeyframeSecondsRequired` — kling-v2-1 keyframe without seconds="10"

### 3.2 Value Limits (Restrict)

Use `op.InputSchema().Restrict(name)` to get allowed values for a param:

```go
schema := op.InputSchema()
if r := schema.Restrict(kling.ParamAspectRatio); r != nil {
    // r.Limit has ["1:1","16:9","9:16",...]
}
```

| Param | Allowed Values |
|-------|----------------|
| aspect_ratio | 1:1, 16:9, 9:16, 4:3, 3:4, 3:2, 2:3, 21:9 |
| image_reference (kling-v1-5) | subject, face |
| resolution (kling-image-o1) | 1K, 2K, 4K |
| mode (video) | std, pro |
| seconds (video) | 5, 10 |
| size (video) | `Size1920x1080`, `Size1280x720`, ... |
| sound (kling-v2-6+) | on, off |

### 3.3 Call vs xai.Call

**Two ways to invoke:**

```go
// 1. Call + Wait (for async or custom flow)
resp, err := op.Call(ctx, svc, svc.Options())
results, err := xai.Wait(ctx, svc, resp, nil)

// 2. xai.Call (one-liner, blocks until done)
results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
```

Always set params **before** calling.

---

## Tutorial 4: Provider Integration

To use Kling in your app, implement `ImageBackend` and `VideoBackend`, then wire Executors:

```go
// Your app implements Backend (delegates to aiprovider.Router etc.)
type myBackend struct { router *aiprovider.Router }

func (b *myBackend) SubmitImage(params kling.ImageParams) (qiniu.ImageSubmitResult, error) {
    // params is typed, e.g. *image.V21ImageParams
    res, err := b.router.SubmitImageGeneration(params.Model(), "qiniu", params)
    return qiniu.ImageSubmitResult{Sync: res.Sync, TaskID: res.ExternalTaskID, Images: res.Images}, err
}

// Wire up
imgExec := qiniu.NewImageGenExecutor(backend)
vidExec := qiniu.NewVideoGenExecutor(backend)
svc := kling.NewService(imgExec, vidExec)
kling.Register(svc)
```

See [provider/qiniu/backend.go](provider/qiniu/backend.go) for interface definitions.

---

## Run Examples

### Real Qnagic API

Requires `QINIU_API_KEY` (except `models`). Optional: `QINIU_BASE_URL`.

```bash
go run ./spec/kling/provider/qiniu/example/ models     # no API key needed
go run ./spec/kling/provider/qiniu/example/ text2image
go run ./spec/kling/provider/qiniu/example/ img2video
go run ./spec/kling/provider/qiniu/example/            # run all
```

### Complete Example Index

Each example is a runnable `Run*` function in [provider/qiniu/example/](provider/qiniu/example/):

| Effect | File | Description |
|--------|------|-------------|
| text2image | [text2image.go](provider/qiniu/example/text2image.go) | Text2Image — prompt only, kling-v2-1 |
| image2image | [image2image.go](provider/qiniu/example/image2image.go) | Image2Image — reference_images, kling-v2-1 |
| img2video | [img2video.go](provider/qiniu/example/img2video.go) | Image2Video — input_reference, kling-v2-1 |
| keyframe | [keyframe.go](provider/qiniu/example/keyframe.go) | Keyframe — input_reference + image_tail, mode=pro, seconds=10 |
| multi_ref | [multi_ref.go](provider/qiniu/example/multi_ref.go) | Multi-ref — image_list, kling-video-o1 |
| text2video | [text2video.go](provider/qiniu/example/text2video.go) | Text2Video — prompt only, kling-v2-5-turbo |

Shared: [service.go](provider/qiniu/example/service.go) — `newQnagicService()` creates Service with Qnagic backend.

### Example: text2image (Full Code)

```go
// From provider/qiniu/example/text2image.go
func RunText2Image() {
    svc, err := newQnagicService()
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    op, err := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    op.Params().Set(kling.ParamPrompt, "a sunset over the ocean, cinematic lighting")
    op.Params().Set(kling.ParamAspectRatio, "16:9")
    op.Params().Set(kling.ParamN, 1)

    ctx := context.Background()
    results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    for i := 0; i < results.Len(); i++ {
        out := results.At(i).(*xai.OutputImage)
        fmt.Println(out.Image.StgUri())
    }
}
```

### Unit Examples (Mock)

```bash
go test ./spec/kling/provider/qiniu/... -v -run Example
```

---

## API Reference

### Model Constants

- Image: `ModelKlingV1`, `ModelKlingV15`, `ModelKlingV2`, `ModelKlingV2New`, `ModelKlingV21`, `ModelKlingImageO1`
- Video: `ModelKlingV21Video`, `ModelKlingV25Turbo`, `ModelKlingVideoO1`, `ModelKlingV26` ~ `ModelKlingV29`

### Helpers

- `IsImageModel(m string) bool`
- `IsVideoModel(m string) bool`
- `ImageModels() []string`
- `VideoModels() []string`
- `SchemaForImage(model string) []xai.Field`
- `SchemaForVideo(model string) []xai.Field`

### Service

- `NewService(imgExec, vidExec, opts ...Option) *Service`
- `Register(svc *Service)`

### Options (xai.OptionBuilder)

- `WithBaseURL(base string)` — reserved; qiniu provider uses `QINIU_BASE_URL` env for override
- `WithUserID(userID string)` — for Executor API key resolution

### Param Constants

`ParamPrompt`, `ParamAspectRatio`, `ParamReferenceImages`, `ParamInputReference`, `ParamImageTail`, `ParamMode`, `ParamSeconds`, `ParamSize`, `ParamResolution`, `ParamSound`, etc. — see [params.go](params.go).

### Errors

- `ErrPromptRequired`
- `ErrInputReferenceRequired`
- `ErrKeyframeModeRequired`
- `ErrKeyframeSecondsRequired`
- `qiniu.ErrTaskFailed` — returned when async task fails; use `errors.Is(err, qiniu.ErrTaskFailed)` to detect

---

## Package Structure

```
spec/kling/
  kling.go      # Register, Scheme, Build*, IsImageModel, SchemaFor*, model constants
  params.go     # Params, param constants
  operation.go  # genImage, genVideo, inputSchema
  service.go    # Service, ImageGenExecutor, VideoGenExecutor, Options
  results.go    # NewOutputImages, NewOutputVideos, Sync/Async OperationResponse
  internal/     # param.go, model.go, params_reader.go, params_helpers.go
  image/        # ImageParams, BuildImageParams, IsImageModel, SchemaForImage, NewOutputImages
  video/        # VideoParams, BuildVideoParams, IsVideoModel, SchemaForVideo, Validate, NewOutputVideos
  provider/qiniu/
    backend.go, executor.go
    example/        # go run
    example_test.go # go test -run Example
```
