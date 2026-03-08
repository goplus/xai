# Kling 图像生成 API 参考

本文档基于 [七牛云 Qnagic API](https://apidocs.qnaigc.com) 整理，描述 Kling 图像生成接口。

---

## 1. 概述

| 项目 | 说明 |
|------|------|
| 国内端点 | `https://api.qnaigc.com` |
| 海外端点 | `https://openai.sufy.com` |
| 认证方式 | Bearer Token：`Authorization: Bearer <token>` |
| Content-Type | `application/json` |

---

## 2. 模型能力对比表

| 能力/参数 | kling-v1 | kling-v1-5 | kling-v2 | kling-v2-new | kling-v2-1 | kling-image-o1 |
|-----------|----------|------------|----------|--------------|------------|----------------|
| 文生图 | ✓ | ✓ | ✓ | - | ✓ | ✓ |
| 单图生图 | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| 多图生图 | - | - | ✓ (/edits) | - | ✓ (/edits) | - |
| negative_prompt | ✓ | ✓ | ✓ | ✓ | ✓ | - |
| image_reference (subject/face) | - | ✓ | - | - | - | - |
| image_fidelity / human_fidelity | ✓ | ✓ | - | - | - | - |
| subject_image_list | - | - | ✓ | - | ✓ | - |
| scene_image / style_image | - | - | ✓ | - | ✓ | - |
| image_urls (<<<image_1>>>) | - | - | - | - | - | ✓ |
| resolution (1K/2K) | - | - | - | - | - | ✓ |

---

## 3. 图像生成 API

### 3.1 kling-v1

**接口**：`POST /v1/images/generations`

**能力**：文生图、单图生图。返回任务 ID，需轮询查询结果。

**参考文档**：[创建文生图或单图生图任务 (kling-v1)](https://apidocs.qnaigc.com/396977229e0)

#### Request 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | ✓ | 固定为 `kling-v1` |
| prompt | string | ✓ | 图像生成的文本描述，≤2500 字符 |
| image | string | | 参考图 URL 或 Base64（单图生图时使用） |
| n | integer | | 生成图像数量，1–10，默认 1 |
| negative_prompt | string | | 负向提示词，≤2500 字符；图生图时不可用 |
| aspect_ratio | string | | 画面纵横比，默认 `16:9` |
| image_fidelity | float | | 参考强度 [0,1] |
| human_fidelity | float | | 面部参考强度 [0,1] |

#### Response 200（成功）

```json
{
  "task_id": "image-1762159125266058362-1383010xxx"
}
```

#### curl 示例

**文生图**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v1",
    "prompt": "一只可爱的橘猫坐在窗台上看着夕阳,照片风格,高清画质",
    "aspect_ratio": "16:9"
}'
```

**单图生图**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v1",
    "image": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
    "prompt": "将这张图片转换为水彩画风格",
    "aspect_ratio": "16:9"
}'
```

**文生图 + 负向提示词**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v1",
    "prompt": "一个美丽的花园",
    "negative_prompt": "模糊,低质量,变形",
    "aspect_ratio": "1:1"
}'
```

**文生图 + 参考强度**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v1",
    "prompt": "一个美丽的花园",
    "negative_prompt": "模糊,低质量,变形",
    "aspect_ratio": "1:1",
    "image_fidelity": 1,
    "human_fidelity": 1
}'
```

