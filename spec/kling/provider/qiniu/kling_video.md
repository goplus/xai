# Kling 视频生成 API 参考

本文档基于 [七牛云 Qnagic API](https://apidocs.qnaigc.com) 整理，描述 Kling 视频生成接口。

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

| 能力/参数 | kling-v2-1 | kling-v2-5-turbo | kling-v2-6~9 | kling-video-o1 | kling-v3 | kling-v3-omni |
|-----------|------------|------------------|--------------|----------------|----------|---------------|
| 文生视频 | - | ✓ | ✓ | ✓ | ✓ | ✓ |
| 图生视频 | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| 首尾帧生视频 | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| 视频生视频 | - | ✓ | - | ✓ | - | ✓ |
| 多图生视频 | - | - | - | ✓ (image_list) | - | ✓ |
| 有声视频 | - | - | ✓ (sound) | - | ✓ | ✓ |
| 动作控制 | - | - | ✓ (image_url+video_url) | - | - | - |
| input_reference / image_tail | ✓ | ✓ | ✓ | - | ✓ | - |
| image_list | - | - | - | ✓ | - | ✓ |
| video_list | - | - | - | ✓ | - | ✓ |
| negative_prompt | ✓ | ✓ | ✓ | - | ✓ | - |
| mode (std/pro) | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| seconds | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| size | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

---

## 3. 视频生成 API

所有视频模型统一使用 **`POST /v1/videos`** 创建任务，返回任务 ID 后需轮询 **`GET /v1/videos/{id}`** 获取生成结果。

---

### 3.1 kling-v2-1

**接口**：`POST /v1/videos`

**能力**：图生视频、首尾帧生视频。**不支持纯文生视频**，必须提供 `input_reference`。

**参考文档**：[创建视频任务 (kling-v2-1)](https://apidocs.qnaigc.com/396127760e0)

#### Request 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | ✓ | 固定为 `kling-v2-1` |
| prompt | string | ✓ | 视频生成的文本描述，≤2500 字符 |
| input_reference | string | ✓ | 参考图 URL 或 Base64（首帧） |
| image_tail | string | | 尾帧参考图 URL 或 Base64（首尾帧生视频时使用） |
| negative_prompt | string | | 负向提示词，≤2500 字符 |
| seconds | string | | 视频时长，`5` 或 `10`，默认 `5` |
| size | string | | 分辨率，默认 `1920x1080` |
| mode | string | | `std` 或 `pro`，默认 `std` |

#### Response 200（成功）

```json
{
    "id": "string",
    "object": "video",
    "model": "string",
    "mode": "string",
    "status": "string",
    "created_at": 0,
    "updated_at": 0,
    "seconds": "string",
    "size": "string"
}
```

#### curl 示例

**图生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2-1",
    "prompt": "人在奔跑",
    "input_reference": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
    "size": "1280x720",
    "mode": "pro"
}'
```

**首尾帧生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2-1",
    "prompt": "人在跑到了天涯海角",
    "input_reference": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
    "image_tail": "https://aitoken-public.qnaigc.com/example/generate-video/end-of-the-earth-and-sky.jpg",
    "size": "1280x720",
    "mode": "pro"
}'
```

---

### 3.2 kling-v2-5-turbo

**接口**：`POST /v1/videos`

**能力**：文生视频、图生视频、视频生视频、首尾帧生视频全功能。

**参考文档**：[创建视频任务 (kling-v2-5-turbo)](https://apidocs.qnaigc.com/396145809e0)

#### Request 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | ✓ | 固定为 `kling-v2-5-turbo` |
| prompt | string | ✓ | 视频生成的文本描述，≤2500 字符 |
| input_reference | string | | 参考图 URL 或 Base64（图生视频时使用） |
| image_tail | string | | 尾帧参考图（首尾帧生视频时使用） |
| negative_prompt | string | | 负向提示词，≤2500 字符 |
| seconds | string | | 视频时长，`5` 或 `10`，默认 `5` |
| size | string | | 分辨率，默认 `1920x1080` |
| mode | string | | `std` 或 `pro`，默认 `std` |

#### curl 示例

**文生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2-5-turbo",
    "prompt": "一只可爱的橘猫在阳光下奔跑，慢镜头，电影质感",
    "size": "1920x1080",
    "mode": "pro",
    "seconds": "5"
}'
```

**图生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2-5-turbo",
    "prompt": "人在奔跑",
    "input_reference": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
    "size": "1280x720",
    "mode": "pro"
}'
```

**首尾帧生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2-5-turbo",
    "prompt": "人在跑到了天涯海角",
    "input_reference": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
    "image_tail": "https://aitoken-public.qnaigc.com/example/generate-video/end-of-the-earth-and-sky.jpg",
    "size": "1280x720",
    "mode": "pro"
}'
```

---

### 3.3 kling-v2-6 ~ kling-v2-9

**接口**：`POST /v1/videos`

**能力**：文生视频、图生视频、多图生视频、有声视频、动作控制等。

**参考文档**：[创建视频任务 (kling-v2-6)](https://apidocs.qnaigc.com/404263380e0)

#### Request 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | ✓ | `kling-v2-6`、`kling-v2-7`、`kling-v2-8`、`kling-v2-9` |
| prompt | string | ✓ | 视频生成的文本描述，≤2500 字符 |
| input_reference | string | | 参考图 URL 或 Base64（图生视频时使用） |
| image_tail | string | | 尾帧参考图（首尾帧生视频时使用） |
| sound | string | | 是否生成声音，`on` 或 `off` |
| image_url | string | | 参考图片 URL（动作控制专用） |
| video_url | string | | 参考视频 URL（动作控制专用） |
| character_orientation | string | | 角色朝向（动作控制专用） |
| keep_original_sound | string | | 是否保留参考视频原声，`yes` 或 `no` |
| negative_prompt | string | | 负向提示词，≤2500 字符 |
| seconds | string | | 视频时长，`5` 或 `10`，默认 `5` |
| size | string | | 分辨率 |
| mode | string | | `std` 或 `pro`，默认 `std` |

#### curl 示例

**文生有声视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2-6",
    "prompt": "一个人在演讲",
    "mode": "pro",
    "sound": "on",
    "seconds": "5"
}'
```

