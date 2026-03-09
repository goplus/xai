# Qiniu Sora API Documentation

This document describes the complete specification of Qiniu's Sora video generation API and how `spec/openai` maps `xai.GenVideo` to Qiniu's video endpoints.

## API Reference

- Create video: https://apidocs.qnaigc.com/412986563e0
- Query status: https://apidocs.qnaigc.com/395453398e0
- Video Remix: https://apidocs.qnaigc.com/395453399e0
- Remix (sora-2-pro): https://apidocs.qnaigc.com/412986565e0

## Implementation

- `spec/openai/operation.go`
- `spec/openai/operation_video.go`

## Usage

- `openai.New(...)`
- `openai.NewV1(...)` — supports extended response parsing (e.g. images) when base and key are provided
- `openai/provider/qiniu.NewService(...)`
- `openai/provider/qiniu.NewVideoService(...)` — video-only capability

---

## Supported Models

| Model | Description | Default Resolution | Supported Resolutions | Duration |
|-------|-------------|-------------------|----------------------|----------|
| `sora-2` | Standard | 1280x720 | 720x1280, 1280x720 | 4, 8, 12 sec |
| `sora-2-pro` | Pro | 1280x720 | 1280x720, 1024x1792, 1792x1024 | 4, 8, 12 sec |

---

## Supported Operations

For model names starting with `sora-`:

- `Actions(model)` returns `[xai.GenVideo]`
- `Operation(model, xai.GenVideo)` returns a video operation

Other models return `xai.ErrNotFound` for `GenVideo`.

---

## Input Parameters

Parameters supported by `Operation(..., xai.GenVideo).Params().Set(...)`:

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `Prompt` | `string` | Yes | Text prompt, max 2500 characters |
| `InputReference` | `string` / `xai.Image` / `xai.Video` | No | Reference image URL or data URI for image-to-video |
| `Seconds` | `string`/number | No | Video duration, options: `4`, `8`, `12` |
| `Size` | `string` | No | Resolution, see table below |
| `RemixFromVideoID` | `string` | No | If set, calls remix endpoint instead of creating new task |

### Resolution Constraints

**sora-2:**
- `720x1280` (portrait)
- `1280x720` (landscape, default)

**sora-2-pro:**
- `1280x720` (landscape)
- `1024x1792` (portrait)
- `1792x1024` (landscape)

### Validation Rules

- `Prompt` is required
- `Seconds` and `Size` must be within allowed values
- In Remix mode (`RemixFromVideoID` set), `InputReference`, `Seconds`, and `Size` are not allowed

---

## API Endpoints

### 1. Create Video

```
POST https://api.qnaigc.com/v1/videos
```

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request body (text-to-video):**
```json
{
    "model": "sora-2",
    "prompt": "A cute orange cat chasing butterflies in a sunny garden, photo style, HD quality",
    "seconds": "4",
    "size": "1280x720"
}
```

**Request body (image-to-video):**
```json
{
    "model": "sora-2",
    "prompt": "This person is running a marathon",
    "input_reference": "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
    "seconds": "4",
    "size": "1280x720"
}
```

**sora-2-pro example (portrait 1024x1792):**
```json
{
    "model": "sora-2-pro",
    "prompt": "A cute orange cat chasing butterflies in a sunny garden, photo style, HD quality",
    "seconds": "4",
    "size": "1024x1792"
}
```

### 2. Query Task Status

```
GET https://api.qnaigc.com/v1/videos/{video_id}
```

**Headers:**
```
Authorization: Bearer <token>
```

### 3. Video Remix

```
POST https://api.qnaigc.com/v1/videos/{video_id}/remix
```

Regenerates video from a completed task with a new prompt, preserving the original duration and resolution.

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request body:**
```json
{
    "prompt": "Change the scene to night, add neon lights, cyberpunk style"
}
```

---

## Response Format

### Create/Remix success response (queued)

```json
{
    "id": "qvideo-user123-1766453713089395279",
    "object": "video",
    "model": "sora-2",
    "status": "queued",
    "created_at": 1766453713,
    "updated_at": 1766453713,
    "seconds": "4",
    "size": "1280x720"
}
```

### Remix response (with source)

```json
{
    "id": "qvideo-user123-1766454050923137689",
    "object": "video",
    "model": "sora-2",
    "status": "queued",
    "created_at": 1766454050,
    "updated_at": 1766454050,
    "seconds": "4",
    "size": "1280x720",
    "remixed_from_video_id": "qvideo-user123-1766453713089395279"
}
```

