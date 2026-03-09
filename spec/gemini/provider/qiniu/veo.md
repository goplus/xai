# Veo 视频模型（Qiniu Provider）API 参考

本文档基于七牛云 Qnagic 接口整理，覆盖 `spec/gemini/provider/qiniu` 对 Veo 视频生成能力的接入方式。

## 1. 概述

| 项目 | 说明 |
| --- | --- |
| 创建任务 | `POST /v1/videos/generations` |
| 查询任务 | `GET /v1/videos/generations/{id}` |
| 认证方式 | `Authorization: Bearer <token>` |
| 基础路径 | `/v1` |

## 2. 模型清单

以下 Veo 视频模型按同一套接口使用，能力差异见下表：

| 模型 | 时长支持 | 参考图片 | 分辨率 | resizeMode |
| --- | --- | --- | --- | --- |
| veo-2.0-generate-001 | 5–8 秒 | - | - | - |
| veo-2.0-generate-exp | 5–8 秒 | 支持 | - | - |
| veo-2.0-generate-preview | 5–8 秒 | - | - | - |
| veo-3.0-generate-preview | 4、6、8 秒 | - | 720p/1080p | pad/crop |
| veo-3.0-fast-generate-preview | 4、6、8 秒 | - | 720p/1080p | pad/crop |
| veo-3.1-generate-preview | 4、6、8 秒 | 支持 | 720p/1080p | pad/crop |
| veo-3.1-fast-generate-preview | 4、6、8 秒 | - | 720p/1080p | pad/crop |

## 3. 创建视频任务（`POST /v1/videos/generations`）

### 3.1 文生视频（带音频）

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "veo-3.0-generate-preview",
  "instances": [
    {
      "prompt": "A golden retriever running on a beach at sunset, cinematic lighting, slow motion"
    }
  ],
  "parameters": {
    "aspectRatio": "16:9",
    "durationSeconds": 8,
    "sampleCount": 1,
    "seed": 42,
    "negativePrompt": "blurry, low quality",
    "personGeneration": "dont_allow",
    "generateAudio": true
  }
}'
```

### 3.2 图生视频

> 说明：Go 示例见 `examples/veo/veo_image_to_video.go`，使用 `DemoURLs.Image`（URL）。

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "veo-3.0-generate-preview",
  "instances": [
    {
      "prompt": "A gentle breeze blowing through the scene",
      "image": {
        "bytesBase64Encoded": "<base64_encoded_image_data>",
        "mimeType": "image/png"
      }
    }
  ],
  "parameters": {
    "aspectRatio": "16:9",
    "durationSeconds": 6,
    "sampleCount": 1,
    "seed": 100,
    "negativePrompt": "",
    "personGeneration": "dont_allow"
  }
}'
```

### 3.3 回调模式

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "veo-3.0-generate-preview",
  "instances": [
    {
      "prompt": "A cat playing with a ball of yarn"
    }
  ],
  "parameters": {
    "aspectRatio": "16:9",
    "durationSeconds": 8,
    "sampleCount": 1,
    "seed": 0,
    "negativePrompt": "",
    "personGeneration": "dont_allow"
  },
  "callback_url": "https://your-server.com/api/veo-callback"
}'
```

### 3.4 首尾帧（image + lastFrame）

指定首帧（`image`）和尾帧（`lastFrame`），生成从首帧过渡到尾帧的视频。参考 [创建视频生成任务](https://apidocs.qnaigc.com/423900623e0)。

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "veo-3.0-generate-preview",
  "instances": [
    {
      "prompt": "Smooth transition from day to night",
      "image": {
        "bytesBase64Encoded": "<base64_encoded_first_frame>",
        "mimeType": "image/png"
      },
      "lastFrame": {
        "bytesBase64Encoded": "<base64_encoded_last_frame>",
        "mimeType": "image/png"
      }
    }
  ],
  "parameters": {
    "aspectRatio": "16:9",
    "durationSeconds": 6,
    "sampleCount": 1,
    "seed": 100,
    "personGeneration": "dont_allow"
  }
}'
```