**图生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2-6",
    "prompt": "让图片中的角色动起来",
    "mode": "pro",
    "input_reference": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
    "seconds": "5"
}'
```

**动作控制**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v2-6",
    "prompt": "女孩穿着灰色宽松T恤和牛仔短裤",
    "image_url": "https://p2-kling.klingai.com/kcdn/cdn-kcdn112452/kling-qa-test/multi-3.ng.png",
    "video_url": "https://p2-kling.klingai.com/kcdn/cdn-kcdn112452/kling-qa-test/dance.mp4",
    "keep_original_sound": "yes",
    "character_orientation": "image",
    "mode": "pro"
}'
```

---

### 3.4 kling-video-o1

**接口**：`POST /v1/videos`

**能力**：文生视频、图生视频、视频生视频、首尾帧生视频。使用 `image_list`、`video_list` 替代 `input_reference` / `image_tail`。

**参考文档**：[创建视频任务 (kling-video-o1)](https://apidocs.qnaigc.com/396194574e0)

#### Request 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | ✓ | 固定为 `kling-video-o1` |
| prompt | string | ✓ | 视频生成的文本描述，≤2500 字符 |
| image_list | array | | 参考图列表，最多 7 张，每项含 `image`、`type`(first_frame/end_frame) |
| video_list | array | | 参考视频列表，最多 1 个，每项含 `video_url`、`refer_type`(feature/base)、`keep_original_sound` |
| seconds | string | | 视频时长，`5` 或 `10`，默认 `5` |
| size | string | | 分辨率，默认 `1920x1080` |
| mode | string | | `std` 或 `pro`，默认 `std` |

#### curl 示例

**文生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-video-o1",
    "prompt": "一只可爱的橘猫在阳光下奔跑，慢镜头，电影质感",
    "size": "1920x1080",
    "mode": "pro",
    "seconds": "5"
}'
```

**图生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-video-o1",
    "prompt": "这个人在跑马拉松",
    "image_list": [
        {
            "image": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg"
        }
    ],
    "size": "1920x1080",
    "mode": "pro"
}'
```

**首尾帧生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-video-o1",
    "prompt": "视频连贯在一起",
    "image_list": [
        {
            "image": "https://picsum.photos/1280/720",
            "type": "first_frame"
        },
        {
            "image": "https://picsum.photos/1280/720",
            "type": "end_frame"
        }
    ],
    "size": "1920x1080",
    "mode": "pro"
}'
```

**视频生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-video-o1",
    "prompt": "画面中增加一个小狗",
    "video_list": [
        {
            "video_url": "https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4",
            "refer_type": "feature",
            "keep_original_sound": "yes"
        }
    ],
    "size": "1920x1080",
    "mode": "pro"
}'
```

---

### 3.5 kling-v3

**接口**：`POST /v1/videos`

**能力**：文生视频、图生视频、有声视频。