### In progress (in_progress)

```json
{
    "id": "qvideo-root-1766453713089395279",
    "object": "video",
    "model": "sora-2",
    "status": "in_progress",
    "created_at": 1766453713,
    "updated_at": 1766453713,
    "seconds": "4",
    "size": "1280x720"
}
```

### Completed (completed)

```json
{
    "id": "qvideo-user123-1766453713089395279",
    "object": "video",
    "model": "sora-2",
    "status": "completed",
    "created_at": 1766453713,
    "updated_at": 1766453836,
    "completed_at": 1766453836,
    "seconds": "4",
    "size": "1280x720",
    "task_result": {
        "videos": [
            {
                "id": "qvideo-user123-1766453713089395279-1",
                "url": "https://aitoken-video.qnaigc.com/user123/qvideo-user123-1766453713089395279/1.mp4?e=1767058636&token=...",
                "duration": "4"
            }
        ]
    }
}
```

### Failed (failed)

```json
{
    "id": "qvideo-user123-1234567890123456789",
    "object": "video",
    "model": "sora-2",
    "status": "failed",
    "created_at": 1766453713,
    "updated_at": 1766453836,
    "completed_at": 1766453836,
    "error": {
        "code": "moderation_blocked",
        "message": "Your request was blocked by our moderation system."
    },
    "seconds": "4",
    "size": "1280x720"
}
```

---

## Status Mapping

Provider status values are normalized as:

| Category | Status Values |
|----------|---------------|
| In progress | `queued`, `pending`, `running`, `in_progress`, `processing` |
| Completed | `completed`, `succeeded`, `success`, `done` |
| Failed | `failed`, `error`, `cancelled`, `canceled`, `aborted` |

Unknown statuses remain pending unless `task_result.videos` already contains output.

---

## Operation Response Behavior

`Operation.Call(...)` returns an `xai.OperationResponse`:

- `TaskID()` — video task ID
- `Done()` — whether the operation is complete
- `Retry(ctx, svc)` — fetch latest status
- `Results()` — `xai.Results` (videos mapped to `xai.OutputVideo`)

Failure responses implement `xai.OperationResponseWithError`:

- `GetError()` returns an error built from API `error.code` and `error.message`.

---

## Output Mapping

`task_result.videos[].url` is mapped to `xai.OutputVideo.URL()`.

Video MIME type inference:

- `.webm` → `video/webm`
- otherwise → `video/mp4`

---

## Examples

### Text-to-Video

```go
ctx := context.Background()
svc := qiniu.NewVideoService(os.Getenv("QINIU_API_KEY"))

op, err := svc.Operation(xai.Model("sora-2"), xai.GenVideo)
if err != nil {
    panic(err)
}

op.Params().
    Set("Prompt", "A cute orange cat chasing butterflies in a sunny garden, photo style, HD quality").
    Set("Seconds", "4").
    Set("Size", "1280x720")

resp, err := xai.CallSync(ctx, svc, op, nil)
if err != nil {
    panic(err)
}

results, err := xai.Wait(ctx, svc, resp, nil)
if err != nil {
    panic(err)
}

for i := 0; i < results.Len(); i++ {
    out := results.At(i).(*xai.OutputVideo)
    fmt.Println(out.URL())
}
```

### Image-to-Video

```go
op.Params().
    Set("Prompt", "This person is running a marathon").
    Set("InputReference", "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg").
    Set("Seconds", "4").
    Set("Size", "1280x720")
```

### Remix

```go
op.Params().
    Set("Prompt", "Change the scene to night, add neon lights, cyberpunk style").
    Set("RemixFromVideoID", "qvideo-user123-1766453713089395279")
```

### Resume Existing Task

```go
resp, err := xai.GetTask(ctx, svc, xai.Model("sora-2"), xai.GenVideo, taskID)
if err != nil {
    panic(err)
}
results, err := xai.Wait(ctx, svc, resp, nil)
```

**Note:** `GetTask` uses the baseURL from when the Service was created. If you specified an overseas endpoint via `opts.WithBaseURL()` when creating the task, `GetTask` will still use the Service's default baseURL when resuming. When resuming tasks, use a Service with the same baseURL as at creation time, e.g. via `NewVideoService(token, WithBaseURL(OverseasBaseURL))`.

---

## Endpoints

| Region | Base URL |
|--------|----------|
| China | `https://api.qnaigc.com/v1/` |
| Overseas | `https://openai.sufy.com/v1/` |

Authentication: `Authorization: Bearer <API_KEY>`
