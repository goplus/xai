# Gemini 图像模型（Qiniu Provider）API 参考

本文档基于七牛云 Qnagic OpenAI 兼容接口整理，覆盖 Gemini 图像模型在 Qiniu Provider 下的接入方式。

> Veo 视频模型文档请查看 [veo.md](veo.md)。

## 1. 概述

| 项目 | 说明 |
| --- | --- |
| 国内端点 | `https://api.qnaigc.com` |
| 海外端点 | `https://openai.sufy.com` |
| 认证方式 | `Authorization: Bearer <token>` |
| Content-Type | `application/json` |
| 基础路径 | `/v1` |

## 2. 模型清单

以下 Gemini 图像模型按同一套接口使用：

- `gemini-2.5-flash-image`
- `gemini-3.0-pro-image-preview`
- `gemini-3.1-flash-image-preview`

## 3. 接口能力矩阵

| 接口 | 用途 | 典型场景 |
| --- | --- | --- |
| `POST /v1/chat/completions` | 多模态对话生成图像 | 文生图、图生图、流式输出 |
| `POST /v1/images/generations` | 标准文生图 | 同步返回 base64 图片 |
| `POST /v1/images/edits` | 图像编辑/融合 | 风格转换、多图融合 |

> 说明：当前 `spec/gemini/provider/qiniu` 已实现双通路：  
> 1) `Gen/GenStream` 走 OpenAI 兼容 `chat/completions`；  
> 2) `Operation` 的 `GenImage/EditImage` 走 `images/generations` 与 `images/edits`。  
>
> provider 内的 `ImageFrom* / VideoFrom* / ReferenceImage` 等 objectFactory 能力直接委托给 `spec/gemini`，qiniu 仅负责协议与端点适配。

## 4. Chat Completions（`/v1/chat/completions`）

### 4.1 文生图（流式）

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/chat/completions' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "gemini-2.5-flash-image",
  "stream": true,
  "messages": [
    {
      "role": "user",
      "content": "画一只可爱的橘猫，坐在窗台上看着夕阳"
    }
  ]
}'
```

### 4.2 图生图（流式，多模态输入）

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/chat/completions' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "gemini-3.0-pro-image-preview",
  "stream": true,
  "messages": [
    {
      "role": "user",
      "content": [
        {
          "type": "text",
          "text": "将这张图片转换为水彩画风格"
        },
        {
          "type": "image_url",
          "image_url": {
            "url": "https://aitoken-public.qnaigc.com/example/generate-image/image-to-image-1.jpg"
          }
        }
      ]
    }
  ]
}'
```

### 4.3 非流式 + 比例控制

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/chat/completions' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "gemini-3.1-flash-image-preview",
  "stream": false,
  "messages": [
    {
      "role": "user",
      "content": "画一只可爱的橘猫"
    }
  ],
  "image_config": {
    "aspect_ratio": "16:9"
  }
}'
```

### 4.4 非流式响应样例（核心字段）

```json
{
  "id": "chatcmpl-2f8236e9f2b34bd391289576d0e23e72",
  "object": "chat.completion",
  "created": 1764574464,
  "model": "gemini-2.5-flash-image",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "",
        "reasoning_content": "Processing the image generation task...",
        "images": [
          {
            "type": "image_url",
            "image_url": {
              "url": "data:image/png;base64,iVBORw0KGgoAAA..."
            },
            "index": 0
          }
        ]
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 19,
    "completion_tokens": 1120,
    "total_tokens": 1139
  }
}
```

## 5. Images Generations（`/v1/images/generations`）

### 5.1 基础文生图

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "gemini-2.5-flash-image",
  "prompt": "梦幻森林中的精灵小屋，魔法光芒环绕",
  "temperature": 0.8,
  "top_p": 0.95
}'
```

### 5.2 文生图 + 画幅比例

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "gemini-3.0-pro-image-preview",
  "prompt": "一只可爱的橘猫坐在窗台上看着夕阳，照片风格，高清画质",
  "image_config": {
    "aspect_ratio": "16:9"
  }
}'
```

### 5.3 竖图比例

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "gemini-3.1-flash-image-preview",
  "prompt": "梦幻森林中的精灵小屋，魔法光芒环绕",
  "image_config": {
    "aspect_ratio": "9:16"
  }
}'
```