**参考文档**：[创建视频任务 (kling-v3)](https://apidocs.qnaigc.com/421288748e0)

#### Request 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | ✓ | 固定为 `kling-v3` |
| prompt | string | ✓ | 视频生成的文本描述，≤2500 字符 |
| input_reference | string | | 参考图 URL 或 Base64（图生视频时使用） |
| sound | string | | 是否生成声音，`on` 或 `off` |
| seconds | string | | 视频时长，`3`~`15`，默认 `5` |
| size | string | | 分辨率 |
| mode | string | | `std` 或 `pro`，默认 `pro` |

#### curl 示例

**文生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v3",
    "prompt": "一只可爱的小猫在阳光下玩耍",
    "mode": "std",
    "seconds": "5",
    "size": "1280x720"
}'
```

**图生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v3",
    "prompt": "让图片中的角色动起来",
    "mode": "pro",
    "input_reference": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
    "seconds": "5"
}'
```

**有声视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v3",
    "prompt": "一个人在演讲",
    "mode": "pro",
    "sound": "on",
    "seconds": "5"
}'
```

---

### 3.6 kling-v3-omni

**接口**：`POST /v1/videos`

**能力**：文生视频、图生视频、视频生视频、首尾帧生视频、有声视频。支持 `image_list`、`video_list`、`multi_shot`、`multi_prompt` 等高级能力。

**参考文档**：[创建视频任务 (kling-v3-omni)](https://apidocs.qnaigc.com/420877484e0)

#### Request 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| model | string | ✓ | 固定为 `kling-v3-omni` |
| prompt | string | | 视频生成的文本描述，≤2500 字符 |
| multi_shot | boolean | | 是否生成多镜头视频，默认 false |
| multi_prompt | array | | 各分镜信息（multi_shot 时使用） |
| image_list | array | | 参考图列表，最多 7 张 |
| video_list | array | | 参考视频列表，最多 1 个 |
| sound | string | | 是否生成声音，`on` 或 `off`，默认 `off` |
| seconds | string | | 视频时长，`3`~`15`，默认 `5` |
| size | string | | 分辨率 |
| mode | string | | `std` 或 `pro`，默认 `pro` |

#### curl 示例

**文生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v3-omni",
    "prompt": "一只可爱的橘猫在阳光下奔跑，慢镜头，电影质感",
    "size": "1920x1080",
    "mode": "pro",
    "seconds": "5"
}'
```

**图生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v3-omni",
    "prompt": "这个人在跑马拉松",
    "image_list": [
        {
            "image": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg"
        }
    ],
    "size": "1920x1080",
    "mode": "pro"
}'
```

**首尾帧生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v3-omni",
    "prompt": "视频连贯在一起",
    "image_list": [
        {
            "image": "https://picsum.photos/1280/720",
            "type": "first_frame"
        },
        {
            "image": "https://picsum.photos/1280/720",
            "type": "end_frame"
        }
    ],
    "size": "1920x1080",
    "mode": "pro"
}'
```