**批量生成（n 参数）**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v1",
    "prompt": "一只可爱的橘猫",
    "aspect_ratio": "16:9",
    "n": 3
}'
```

---

### 3.2 kling-v1-5

**接口**：`POST /v1/images/generations`

**能力**：文生图、单图生图（含角色特征参考、人物长相参考）。返回任务 ID，需轮询查询结果。

**参考文档**：[创建文生图或单图生图任务 (kling-v1-5)](https://apidocs.qnaigc.com/397001380e0)

#### Request 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | ✓ | 固定为 `kling-v1-5` |
| prompt | string | ✓ | 图像生成的文本描述，≤2500 字符 |
| image | string | | 参考图 URL 或 Base64（单图生图时使用） |
| n | integer | | 生成图像数量，1–10，默认 1 |
| negative_prompt | string | | 负向提示词，≤2500 字符；图生图时不可用 |
| aspect_ratio | string | | 画面纵横比，默认 `16:9` |
| image_reference | string | | `subject`：角色特征参考；`face`：人物长相参考 |
| image_fidelity | float | | 参考强度 [0,1]，默认 0.5 |
| human_fidelity | float | | 面部参考强度 [0,1]，仅 `image_reference=subject` 时生效，默认 0.45 |

#### curl 示例

**文生图**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v1-5",
    "prompt": "一只可爱的橘猫坐在窗台上看着夕阳,照片风格,高清画质",
    "aspect_ratio": "16:9"
}'
```

**单图生图 - 角色特征参考**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v1-5",
    "prompt": "一个穿着西装的商务人士",
    "image": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
    "image_reference": "subject",
    "image_fidelity": 0.7,
    "human_fidelity": 0.6,
    "aspect_ratio": "1:1"
}'
```

**单图生图 - 人物长相参考**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v1-5",
    "prompt": "一位微笑的女士,穿着西装",
    "image": "https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png",
    "image_reference": "face",
    "image_fidelity": 0.8,
    "aspect_ratio": "9:16"
}'
```

---

### 3.3 kling-v2

**接口**：`POST /v1/images/generations`（文生图/单图生图）、`POST /v1/images/edits`（多图生图）

**能力**：文生图、单图生图、多图生图。返回任务 ID，需轮询查询结果。

**参考文档**：
- [创建文生图或单图生图任务 (kling-v2)](https://apidocs.qnaigc.com/397002020e0)
- [创建多图生图任务 (kling-v2)](https://apidocs.qnaigc.com/397002021e0)

#### 3.3.1 文生图 / 单图生图（POST /v1/images/generations）

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | ✓ | 固定为 `kling-v2` |
| prompt | string | ✓ | 图像生成的文本描述，≤2500 字符 |
| image | string | | 参考图 URL 或 Base64（单图生图时使用） |
| n | integer | | 生成图像数量，1–10，默认 1 |
| negative_prompt | string | | 负向提示词，≤2500 字符 |
| aspect_ratio | string | | 画面纵横比，默认 `16:9` |

**文生图**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2",
    "prompt": "一只可爱的橘猫坐在窗台上看着夕阳,照片风格,高清画质",
    "aspect_ratio": "16:9"
}'
```

**单图生图**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2",
    "image": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
    "prompt": "将这张图片转换为水彩画风格",
    "aspect_ratio": "16:9"
}'
```

**文生图 + 负向提示词**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2",
    "prompt": "一个美丽的花园",
    "negative_prompt": "模糊,低质量,变形",
    "aspect_ratio": "1:1",
    "n": 1
}'
```

#### 3.3.2 多图生图（POST /v1/images/edits）

基于多张参考图像和文本描述生成新图像。需提供 2–4 张参考图。

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | ✓ | 固定为 `kling-v2` |
| image | string | ✓ | 多图生图时可传空字符串 |
| prompt | string | ✓ | 图像编辑的文本描述，≤2500 字符 |
| subject_image_list | array | ✓ | 参考图列表，2–4 张，每项含 `subject_image`（URL） |
| n | integer | | 生成图像数量，1–10，默认 1 |
| scene_image | string | | 场景参考图 URL |
| style_image | string | | 风格参考图 URL |
| aspect_ratio | string | | 画面纵横比，默认 `16:9` |

**多图生图 - 仅 subject_image_list**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/edits' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2",
    "image": "",
    "prompt": "综合两个图像画一个漫画图",
    "subject_image_list": [
        {
            "subject_image": "https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png"
        },
        {
            "subject_image": "https://aitoken-public.qnaigc.com/example/generate-image/image-to-image-with-mask-1.jpg"
        }
    ],
    "aspect_ratio": "16:9"
}'
```

