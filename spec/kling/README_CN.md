# Kling 使用教程

[English](README.md) | **中文**

---

## 教程概述

本教程带你使用 `kling` 包进行 Kling AI 图像与视频生成。你将学习：

1. **快速开始** — 5 行代码完成首次文生图
2. **图像生成** — 文生图、图生图、模型选择
3. **视频生成** — 图生视频、文生视频、首尾帧
4. **参数与校验** — 必填/可选、Restrict、错误处理
5. **Provider 集成** — 将 Kling 接入你的应用

**设计原则**：仅使用 xai 概念——`Actions`、`Operation`、`Params`、`Call`、`Wait`、`Results`。Executor 与 Backend 为内部实现。

---

## 前置条件

- Go 1.21+
- 真实 API 调用需配置：`QINIU_API_KEY`（见 [provider/qiniu/example](provider/qiniu/example/)）

---

## 快速开始

**步骤 1.** 创建带 Executor 的 Service（来自 `provider/qiniu`）：

```go
import (
    "context"
    xai "github.com/goplus/xai/spec"
    "github.com/goplus/xai/spec/kling"
    "github.com/goplus/xai/spec/kling/provider/qiniu"
)

// 使用 mock 或真实 backend
imgExec := qiniu.NewImageGenExecutor(imgBackend)
vidExec := qiniu.NewVideoGenExecutor(vidBackend)
svc := kling.NewService(imgExec, vidExec)
```

**步骤 2.** 获取 Operation 并设置参数：

```go
op, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
op.Params().Set(kling.ParamPrompt, "a sunset over the ocean")
op.Params().Set(kling.ParamAspectRatio, "16:9")
```

**步骤 3.** 调用并获取结果：

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

**运行示例：**

```bash
go run ./spec/kling/provider/qiniu/example/ text2image
```

以下教程假设你已有 `svc`、`ctx := context.Background()`，以及 `xai`/`kling` 的 import。

---

## 图像生成

### 1.1 文生图

仅使用 `prompt` 调用 `GenImage`。所有图像模型均需 `prompt`。

```go
op, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
op.Params().Set(kling.ParamPrompt, "a cat sitting on a windowsill, soft lighting")
op.Params().Set(kling.ParamAspectRatio, "16:9")
op.Params().Set(kling.ParamN, 1)

results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
```

**常用参数：** `ParamPrompt`（必填）、`ParamAspectRatio`、`ParamN`。

### 1.2 图生图

添加 `reference_images` 实现风格迁移。支持 kling-v2、kling-v2-1、kling-image-o1。`reference_images` 可为单张 URL（string）或多张 URL 列表（`[]string`）。

```go
op, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
op.Params().Set(kling.ParamPrompt, "same style but with a mountain landscape")
op.Params().Set(kling.ParamReferenceImages, []string{"https://example.com/ref.png"})
op.Params().Set(kling.ParamAspectRatio, "16:9")

results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
```

### 1.3 模型选择

| 模型 | 用途 |
|------|------|
| kling-v1 | 基础文生图，单张参考图 |
| kling-v1-5 | + `image_reference`（subject/face） |
| kling-v2, kling-v2-new | + `reference_images`（列表） |
| kling-v2-1 | 图像+视频，reference_images |
| kling-image-o1 | + `resolution`（1K/2K/4K） |

```go
// 列出可用图像模型
for _, m := range kling.ImageModels() {
    fmt.Println(m)
}

// 查看模型支持的操作
actions := svc.Actions(xai.Model(kling.ModelKlingV21))
// [GenImage, GenVideo]
```

---

## 视频生成

### 2.1 图生视频

kling-v2-1 需要 `input_reference`（首帧）。可用 `mode`、`seconds`、`size` 控制。

```go
op, _ := svc.Operation(xai.Model(kling.ModelKlingV21Video), xai.GenVideo)
op.Params().Set(kling.ParamPrompt, "camera pans slowly to the right")
op.Params().Set(kling.ParamInputReference, "https://example.com/first-frame.png")
op.Params().Set(kling.ParamMode, kling.ModePro)
op.Params().Set(kling.ParamSeconds, kling.Seconds5)
op.Params().Set(kling.ParamSize, kling.Size1920x1080)

results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
```