### 5.4 响应样例

```json
{
  "created": 1234567890,
  "data": [
    {
      "b64_json": "iVBORw0KGgoAAAANSUhEUgA..."
    }
  ],
  "output_format": "png",
  "usage": {
    "total_tokens": 4234,
    "input_tokens": 234,
    "output_tokens": 4000,
    "input_tokens_details": {
      "text_tokens": 234,
      "image_tokens": 0
    }
  }
}
```

## 6. Images Edits（`/v1/images/edits`）

### 6.1 Base64 图编辑

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/edits' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "gemini-2.5-flash-image",
  "image": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgA...",
  "prompt": "将这张图片转换为油画风格"
}'
```

### 6.2 多图融合

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/edits' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "gemini-3.0-pro-image-preview",
  "image": [
    "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
    "https://aitoken-public.qnaigc.com/example/generate-video/lawn.jpg"
  ],
  "prompt": "结合这两张图片的风格，生成一张新的艺术作品"
}'
```

### 6.3 URL 图编辑 + 比例

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/edits' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "model": "gemini-3.1-flash-image-preview",
  "image": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
  "prompt": "将这张照片转换为水彩画风格，保持主体清晰",
  "image_config": {
    "aspect_ratio": "16:9"
  }
}'
```

### 6.4 响应样例

```json
{
  "created": 1234567890,
  "data": [
    {
      "b64_json": "iVBORw0KGgoAAAANSUhEUgA..."
    }
  ],
  "output_format": "png",
  "usage": {
    "total_tokens": 5234,
    "input_tokens": 1234,
    "output_tokens": 4000,
    "input_tokens_details": {
      "text_tokens": 234,
      "image_tokens": 1000
    }
  }
}
```

## 7. Provider 设计（基于 spec/gemini Backend）

`spec/gemini/provider/qiniu` 不再重复定义一套 Gemini Service，而是实现 `spec/gemini.Backend` 并注入到 `gemini.NewWithBackend(...)`：

- `NewBackend(token, opts...) gemini.Backend`
- `NewService(token, opts...) *gemini.Service`
- `Register(token, opts...)`
- 支持 `WithBaseURL(...)`
- 支持 `WithHTTPClient(...)`
- 默认读取 `QINIU_API_KEY`
- 默认 base URL 为 `https://api.qnaigc.com/v1/`

能力分层：

- `Gen/GenStream`：通过 OpenAI 兼容 `/v1/chat/completions` 适配到 `genai.GenerateContent*`
- `Operation(GenImage/EditImage)`：通过 `/v1/images/generations` 与 `/v1/images/edits` 适配到 `genai.GenerateImagesResponse/EditImageResponse`
- objectFactory（`ImageFrom*`、`ReferenceImage` 等）全部复用 `spec/gemini`

图像操作参数（`Operation.Params().Set(...)`，按 `spec/gemini` 标准）：

- `GenImage`：`Prompt`、`AspectRatio`、`NumberOfImages`（来自 `genai.GenerateImagesConfig`）
- `EditImage`：`Prompt`、`References`、`AspectRatio`、`NumberOfImages`（来自 `genai.EditImageConfig`）

Veo 视频能力请查看 [veo.md](veo.md)。

使用示例：

```go
import (
  "context"

  xai "github.com/goplus/xai/spec"
  "github.com/goplus/xai/spec/gemini/provider/qiniu"
)

func main() {
  qiniu.Register("your_token")
  svc, _ := xai.New(context.Background(), "gemini-qiniu:")
  _ = svc
}
```

## 8. 参考链接

- [Gemini Chat Completions（397191373e0）](https://apidocs.qnaigc.com/397191373e0)
- [Gemini Images Generations（397191374e0）](https://apidocs.qnaigc.com/397191374e0)
- [Gemini Images Edits（397191375e0）](https://apidocs.qnaigc.com/397191375e0)
- [Veo 文档（veo.md）](veo.md)