**多图生图 - 含 scene_image / style_image**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/edits' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2",
    "image": "",
    "prompt": "一个梦幻般的森林场景",
    "subject_image_list": [
        {
            "subject_image": "https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png"
        },
        {
            "subject_image": "https://aitoken-public.qnaigc.com/example/generate-image/image-to-image-with-mask-1.jpg"
        }
    ],
    "scene_image": "https://aitoken-public.qnaigc.com/example/generate-image/image-to-image-with-mask-1.jpg",
    "style_image": "https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png",
    "aspect_ratio": "9:16"
}'
```

---

### 3.4 kling-v2-new

**接口**：`POST /v1/images/generations`

**能力**：仅单图生图（不支持文生图）。返回任务 ID，需轮询查询结果。

**参考文档**：[创建单图生图任务 (kling-v2-new)](https://apidocs.qnaigc.com/397002374e0)

#### Request 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | ✓ | 固定为 `kling-v2-new` |
| prompt | string | ✓ | 图像生成的文本描述，≤2500 字符 |
| image | string | ✓ | 参考图 URL 或 Base64 |
| n | integer | | 生成图像数量，1–10，默认 1 |
| negative_prompt | string | | 负向提示词，≤2500 字符 |
| aspect_ratio | string | | 画面纵横比，默认 `16:9` |

#### curl 示例

**单图生图**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2-new",
    "image": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
    "prompt": "将这张图片转换为赛博朋克风格",
    "aspect_ratio": "16:9"
}'
```

---

### 3.5 kling-v2-1

**接口**：`POST /v1/images/generations`（文生图/单图生图）、`POST /v1/images/edits`（多图生图）

**能力**：文生图、单图生图、多图生图。返回任务 ID，需轮询查询结果。

**参考文档**：
- [创建文生图或单图生图任务 (kling-v2-1)](https://apidocs.qnaigc.com/397002563e0)
- [创建多图生图任务 (kling-v2-1)](https://apidocs.qnaigc.com/397002564e0)

#### 3.5.1 文生图 / 单图生图（POST /v1/images/generations）

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | ✓ | 固定为 `kling-v2-1` |
| prompt | string | ✓ | 图像生成的文本描述，≤2500 字符 |
| image | string | | 参考图 URL 或 Base64（单图生图时使用） |
| n | integer | | 生成图像数量，1–10，默认 1 |
| negative_prompt | string | | 负向提示词，≤2500 字符 |
| aspect_ratio | string | | 画面纵横比，默认 `16:9` |

**文生图**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2-1",
    "prompt": "一只可爱的橘猫坐在窗台上看着夕阳,照片风格,高清画质",
    "aspect_ratio": "16:9"
}'
```

**单图生图**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2-1",
    "image": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
    "prompt": "将这张图片转换为动漫风格",
    "aspect_ratio": "16:9"
}'
```

**文生图 + 负向提示词**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/generations' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2-1",
    "prompt": "一个美丽的花园",
    "negative_prompt": "模糊,低质量,变形",
    "aspect_ratio": "1:1",
    "n": 1
}'
```

#### 3.5.2 多图生图（POST /v1/images/edits）

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | ✓ | 固定为 `kling-v2-1` |
| image | string | ✓ | 多图生图时可传空字符串 |
| prompt | string | ✓ | 图像编辑的文本描述，≤2500 字符 |
| subject_image_list | array | ✓ | 参考图列表，2–4 张，每项含 `subject_image`（URL） |
| n | integer | | 生成图像数量，1–10，默认 1 |
| scene_image | string | | 场景参考图 URL |
| style_image | string | | 风格参考图 URL |
| aspect_ratio | string | | 画面纵横比，默认 `16:9` |

**多图生图**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/images/edits' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2-1",
    "image": "",
    "prompt": "综合两个图像生成一个新场景",
    "subject_image_list": [
        {
            "subject_image": "https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png"
        },
        {
            "subject_image": "https://aitoken-public.qnaigc.com/example/generate-image/image-to-image-with-mask-1.jpg"
        }
    ],
    "aspect_ratio": "16:9"
}'
```