### 3.4.1 引用视频（video）

提供已有视频（≤10MB）作为输入，模型可基于该视频进行延伸或修改。支持 `bytesBase64Encoded` 或 `uri`。

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "veo-3.0-generate-preview",
  "instances": [
    {
      "prompt": "Continue the scene with smooth motion",
      "video": {
        "bytesBase64Encoded": "<base64_encoded_video_data>",
        "mimeType": "video/mp4"
      }
    }
  ],
  "parameters": {
    "aspectRatio": "16:9",
    "durationSeconds": 6,
    "sampleCount": 1,
    "personGeneration": "dont_allow"
  }
}'
```

### 3.4.2 多参考图（referenceImages）

仅 veo-2.0-generate-exp、veo-3.1-generate-preview 支持。最多 3 张 asset 或 1 张 style；与 image 互斥；时长必须 8 秒。

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "veo-3.1-generate-preview",
  "instances": [
    {
      "prompt": "A cinematic scene with the characters walking through a garden",
      "referenceImages": [
        {
          "image": {"uri": "https://example.com/ref1.jpg", "mimeType": "image/jpeg"},
          "referenceType": "asset"
        },
        {
          "image": {"uri": "https://example.com/ref2.png", "mimeType": "image/png"},
          "referenceType": "asset"
        }
      ]
    }
  ],
  "parameters": {
    "aspectRatio": "16:9",
    "durationSeconds": 8,
    "sampleCount": 1,
    "personGeneration": "dont_allow"
  }
}'
```

### 3.5 提交响应示例

```json
{
  "id": "chatvideo-1709712000000000000-uid123"
}
```

### 3.6 请求参数完整说明

| 字段 | 位置 | 类型 | 必填 | 说明 | 约束 |
| --- | --- | --- | --- | --- | --- |
| `model` | body | string | 是 | 模型名称 | 仅支持 veo 系列 |
| `instances` | body | array | 是 | 生成实例 | 恰好 1 个元素 |
| `instances[0].prompt` | body | string | 文生视频必填 | 文本提示词 | - |
| `instances[0].image` | body | object | 否 | 图生视频输入 | 图片 ≤10MB；bytesBase64Encoded 或 uri |
| `instances[0].lastFrame` | body | object | 否 | 结束帧图片 | 图片 ≤10MB |
| `instances[0].video` | body | object | 否 | 视频输入 | 视频 ≤10MB |
| `instances[0].referenceImages` | body | array | 否 | 参考图片 | 仅 veo-2.0-generate-exp、veo-3.1-generate-preview；最多 3 张素材或 1 张风格；与 image 互斥；时长必须 8 秒 |
| `parameters.aspectRatio` | body | string | 否 | 宽高比 | `16:9`、`9:16` |
| `parameters.durationSeconds` | body | int64 | 否 | 视频时长（秒） | Veo 2：5–8；Veo 3：4、6、8；参考图片模式：仅 8 |
| `parameters.sampleCount` | body | int64 | 否 | 生成数量 | 1–4 |
| `parameters.seed` | body | int64 | 否 | 随机种子 | 0–4294967295 |
| `parameters.negativePrompt` | body | string | 否 | 负向提示词 | - |
| `parameters.personGeneration` | body | string | 否 | 人物生成策略 | `dont_allow`、`allow_adult` |
| `parameters.generateAudio` | body | bool | 否 | 是否生成音频 | 默认 false |
| `parameters.resolution` | body | string | 否 | 分辨率 | 仅 Veo 3：`720p`、`1080p` |
| `parameters.fps` | body | int32 | 否 | 帧率 | - |
| `parameters.enhancePrompt` | body | bool | 否 | 提示词增强 | 默认 false |
| `parameters.compressionQuality` | body | string | 否 | 压缩质量 | `high`、`medium`、`low` |
| `parameters.resizeMode` | body | string | 否 | 图生视频缩放模式 | 仅 Veo 3：`pad`、`crop` |
| `callback_url` | body | string | 否 | 终态回调 URL | 任务 Completed/Failed 时 POST；最多重试 3 次 |