### 2.2 文生视频

kling-v2-5-turbo 支持纯文本（无需 `input_reference`）：

```go
op, _ := svc.Operation(xai.Model(kling.ModelKlingV25Turbo), xai.GenVideo)
op.Params().Set(kling.ParamPrompt, "a hero enters the battlefield, dramatic lighting")
op.Params().Set(kling.ParamMode, kling.ModePro)
op.Params().Set(kling.ParamSeconds, kling.Seconds5)

results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
```

### 2.3 首尾帧视频

首帧 + 尾帧实现平滑过渡。**要求** `mode="pro"`，kling-v2-1 还需 `seconds="10"`。

```go
op, _ := svc.Operation(xai.Model(kling.ModelKlingV21Video), xai.GenVideo)
op.Params().Set(kling.ParamPrompt, "smooth transition from day to night")
op.Params().Set(kling.ParamInputReference, "https://example.com/first.png")
op.Params().Set(kling.ParamImageTail, "https://example.com/end.png")
op.Params().Set(kling.ParamMode, kling.ModePro)   // 首尾帧必需
op.Params().Set(kling.ParamSeconds, kling.Seconds10) // kling-v2-1 首尾帧必需

results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
```

### 2.4 多参视频

kling-video-o1 使用 `image_list` 和 `video_list`。结构说明：

- **image_list**：`[]kling.ImageInput`，每项含 `Image`（URL 必填）、`Type`（可选：`first_frame` 首帧、`end_frame` 尾帧）
- **video_list**：`[]kling.VideoRef`，每项含 `VideoURL`（必填）、`ReferType`（`feature` 或 `base`）、`KeepOriginalSound`（可选：`yes`/`no`）

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

## 参数与校验

### 3.1 必填与可选

| 操作 | 必填 | 可选 |
|------|------|------|
| GenImage | prompt | aspect_ratio, reference_images, n, ... |
| GenVideo (kling-v2-1) | prompt, input_reference | mode, seconds, size, image_tail, ... |
| GenVideo (kling-v2-5-turbo) | prompt | input_reference, mode, seconds, ... |

缺少必填参数会返回：
- `ErrPromptRequired` — prompt 为空或未设置
- `ErrInputReferenceRequired` — kling-v2-1 缺少 input_reference
- `ErrKeyframeModeRequired` — 设置了 image_tail 但 mode != "pro"
- `ErrKeyframeSecondsRequired` — kling-v2-1 首尾帧缺少 seconds="10"

### 3.2 取值限制 (Restrict)

使用 `op.InputSchema().Restrict(name)` 获取参数允许值：

```go
schema := op.InputSchema()
if r := schema.Restrict(kling.ParamAspectRatio); r != nil {
    // r.Limit 包含 ["1:1","16:9","9:16",...]
}
```

| 参数 | 允许值 |
|------|--------|
| aspect_ratio | 1:1, 16:9, 9:16, 4:3, 3:4, 3:2, 2:3, 21:9 |
| image_reference (kling-v1-5) | subject, face |
| resolution (kling-image-o1) | 1K, 2K, 4K |
| mode (视频) | std, pro |
| seconds (视频) | 5, 10 |
| size (视频) | `Size1920x1080`、`Size1280x720` 等 |
| sound (kling-v2-6+) | on, off |

### 3.3 Call 与 xai.Call

**两种调用方式：**

```go
// 1. Call + Wait（异步或自定义流程）
resp, err := op.Call(ctx, svc, svc.Options())
results, err := xai.Wait(ctx, svc, resp, nil)

// 2. xai.Call（一行搞定，阻塞直到完成）
results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
```

务必在调用**之前**设置参数。

---

## Provider 集成

要在应用中使用 Kling，需实现 `ImageBackend` 和 `VideoBackend`，再接入 Executor：