---

### 3.6 kling-image-o1（可灵 OmniImage）

**接口**：`POST /queue/fal-ai/kling-image/o1`

**能力**：文生图、多模态（文本 + 参考图）。支持在 prompt 中用 `<<<image_1>>>` 引用参考图（1-indexed）。任务为异步执行，创建成功后返回任务 ID，需通过查询接口获取生成结果。

**参考文档**：
- [创建图像生成任务 (kling-image-o1)](https://apidocs.qnaigc.com/411327685e0)
- [查询图像生成任务 (kling-image-o1)](https://apidocs.qnaigc.com/411327686e0)

#### Request 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| prompt | string | ✓ | 文本提示词，使用 `<<<image_1>>>` 引用参考图，最大 2500 字符 |
| image_urls | array | | 参考图 URL 列表，在 prompt 中用 `<<<image_1>>>` 引用（1-indexed），最多 10 张 |
| resolution | string | | 分辨率：`1K` 或 `2K`，默认 1K（兼容小写 1k/2k） |
| num_images | integer | | 生成数量 1–9，默认 1 |
| aspect_ratio | string | | 纵横比，默认 `auto` |

#### aspect_ratio 可选值

`auto`、`16:9`、`9:16`、`1:1`、`4:3`、`3:4`、`3:2`、`2:3`、`21:9`

#### Response 200（创建成功）

```json
{
  "status": "IN_QUEUE",
  "request_id": "qimage-root-1770199726278452760",
  "response_url": "https://api.qnaigc.com/queue/fal-ai/kling-image/requests/qimage-root-1770199726278452760",
  "status_url": "https://api.qnaigc.com/queue/fal-ai/kling-image/requests/qimage-root-1770199726278452760/status",
  "cancel_url": ""
}
```

#### 查询任务状态

**接口**：`GET /queue/fal-ai/kling-image/requests/{task_id}/status`

#### Response 200 - 已完成

```json
{
  "status": "COMPLETED",
  "request_id": "qimage-root-1770199726278452760",
  "metrics": { "inference_time": 38 },
  "result": {
    "images": [
      { "url": "https://...", "content_type": "image/png" },
      { "url": "https://...", "content_type": "image/png" }
    ]
  }
}
```

#### curl 示例

**纯文本生成**

```bash
curl --location --request POST 'https://api.qnaigc.com/queue/fal-ai/kling-image/o1' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "prompt": "一只可爱的橘猫在阳光下打盹",
    "num_images": 2,
    "resolution": "2K",
    "aspect_ratio": "16:9"
}'
```

**带参考图片生成**

```bash
curl --location --request POST 'https://api.qnaigc.com/queue/fal-ai/kling-image/o1' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "prompt": "参考 <<<image_1>>> 的风格，增加一群人，保持背景不变",
    "image_urls": [
        "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg"
    ],
    "num_images": 2,
    "resolution": "2K",
    "aspect_ratio": "16:9"
}'
```

**查询任务状态**

```bash
curl --location --request GET 'https://api.qnaigc.com/queue/fal-ai/kling-image/requests/qimage-root-1770105770654149000/status' \
--header 'Authorization: Bearer <token>'
```

---

### 3.7 查询图像任务状态（kling-v1 / v1-5 / v2 / v2-new / v2-1）

**接口**：`GET /v1/images/tasks/{task_id}`

#### Response 200 - 处理中

```json
{
  "task_id": "image-1762159125266058362-1383010xxx",
  "created": 1761793032,
  "status": "processing",
  "status_message": "处理中"
}
```

#### Response 200 - 已完成

```json
{
  "task_id": "image-1762159125266058362-1383010xxx",
  "created": 1761793032,
  "status": "completed",
  "data": [
    { "url": "https://example.com/generated-image-1.png" },
    { "url": "https://example.com/generated-image-2.png" }
  ]
}
```

#### Response 200 - 失败

```json
{
  "task_id": "image-1762159125266058362-1383010xxx",
  "created": 1761793032,
  "status": "failed",
  "status_message": "错误信息"
}
```

| status | 说明 |
|--------|------|
| processing | 处理中 |
| completed / succeeded | 已完成 |
| failed | 失败 |

#### curl 示例

```bash
curl --location --request GET 'https://api.qnaigc.com/v1/images/tasks/image-1762159125266058362-1383010xxx' \
--header 'Authorization: Bearer <token>'
```

---

## 4. xai 包参数映射

在 `github.com/goplus/xai/spec/kling` 中，通过 `Params.Set(name, value)` 设置参数：

| API 参数 | xai 常量 | 类型 | 说明 |
|----------|----------|------|------|
| prompt | `ParamPrompt` | string | 必填 |
| image | `ParamImage` | string | 参考图 URL |
| n | `ParamN` | int | 可选，生成数量 1–9（kling-image-o1）或 1–10，默认 1 |
| negative_prompt | `ParamNegativePrompt` | string | |
| aspect_ratio | `ParamAspectRatio` | string | 可选，用 `AspectAuto`/`Aspect16x9`/`Aspect9x16`/`Aspect1x1` 等，默认 auto |
| image_reference | `ParamImageReference` | string | subject / face |
| image_fidelity | `ParamImageFidelity` | float64 | |
| human_fidelity | `ParamHumanFidelity` | float64 | |
| subject_image_list | `ParamSubjectImageList` | []SubjectImageItem | 多图生图参考列表（kling-v2, kling-v2-1） |
| scene_image | `ParamSceneImage` | string | 场景参考图（kling-v2, kling-v2-1） |
| style_image | `ParamStyleImage` | string | 风格参考图（kling-v2, kling-v2-1） |
| reference_images | `ParamReferenceImages` | string / []string | 参考图 URL 列表，kling-image-o1 中对应 API 的 image_urls，prompt 用 `<<<image_1>>>` 引用 |
| resolution | `ParamResolution` | string | 可选，用 `Resolution1K`/`Resolution2K`（kling-image-o1，默认 1K） |

**设置 n 示例**：

```go
op.Params().Set(kling.ParamN, 3)  // 一次生成 3 张图
```

**kling-image-o1 带参考图示例**：

```go
op.Params().Set(kling.ParamPrompt, "参考 <<<image_1>>> 的风格，增加一群人，保持背景不变")
op.Params().Set(kling.ParamReferenceImages, []string{"https://example.com/ref.jpg"})
op.Params().Set(kling.ParamN, 2)
op.Params().Set(kling.ParamResolution, kling.Resolution2K)
op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)
```

---

## 5. 错误响应

```json
{
  "status": false,
  "message": "error message"
}
```

| HTTP 状态码 | 说明 |
|-------------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 认证失败 |

---

## 6. 参考链接

- [创建文生图或单图生图任务 (kling-v1)](https://apidocs.qnaigc.com/396977229e0)
- [创建文生图或单图生图任务 (kling-v1-5)](https://apidocs.qnaigc.com/397001380e0)
- [创建文生图或单图生图任务 (kling-v2)](https://apidocs.qnaigc.com/397002020e0)
- [创建多图生图任务 (kling-v2)](https://apidocs.qnaigc.com/397002021e0)
- [创建单图生图任务 (kling-v2-new)](https://apidocs.qnaigc.com/397002374e0)
- [创建文生图或单图生图任务 (kling-v2-1)](https://apidocs.qnaigc.com/397002563e0)
- [创建多图生图任务 (kling-v2-1)](https://apidocs.qnaigc.com/397002564e0)
- [创建图像生成任务 (kling-image-o1)](https://apidocs.qnaigc.com/411327685e0)
- [查询图像生成任务 (kling-image-o1)](https://apidocs.qnaigc.com/411327686e0)