## 4. 查询任务状态（`GET /v1/videos/generations/{id}`）

### 4.1 请求示例

```bash
curl --location --request GET 'https://api.qnaigc.com/v1/videos/generations/chatvideo-1709712000000000000-uid123' \
--header 'Authorization: Bearer <token>'
```

### 4.2 响应示例

```json
{
  "id": "chatvideo-1709712000000000000-uid123",
  "model": "veo-3.0-generate-preview",
  "status": "Completed",
  "message": "完成",
  "data": {
    "videos": [
      {
        "url": "https://cdn.example.com/videos/sample_0.mp4?token=xxx&e=1711904000",
        "mimeType": "video/mp4"
      }
    ],
    "raiMediaFilteredCount": 0
  },
  "created_at": "2025-03-06T12:00:00Z",
  "updated_at": "2025-03-06T12:05:30Z"
}
```

### 4.3 响应字段说明

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | string | 任务 ID，用于查询任务状态和结果 |
| `model` | string | 使用的模型名称 |
| `status` | string | 任务状态，见 4.4 |
| `message` | string | 任务状态的中文描述；失败时包含具体原因 |
| `data.videos` | array | 生成的视频列表；任务未完成时为空数组 |
| `data.videos[].url` | string | 视频下载链接，**有效期 14 天** |
| `data.videos[].mimeType` | string | 视频 MIME 类型，如 `video/mp4` |
| `data.raiMediaFilteredCount` | int32 | 因违反使用准则而被过滤的视频数量 |
| `created_at` | string | 任务创建时间（ISO 8601） |
| `updated_at` | string | 任务最后更新时间（ISO 8601） |

### 4.4 任务状态枚举

| 状态 | 说明 |
| --- | --- |
| `Initializing` | 初始化中 |
| `Queued` | 排队中 |
| `Running` | 运行中 |
| `Completed` | 已完成 |
| `Failed` | 失败 |
| `Uploading` | 上传中 |
| `Unknown` | 未知 |

## 5. 参数约束与错误码

### 5.1 输入约束

- **图片/视频大小**：单文件不超过 10MB
- **视频时长**：`durationSeconds` 必须为 4、5、6、8 之一（按模型区分，见 2. 模型清单）
- **参考图片模式**：使用 `referenceImages` 时，`durationSeconds` 必须为 8 秒，且不能同时设置 `image`

### 5.2 错误码

| HTTP 状态码 | 错误类型 | 说明 |
| --- | --- | --- |
| 400 | `invalid_request_error` | 请求参数错误（如 model 为空、durationSeconds 不合法、请求体格式错误） |
| 401 | `authentication_error` | 认证失败 |
| 403 | `quota_exceeded` | API Key 额度不足 |
| 500 | `view_vertex_video_job_failed` 等 | 服务器内部错误 |

### 5.3 典型错误示例

```json
{"error": {"message": "image size > 10MB", "type": "file_size_limit_exceeded"}}
{"error": {"message": "do not support durationSeconds not in [4, 6, 8]", "type": "invalid_request_error"}}
{"error": {"message": "model is empty", "type": "invalid_request_error"}}
{"error": {"message": "record not found", "type": "view_vertex_video_job_failed"}}
```

## 6. 与 spec/gemini 参数映射

`Operation(GenVideo).Params().Set(...)` 与 Qiniu Veo 请求的主要映射：

| `spec/gemini` 字段 | Qiniu Veo 字段 |
| --- | --- |
| `Prompt` | `instances[0].prompt` |
| `AspectRatio` | `parameters.aspectRatio` |
| `DurationSeconds` | `parameters.durationSeconds` |
| `NumberOfVideos` | `parameters.sampleCount` |
| `Seed` | `parameters.seed` |
| `NegativePrompt` | `parameters.negativePrompt` |
| `PersonGeneration` | `parameters.personGeneration` |
| `GenerateAudio` | `parameters.generateAudio` |
| `Resolution` | `parameters.resolution` |
| `FPS` | `parameters.fps` |
| `EnhancePrompt` | `parameters.enhancePrompt` |
| `CompressionQuality` | `parameters.compressionQuality` |
| `ResizeMode` | `parameters.resizeMode` |
| `PubsubTopic` | `callback_url` |
| `Image` | `instances[0].image`（首帧） |
| `LastFrame` | `instances[0].lastFrame`（尾帧） |
| `Video` | `instances[0].video`（引用视频输入） |
| `ReferenceImages` | `instances[0].referenceImages`（多参考图，仅 veo-2.0-generate-exp、veo-3.1-generate-preview；最多 3 张 asset 或 1 张 style；与 image 互斥；时长必须 8 秒） |