```go
// 你的应用实现 Backend（委托给 aiprovider.Router 等）
type myBackend struct { router *aiprovider.Router }

func (b *myBackend) SubmitImage(params kling.ImageParams) (qiniu.ImageSubmitResult, error) {
    // params 为强类型，如 *image.V21ImageParams
    res, err := b.router.SubmitImageGeneration(params.Model(), "qiniu", params)
    return qiniu.ImageSubmitResult{Sync: res.Sync, TaskID: res.ExternalTaskID, Images: res.Images}, err
}

// 接入
imgExec := qiniu.NewImageGenExecutor(backend)
vidExec := qiniu.NewVideoGenExecutor(backend)
svc := kling.NewService(imgExec, vidExec)
kling.Register(svc)
```

接口定义见 [provider/qiniu/backend.go](provider/qiniu/backend.go)。

---

## 运行示例

### 真实 Qnagic API

需配置 `QINIU_API_KEY`。可选：`QINIU_BASE_URL`。

```bash
go run ./spec/kling/provider/qiniu/example/ text2image
go run ./spec/kling/provider/qiniu/example/ img2video
go run ./spec/kling/provider/qiniu/example/   # 运行全部
```

### 完整示例目录

每个示例对应 [provider/qiniu/example/](provider/qiniu/example/) 中的可运行 `Run*` 函数：

| 效果 | 文件 | 说明 |
|------|------|------|
| text2image | [text2image.go](provider/qiniu/example/text2image.go) | 文生图 — 仅 prompt，kling-v2-1 |
| image2image | [image2image.go](provider/qiniu/example/image2image.go) | 图生图 — reference_images，kling-v2-1 |
| img2video | [img2video.go](provider/qiniu/example/img2video.go) | 图生视频 — input_reference，kling-v2-1 |
| keyframe | [keyframe.go](provider/qiniu/example/keyframe.go) | 首尾帧 — input_reference + image_tail，mode=pro，seconds=10 |
| multi_ref | [multi_ref.go](provider/qiniu/example/multi_ref.go) | 多参视频 — image_list，kling-video-o1 |
| text2video | [text2video.go](provider/qiniu/example/text2video.go) | 文生视频 — 仅 prompt，kling-v2-5-turbo |

共享：[service.go](provider/qiniu/example/service.go) — `newQnagicService()` 创建带 Qnagic backend 的 Service。

### 示例：text2image（完整代码）

```go
// 来自 provider/qiniu/example/text2image.go
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

### 单元示例（Mock）

```bash
go test ./spec/kling/provider/qiniu/... -v -run Example
```

---

## API 参考

### 模型常量

- 图像：`ModelKlingV1`、`ModelKlingV15`、`ModelKlingV2`、`ModelKlingV2New`、`ModelKlingV21`、`ModelKlingImageO1`
- 视频：`ModelKlingV21Video`、`ModelKlingV25Turbo`、`ModelKlingVideoO1`、`ModelKlingV26` ~ `ModelKlingV29`

### 辅助函数

- `IsImageModel(m string) bool`
- `IsVideoModel(m string) bool`
- `ImageModels() []string`
- `VideoModels() []string`
- `SchemaForImage(model string) []xai.Field`
- `SchemaForVideo(model string) []xai.Field`

### Service

- `NewService(imgExec, vidExec, opts ...Option) *Service`
- `Register(svc *Service)`

### Options（xai.OptionBuilder）

- `WithBaseURL(base string)` — 预留，qiniu provider 当前通过 `QINIU_BASE_URL` 环境变量覆盖
- `WithUserID(userID string)` — 用于 Executor 的 API key 解析

### 参数常量

`ParamPrompt`、`ParamAspectRatio`、`ParamReferenceImages`、`ParamInputReference`、`ParamImageTail`、`ParamMode`、`ParamSeconds`、`ParamSize`、`ParamResolution`、`ParamSound` 等 — 见 [params.go](params.go)。

### 错误

- `ErrPromptRequired`
- `ErrInputReferenceRequired`
- `ErrKeyframeModeRequired`
- `ErrKeyframeSecondsRequired`
- `qiniu.ErrTaskFailed` — 异步任务失败时返回，可用 `errors.Is(err, qiniu.ErrTaskFailed)` 判断

---

## 包结构

```
spec/kling/
  kling.go      # Register, Scheme, Build*, IsImageModel, SchemaFor*, 模型常量
  params.go     # Params, 参数常量
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
