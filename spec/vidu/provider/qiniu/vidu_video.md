# Vidu Video Generation API Reference

This document is based on [Qiniu Qnagic API](https://apidocs.qnaigc.com) and describes the Vidu (Q1/Q2) video generation interfaces.

---

## 1. Overview

| Item | Description |
|------|-------------|
| China endpoint | `https://api.qnaigc.com` |
| Overseas endpoint | `https://openai.sufy.com` |
| Authentication | Bearer Token: `Authorization: Bearer <token>` |
| Content-Type | `application/json` |
| Task mode | Async: create task first, then poll status |

---

## 2. Model Capability Comparison

| Capability/Parameter | vidu-q1 | vidu-q2 | viduq2-turbo | viduq2-pro |
|----------------------|---------|---------|--------------|------------|
| Text-to-video | ✓ (`/q1/text-to-video`) | ✓ (`/q2/text-to-video`) | ✓ | ✓ |
| Reference-to-video (`reference_image_urls`) | ✓ | ✓ | ✓ | ✓ |
| Reference subjects (`subjects`) | ✓ | ✓ | ✓ | ✓ |
| Image-to-video (`image_url`) | - | ✓ (`/q2/image-to-video/pro`) | ✓ (`/q2/image-to-video/turbo`) | ✓ (`/q2/image-to-video/pro`) |
| Start-end-to-video (`start_image_url` + `end_image_url`) | - | ✓ (`/q2/start-end-to-video/pro`) | ✓ (`/q2/start-end-to-video/turbo`) | ✓ (`/q2/start-end-to-video/pro`) |
| Common params: `prompt`/`seed`/`duration`/`resolution`/`movement_amplitude` | ✓ | ✓ | ✓ | ✓ |
| `watermark` | ✓ | ✓ | ✓ | ✓ |

---

### 2.5 Common Parameters (All Interfaces)

| Parameter | Type | Required | Default | Limits / Allowed Values |
|-----------|------|----------|---------|------------------------|
| prompt | string | ✓ (most) | - | Max 2000 characters |
| seed | int | | 0 (random) | 0 or omit = random; manual value = fixed seed |
| duration | int | | 5 | **q1**: only 5; **q2 text-to-video**: 1–10; **q2 others**: 5 |
| resolution | string | | model-dependent | q1: 1080p; q2: 720p, 1080p |
| movement_amplitude | string | | auto | auto, small, medium, large |
| aspect_ratio | string | | 16:9 | 16:9, 9:16, 3:4, 4:3, 1:1 (3:4, 4:3 q2 only) |
| style | string | | general | general, anime (**q2 not effective**) |
| bgm | bool | | false | Add background music |
| watermark | bool | | - | Add watermark |

---

## 3. Video Generation API

### 3.1 vidu-q1 Text-to-Video

**Endpoint**: `POST /queue/fal-ai/vidu/q1/text-to-video`

**Reference**: [Create task (q1 text-to-video)](https://apidocs.qnaigc.com/417903311e0)

#### Request Parameters

| Parameter | Type | Required | Default | Limits / Allowed Values |
|-----------|------|----------|---------|------------------------|
| prompt | string | ✓ | - | Max 2000 characters |
| seed | int | | 0 (random) | 0 or omit = random |
| duration | int | | 5 | Only 5 |
| resolution | string | | 1080p | 1080p |
| movement_amplitude | string | | auto | auto, small, medium, large |
| style | string | | general | general, anime |
| aspect_ratio | string | | 16:9 | 16:9, 9:16, 1:1 |
| bgm | bool | | false | Add BGM |
| watermark | bool | | - | Add watermark |

#### curl Example

```bash
curl --location --request POST 'https://api.qnaigc.com/queue/fal-ai/vidu/q1/text-to-video' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "prompt": "A cute orange cat chasing butterflies in the sunlight, slow motion, cinematic look, warm lighting",
  "seed": 1,
  "duration": 5,
  "resolution": "1080p",
  "movement_amplitude": "auto"
}'
```

#### Create Task Response

```json
{
  "status": "IN_QUEUE",
  "request_id": "qvideo-root-1770718628185826174",
  "response_url": "https://api.qnaigc.com/queue/fal-ai/vidu/requests/qvideo-root-1770718628185826174",
  "status_url": "https://api.qnaigc.com/queue/fal-ai/vidu/requests/qvideo-root-1770718628185826174/status",
  "cancel_url": ""
}
```

---

### 3.2 vidu-q1 Reference-to-Video

**Endpoint**: `POST /queue/fal-ai/vidu/q1/reference-to-video`

**Reference**:
- [Create task (q1 reference-to-video)](https://apidocs.qnaigc.com/417907439e0)
- [Create task (q1 reference-to-video subjects)](https://apidocs.qnaigc.com/417907829e0)

Two reference input modes (choose one):

- `reference_image_urls`: Multiple reference image URLs (1–7 images)
- `subjects`: Named subject list (1–7 subjects; each subject 1–3 images; total 1–7 images; supports `@id` in prompt)

#### Mode A: `reference_image_urls`

| Parameter | Type | Required | Default | Limits |
|-----------|------|----------|---------|--------|
| prompt | string | ✓ | - | Max 2000 chars |
| reference_image_urls | []string | ✓ | - | 1–7 images; see [Image Constraints](#4-image-constraints) |
| seed | int | | 0 | - |
| duration | int | | 5 | Only 5 |
| resolution | string | | 1080p | 1080p |
| movement_amplitude | string | | auto | auto, small, medium, large |
| aspect_ratio | string | | 16:9 | 16:9, 9:16, 1:1 |
| bgm | bool | | false | - |
| watermark | bool | | - | - |

```bash
curl --location --request POST 'https://api.qnaigc.com/queue/fal-ai/vidu/q1/reference-to-video' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "prompt": "The little devil is looking at the apple on the beach and walking around it.",
  "reference_image_urls": [
    "https://storage.googleapis.com/falserverless/web-examples/vidu/new-examples/reference1.png",
    "https://storage.googleapis.com/falserverless/web-examples/vidu/new-examples/reference2.png",
    "https://storage.googleapis.com/falserverless/web-examples/vidu/new-examples/reference3.png"
  ],
  "seed": 2,
  "duration": 5,
  "resolution": "1080p",
  "movement_amplitude": "auto",
  "watermark": true
}'
```

#### Mode B: `subjects`

| Parameter | Type | Required | Default | Limits |
|-----------|------|----------|---------|--------|
| prompt | string | ✓ | - | Max 2000 chars; use `@subject_id` in text |
| subjects | []Subject | ✓ | - | 1–7 subjects; each subject: id, images (1–3), voice_id |
| seed | int | | 0 | - |
| duration | int | | 5 | Only 5 |
| resolution | string | | 1080p | 1080p |
| movement_amplitude | string | | auto | auto, small, medium, large |
| aspect_ratio | string | | 16:9 | 16:9, 9:16, 1:1 |
| bgm | bool | | false | - |
| audio | bool | | - | true = audio-video output |

```bash
curl --location --request POST 'https://api.qnaigc.com/queue/fal-ai/vidu/q1/reference-to-video' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "prompt": "The @devil is looking at the @apple on the @beach and walking around @beach.",
  "subjects": [
    {"id": "devil", "images": ["https://example.com/reference1.png"], "voice_id": ""},
    {"id": "apple", "images": ["https://example.com/reference3.png"], "voice_id": ""},
    {"id": "beach", "images": ["https://example.com/reference2.png"], "voice_id": ""}
  ],
  "seed": 2,
  "duration": 5,
  "resolution": "1080p",
  "movement_amplitude": "auto"
}'
```

---

### 3.3 vidu-q2 Text-to-Video

**Endpoint**: `POST /queue/fal-ai/vidu/q2/text-to-video`

**Reference**: [Create task (q2 text-to-video)](https://apidocs.qnaigc.com/417911128e0)

#### Request Parameters

| Parameter | Type | Required | Default | Limits / Allowed Values |
|-----------|------|----------|---------|------------------------|
| prompt | string | ✓ | - | Max 2000 characters |
| seed | int | | 0 (random) | - |
| duration | int | | 5 | **1–10** (q2 text-to-video only) |
| resolution | string | | 1080p | 720p, 1080p |
| movement_amplitude | string | | auto | auto, small, medium, large |
| aspect_ratio | string | | 16:9 | 16:9, 9:16, 3:4, 4:3, 1:1 |
| bgm | bool | | false | - |
| watermark | bool | | - | - |

#### curl Example

```bash
curl --location --request POST 'https://api.qnaigc.com/queue/fal-ai/vidu/q2/text-to-video' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "prompt": "A cute orange cat chasing butterflies in the sunlight, slow motion, cinematic look, warm lighting",
  "seed": 1,
  "duration": 5,
  "resolution": "1080p",
  "movement_amplitude": "auto"
}'
```

---

### 3.4 vidu-q2 Reference-to-Video

**Endpoint**: `POST /queue/fal-ai/vidu/q2/reference-to-video`

**Reference**:
- [Create task (q2 reference-to-video)](https://apidocs.qnaigc.com/417911896e0)
- [Create task (q2 reference-to-video subjects)](https://apidocs.qnaigc.com/417911901e0)

#### Mode A: `reference_image_urls`

| Parameter | Type | Required | Default | Limits |
|-----------|------|----------|---------|--------|
| prompt | string | ✓ | - | Max 2000 chars |
| reference_image_urls | []string | ✓ | - | 1–7 images |
| seed | int | | 0 | - |
| duration | int | | 5 | 5 |
| resolution | string | | 720p | 720p, 1080p |
| movement_amplitude | string | | auto | auto, small, medium, large |
| aspect_ratio | string | | 16:9 | 16:9, 9:16, 3:4, 4:3, 1:1 |
| bgm | bool | | false | - |
| watermark | bool | | - | - |

#### Mode B: `subjects`

Same structure as q1; 1–7 subjects, each with 1–3 images, total 1–7 images.

---

### 3.5 vidu-q2 Image-to-Video (Pro)

**Endpoint**: `POST /queue/fal-ai/vidu/q2/image-to-video/pro`

**Reference**: [Create task (q2 image-to-video/pro)](https://apidocs.qnaigc.com/417912235e0)

#### Request Parameters

| Parameter | Type | Required | Default | Limits |
|-----------|------|----------|---------|--------|
| prompt | string | | - | Max 2000 chars |
| image_url | string | ✓ | - | 1 image; see [Image Constraints](#4-image-constraints) |
| seed | int | | 0 | - |
| duration | int | | 5 | 5 |
| resolution | string | | 1080p | 720p, 1080p |
| movement_amplitude | string | | auto | auto, small, medium, large |
| watermark | bool | | - | - |
| audio | bool | | false | Audio-video output |
| voice_id | string | | - | Voice ID when audio=true |
| is_rec | bool | | false | Use recommended prompt (+10 credits) |

#### curl Example

```bash
curl --location --request POST 'https://api.qnaigc.com/queue/fal-ai/vidu/q2/image-to-video/pro' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "prompt": "A woman walking through a vibrant city street at night, neon lights reflecting off wet pavement.",
  "image_url": "http://t9vdmzkpm.hb-bkt.clouddn.com/testIMG_9179.jpeg?e=1770100380&token=Dzm7Yw6NngN0lPssNlIqyQyLGHeeRdHWfAMrxiCW:vYjos8ylxZQCumL2cibtugUNCtI=",
  "seed": 2,
  "duration": 4,
  "resolution": "720p",
  "movement_amplitude": "auto",
  "watermark": true
}'
```

---

### 3.6 vidu-q2 Start-End-to-Video (Pro)

**Endpoint**: `POST /queue/fal-ai/vidu/q2/start-end-to-video/pro`

**Reference**: [Create task (q2 start-end-to-video/pro)](https://apidocs.qnaigc.com/417918627e0)

#### Request Parameters

| Parameter | Type | Required | Default | Limits |
|-----------|------|----------|---------|--------|
| prompt | string | | - | Max 2000 chars |
| start_image_url | string | ✓ | - | 1 image |
| end_image_url | string | ✓ | - | 1 image |
| seed | int | | 0 | - |
| duration | int | | 5 | 5 |
| resolution | string | | 1080p | 720p, 1080p |
| movement_amplitude | string | | auto | auto, small, medium, large |
| watermark | bool | | - | - |
| is_rec | bool | | false | Use recommended prompt (+10 credits) |

#### curl Example

```bash
curl --location --request POST 'https://api.qnaigc.com/queue/fal-ai/vidu/q2/start-end-to-video/pro' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "prompt": "Dragon lands on a rock",
  "start_image_url": "https://v3.fal.media/files/zebra/sgsdKvPigPhJ1S7Hl5bWc_first_frame_q1.png",
  "end_image_url": "https://v3.fal.media/files/kangaroo/CASBu_OmOnZ8IafirarFL_last_frame_q1.png",
  "seed": 2,
  "duration": 4,
  "resolution": "720p",
  "movement_amplitude": "auto",
  "watermark": true
}'
```

---

### 3.7 vidu-q2 Image-to-Video (Turbo)

**Endpoint**: `POST /queue/fal-ai/vidu/q2/image-to-video/turbo`

Turbo variant for faster generation; same parameters as Pro.

#### Request Parameters

| Parameter | Type | Required | Default | Limits |
|-----------|------|----------|---------|--------|
| prompt | string | | - | Max 2000 chars |
| image_url | string | ✓ | - | 1 image; see [Image Constraints](#4-image-constraints) |
| seed | int | | 0 | - |
| duration | int | | 5 | 1–10 |
| resolution | string | | 720p | 720p, 1080p |
| movement_amplitude | string | | auto | auto, small, medium, large |
| watermark | bool | | - | - |

#### curl Example

```bash
curl --location --request POST 'https://api.qnaigc.com/queue/fal-ai/vidu/q2/image-to-video/turbo' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "prompt": "A woman walking through a vibrant city street at night, neon lights reflecting off wet pavement.",
  "image_url": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
  "seed": 2,
  "duration": 4,
  "resolution": "720p",
  "movement_amplitude": "auto",
  "watermark": true
}'
```

---

### 3.8 vidu-q2 Start-End-to-Video (Turbo)

**Endpoint**: `POST /queue/fal-ai/vidu/q2/start-end-to-video/turbo`

Turbo variant for faster generation; same parameters as Pro.

#### Request Parameters

| Parameter | Type | Required | Default | Limits |
|-----------|------|----------|---------|--------|
| prompt | string | | - | Max 2000 chars |
| start_image_url | string | ✓ | - | 1 image |
| end_image_url | string | ✓ | - | 1 image |
| seed | int | | 0 | - |
| duration | int | | 5 | 1–10 |
| resolution | string | | 720p | 720p, 1080p |
| movement_amplitude | string | | auto | auto, small, medium, large |
| watermark | bool | | - | - |

#### curl Example

```bash
curl --location --request POST 'https://api.qnaigc.com/queue/fal-ai/vidu/q2/start-end-to-video/turbo' \
--header 'Authorization: Bearer <token>' \
--header 'Content-Type: application/json' \
--data-raw '{
  "prompt": "Dragon lands on a rock",
  "start_image_url": "https://v3.fal.media/files/zebra/sgsdKvPigPhJ1S7Hl5bWc_first_frame_q1.png",
  "end_image_url": "https://v3.fal.media/files/kangaroo/CASBu_OmOnZ8IafirarFL_last_frame_q1.png",
  "seed": 2,
  "duration": 4,
  "resolution": "720p",
  "movement_amplitude": "auto",
  "watermark": true
}'
```

---

### 3.9 Query Task Status

**Endpoint**: `GET /queue/fal-ai/vidu/requests/{request_id}/status`

**Reference**:
- [Query task status (q1)](https://apidocs.qnaigc.com/417919085e0)
- [Query task status (q2)](https://apidocs.qnaigc.com/417938129e0)

#### Response Example (Completed)

```json
{
  "status": "COMPLETED",
  "request_id": "qvideo-root-1770720203757316283",
  "response_url": "https://api.qnaigc.com/queue/fal-ai/vidu/requests/qvideo-root-1770720203757316283",
  "status_url": "https://api.qnaigc.com/queue/fal-ai/vidu/requests/qvideo-root-1770720203757316283/status",
  "cancel_url": "",
  "metrics": {
    "inference_time": 2252
  },
  "result": {
    "video": {
      "url": "https://aitoken-video.qnaigc.com/root/qvideo-root-1770720203757316283/1.mp4?...",
      "content_type": "video/mp4"
    }
  }
}
```

#### Common Statuses

| status | Description |
|--------|-------------|
| `IN_QUEUE` | Queued |
| `IN_PROGRESS` / `PROCESSING` / `RUNNING` | Processing |
| `COMPLETED` | Completed |
| `FAILED` / `ERROR` | Failed |
| `CANCELLED` / `CANCELED` | Cancelled |

---

## 4. Image Constraints

All image parameters (`reference_image_urls`, `subjects[].images`, `image_url`, `start_image_url`, `end_image_url`) must satisfy:

| Constraint | Limit |
|------------|-------|
| Formats | png, jpeg, jpg, webp |
| Min pixels | 128×128 |
| Aspect ratio | < 1:4 or 4:1 |
| Single image size | ≤ 50 MB |
| POST body total | ≤ 20 MB |
| Base64 | Must include content-type, e.g. `data:image/png;base64,{base64}` |

**Count limits**:

- `reference_image_urls`: 1–7 images
- `subjects`: 1–7 subjects; each subject 1–3 images; total 1–7 images
- `image_url`: 1 image
- `start_image_url` / `end_image_url`: 1 image each

---

## 5. xai Parameter Mapping (`spec/vidu`)

In `github.com/goplus/xai/spec/vidu`, set parameters via `Params.Set(name, value)`:

| API Parameter | xai Constant | Type | Description |
|---------------|--------------|------|-------------|
| prompt | `ParamPrompt` | string | Required; max 2000 chars |
| seed | `ParamSeed` | int | 0 = random |
| duration | `ParamDuration` | int | q1: 5; q2 text-to-video: 1–10; q2 others: 5 |
| resolution | `ParamResolution` | string | `Resolution720p` / `Resolution1080p` |
| movement_amplitude | `ParamMovementAmplitude` | string | auto, small, medium, large |
| watermark | `ParamWatermark` | bool | - |
| reference_image_urls | `ParamReferenceImageURLs` | []string | 1–7; mutually exclusive with subjects |
| subjects | `ParamSubjects` | []vidu.Subject | 1–7; mutually exclusive with reference_image_urls |
| image_url | `ParamImageURL` | string | q2 image-to-video only |
| start_image_url | `ParamStartImageURL` | string | Must pair with end_image_url |
| end_image_url | `ParamEndImageURL` | string | Must pair with start_image_url |

**Example**:

```go
op.Params().Set(vidu.ParamPrompt, "A woman walking through a vibrant city street at night.")
op.Params().Set(vidu.ParamImageURL, "https://example.com/ref.jpg")
op.Params().Set(vidu.ParamDuration, 4)
op.Params().Set(vidu.ParamResolution, vidu.Resolution720p)
op.Params().Set(vidu.ParamMovementAmplitude, vidu.MovementAuto)
op.Params().Set(vidu.ParamWatermark, true)
```

---

## 6. Constraints and Errors

Validation in `spec/vidu` (`params.go`):

| Constraint | Error |
|------------|-------|
| prompt required | `ErrPromptRequired` |
| prompt length > 2000 | `ErrPromptTooLong` |
| duration <= 0 | `ErrInvalidDuration` |
| vidu-q1 duration != 5 | `ErrInvalidQ1Duration` |
| vidu-q2 / viduq2-turbo / viduq2-pro text-to-video duration not in 1–10 | `ErrInvalidQ2Duration` |
| vidu-q2 / viduq2-turbo / viduq2-pro others duration != 5 | `ErrInvalidQ2Duration` |
| reference_image_urls and subjects both set | `ErrReferenceInputsConflict` |
| image_url mixed with reference params | `ErrConflictingGenerationMode` |
| start/end mixed with image/reference | `ErrConflictingGenerationMode` |
| start_image_url without end_image_url (or vice versa) | `ErrStartEndPairRequired` |
| reference_image_urls count not 1–7 | `ErrInvalidReferenceCount` |
| subjects count not 1–7 | `ErrInvalidSubjectsCount` |
| subject images per subject > 3 or total not 1–7 | `ErrInvalidSubjectImages` |
| vidu-q1 image-to-video / start-end-to-video | `ErrRouteNotSupported` |
| resolution not in enum | `ErrValueNotAllowed` |
| movement_amplitude not in enum | `ErrValueNotAllowed` |
| aspect_ratio not in enum | `ErrValueNotAllowed` |

---

## 7. Reference Links

- [q1 text-to-video](https://apidocs.qnaigc.com/417903311e0)
- [q1 reference-to-video](https://apidocs.qnaigc.com/417907439e0)
- [q1 reference-to-video (subjects)](https://apidocs.qnaigc.com/417907829e0)
- [q1 status](https://apidocs.qnaigc.com/417919085e0)
- [q2 text-to-video](https://apidocs.qnaigc.com/417911128e0)
- [q2 reference-to-video](https://apidocs.qnaigc.com/417911896e0)
- [q2 reference-to-video (subjects)](https://apidocs.qnaigc.com/417911901e0)
- [q2 image-to-video/pro](https://apidocs.qnaigc.com/417912235e0)
- [q2 start-end-to-video/pro](https://apidocs.qnaigc.com/417918627e0)
- q2 image-to-video/turbo: `POST /queue/fal-ai/vidu/q2/image-to-video/turbo`
- q2 start-end-to-video/turbo: `POST /queue/fal-ai/vidu/q2/start-end-to-video/turbo`
- [q2 status](https://apidocs.qnaigc.com/417938129e0)