说明：

- **文生视频**：仅需 `Prompt`。**图生视频**：需 `Image` + `Prompt`。**首尾帧**：`Image`（首帧）+ `LastFrame`（尾帧）+ `Prompt`。**引用视频**：需 `Video` + `Prompt`。**多参考图**：需 `ReferenceImages` + `Prompt`，仅 veo-2.0-generate-exp、veo-3.1-generate-preview，最多 3 张 asset 或 1 张 style，时长 8 秒，见 `examples/veo/veo_reference_images.go`。图/视频输入均使用 URL。`Mask` 暂不支持。
- 除 `Prompt` 外，其余参数均为可选。设置时请使用 `gemini.Aspect16x9`、`gemini.Duration8` 等常量，避免硬编码非法值。
- `AspectRatio` 传入 `16:9 (landscape)` 或 `9:16 (portrait)` 时会自动归一化为 `16:9` / `9:16`。

## 7. 使用示例（xai Operation）

**必填**：`Prompt`。其余参数均为可选，设置时请使用 `gemini.*` 常量以保证合法值：

```go
import "github.com/goplus/xai/spec/gemini"

// 最简：仅 Prompt
op, _ := svc.Operation("veo-3.0-generate-preview", xai.GenVideo)
op.Params().Set("Prompt", "A cat playing with a ball of yarn")

// 可选参数（使用 gemini 常量）
op.Params().
    Set("Prompt", "A cat playing with a ball of yarn").
    Set("AspectRatio", gemini.Aspect16x9).            // 或 gemini.Aspect9x16
    Set("DurationSeconds", gemini.Duration8).         // Veo 3: Duration4/6/8; Veo 2: Duration5-8
    Set("Resolution", gemini.Res1080p).                // 或 gemini.Res720p（仅 Veo 3）
    Set("PersonGeneration", gemini.PersonDontAllow).  // 或 gemini.PersonAllowAdult
    Set("CompressionQuality", gemini.CompressionOptimized)

// 首尾帧：Image（首帧）+ LastFrame（尾帧）+ Prompt
op.Params().
    Set("Image", firstFrameImg).
    Set("LastFrame", lastFrameImg).
    Set("Prompt", "Smooth transition from day to night")

// 引用视频：Video + Prompt（视频 ≤10MB）
op.Params().
    Set("Video", inputVideo).
    Set("Prompt", "Continue the scene with smooth motion")

// 多参考图：ReferenceImages + Prompt（仅 veo-2.0-exp、veo-3.1，最多 3 asset 或 1 style，时长 8s）
refs := svc.GenVideoReferenceImages(
    xai.GenVideoReferenceImage{Image: img1, ReferenceType: "asset"},
    xai.GenVideoReferenceImage{Image: img2, ReferenceType: "asset"},
)
op.Params().Set("Prompt", "A cinematic scene").Set("ReferenceImages", refs).Set("DurationSeconds", 8)

resp, _ := op.Call(ctx, svc, nil)
for !resp.Done() {
    resp, _ = resp.Retry(ctx, svc)
}
video := resp.Results().At(0).(*xai.OutputVideo).Video
_ = video
```

可选常量定义见 `spec/gemini/veo_opt.go`。

## 8. 参考链接

- [Veo 创建任务（423900623e0）](https://apidocs.qnaigc.com/423900623e0)
- [Veo 查询任务（423900624e0）](https://apidocs.qnaigc.com/423900624e0)
- [Gemini 图像文档（gemini.md）](gemini.md)