**视频生视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v3-omni",
    "prompt": "画面中增加一个小狗",
    "video_list": [
        {
            "video_url": "https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4",
            "refer_type": "feature",
            "keep_original_sound": "yes"
        }
    ],
    "size": "1920x1080",
    "mode": "pro"
}'
```

**文生有声视频**

```bash
curl --location --request POST 'https://api.qnaigc.com/v1/videos' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
    "model": "kling-v3-omni",
    "prompt": "一只可爱的橘猫在阳光下奔跑，慢镜头，电影质感",
    "size": "1920x1080",
    "mode": "pro",
    "seconds": "3",
    "sound": "on"
}'
```

---

### 3.7 查询视频任务状态

**接口**：`GET /v1/videos/{id}`

**参考文档**：[查询视频生成状态](https://apidocs.qnaigc.com/396127761e0)

#### Response 200 - 处理中

```json
{
    "id": "string",
    "object": "video",
    "model": "string",
    "mode": "string",
    "status": "initializing",
    "created_at": 0,
    "updated_at": 0,
    "completed_at": 0,
    "seconds": "string",
    "size": "string"
}
```

#### Response 200 - 已完成

```json
{
    "id": "string",
    "object": "video",
    "model": "string",
    "mode": "string",
    "status": "completed",
    "created_at": 0,
    "updated_at": 0,
    "completed_at": 0,
    "seconds": "string",
    "size": "string",
    "task_result": {
        "videos": [
            {
                "id": "string",
                "url": "string",
                "duration": "string"
            }
        ]
    }
}
```

#### Response 200 - 失败

```json
{
    "id": "string",
    "object": "video",
    "status": "failed",
    "error": {
        "code": "string",
        "message": "string"
    }
}
```

| status | 说明 |
|--------|------|
| initializing | 初始化中 |
| queued | 排队中 |
| in_progress | 处理中 |
| downloading | 下载中 |
| uploading | 上传中 |
| completed | 已完成 |
| failed | 失败 |
| cancelled | 已取消 |

#### curl 示例

```bash
curl --location --request GET 'https://api.qnaigc.com/v1/videos/qvideo-user123-1766391125174150336' \
--header 'Authorization: Bearer <token>'
```

---

## 4. xai 包参数映射

在 `github.com/goplus/xai/spec/kling` 中，通过 `Params.Set(name, value)` 设置参数：

| API 参数 | xai 常量 | 类型 | 说明 |
|----------|----------|------|------|
| prompt | `ParamPrompt` | string | 必填 |
| input_reference | `ParamInputReference` | string | 参考图 URL（首帧） |
| image_tail | `ParamImageTail` | string | 尾帧参考图 |
| image_list | `ParamImageList` | []ImageInput | 参考图列表（O1/V3-omni） |
| video_list | `ParamVideoList` | []VideoRef | 参考视频列表（O1/V3-omni） |
| negative_prompt | `ParamNegativePrompt` | string | |
| mode | `ParamMode` | string | 可选，用 `ModeStd` / `ModePro` |
| seconds | `ParamSeconds` | string | 可选，用 `Seconds5` / `Seconds10`（V3/V3-omni 支持 3~15） |
| size | `ParamSize` | string | 可选，用 `Size1920x1080` / `Size1280x720` 等 |
| sound | `ParamSound` | string | 可选，用 `SoundOn` / `SoundOff`（V2.6/V3/V3-omni） |
| image_url | `ParamImageUrl` | string | 动作控制参考图（V2.6） |
| video_url | `ParamVideoUrl` | string | 动作控制参考视频（V2.6） |
| character_orientation | `ParamCharacterOrientation` | string | 角色朝向（V2.6 动作控制） |
| keep_original_sound | `ParamKeepOriginalSound` | string | 是否保留参考视频原声（V2.6） |
| multi_shot | `ParamMultiShot` | bool | 是否多镜头（V3-omni） |
| shot_type | `ParamShotType` | string | 分镜方式（V3-omni） |
| multi_prompt | `ParamMultiPrompt` | []MultiPromptItem | 各分镜信息（V3-omni） |

**可选参数常量**（见 `params.go`）：

- `Size1920x1080`, `Size1080x1920`, `Size1280x720`, `Size720x1280`, `Size1080x1080`, `Size720x720`
- `ModeStd`, `ModePro`
- `Seconds5`, `Seconds10`
- `SoundOn`, `SoundOff`

**设置参数示例**：

```go
op.Params().Set(kling.ParamPrompt, "人在奔跑")
op.Params().Set(kling.ParamInputReference, "https://example.com/first-frame.png")
op.Params().Set(kling.ParamImageTail, "https://example.com/end-frame.png")
op.Params().Set(kling.ParamMode, kling.ModePro)
op.Params().Set(kling.ParamSeconds, kling.Seconds5)
op.Params().Set(kling.ParamSize, kling.Size1920x1080)
op.Params().Set(kling.ParamSound, kling.SoundOn)  // kling-v2-6, kling-v3
op.Params().Set(kling.ParamImageUrl, "https://example.com/character.png")   // V2.6 动作控制
op.Params().Set(kling.ParamVideoUrl, "https://example.com/dance.mp4")
op.Params().Set(kling.ParamCharacterOrientation, "image")
op.Params().Set(kling.ParamKeepOriginalSound, "yes")
op.Params().Set(kling.ParamMultiShot, true)   // V3-omni 多镜头
op.Params().Set(kling.ParamMultiPrompt, []map[string]interface{}{{"index": 1, "prompt": "场景1", "duration": "5"}})
```

---

## 5. 错误响应

```json
{
  "error": {
    "message": "error message",
    "type": "invalid_request_error",
    "code": "invalid_request"
  }
}
```

| HTTP 状态码 | 说明 |
|-------------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 认证失败 |
| 404 | 视频任务不存在 |

---

## 6. 参考链接

- [创建视频任务 (kling-v2-1)](https://apidocs.qnaigc.com/396127760e0)
- [查询视频生成状态](https://apidocs.qnaigc.com/396127761e0)
- [创建视频任务 (kling-v2-5-turbo)](https://apidocs.qnaigc.com/396145809e0)
- [创建视频任务 (kling-v2-6)](https://apidocs.qnaigc.com/404263380e0)
- [创建视频任务 (kling-video-o1)](https://apidocs.qnaigc.com/396194574e0)
- [创建视频任务 (kling-v3)](https://apidocs.qnaigc.com/421288748e0)
- [创建视频任务 (kling-v3-omni)](https://apidocs.qnaigc.com/420877484e0)
