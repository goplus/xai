# Vidu Examples

Runnable demos for Vidu video generation, organized in the same style as `examples/kling/video`.

## Quick Start

```bash
# List demos
go run ./examples/vidu/video

# Run individual demos
go run ./examples/vidu/video q1-text
go run ./examples/vidu/video q1-ref-urls
go run ./examples/vidu/video q1-ref-subjects
go run ./examples/vidu/video q1-ref-subjects-audio
go run ./examples/vidu/video q2-text
go run ./examples/vidu/video q2-ref-urls
go run ./examples/vidu/video q2-ref-subjects
go run ./examples/vidu/video q2-image-pro
go run ./examples/vidu/video q2-image-pro-audio
go run ./examples/vidu/video q2-image-turbo
go run ./examples/vidu/video q2-start-end-pro

# CallSync + TaskID + GetTask
go run ./examples/vidu/video call-sync

# Run all demos
go run ./examples/vidu/video all
```

## Backend Mode

- **Mock** (default): no API key required, async mock executor returns placeholder URLs.
- **Real**: set `QINIU_API_KEY` to call Qnagic API.

```bash
export QINIU_API_KEY=your-key
go run ./examples/vidu/video q2-text
go run ./examples/vidu/video q2-image-pro-audio
```

## Request Coverage

| Demo | API |
|------|-----|
| `q1-text` | `POST .../q1/text-to-video` with `style` / `aspect_ratio` / `bgm` |
| `q1-ref-urls` | `POST .../q1/reference-to-video` (`reference_image_urls`) |
| `q1-ref-subjects` | `POST .../q1/reference-to-video` (`subjects`) |
| `q1-ref-subjects-audio` | `POST .../q1/reference-to-video` (`subjects` + `audio=true`) |
| `q2-text` | `POST .../q2/text-to-video` with `aspect_ratio` / `bgm` |
| `q2-ref-urls` | `POST .../q2/reference-to-video` (`reference_image_urls`) |
| `q2-ref-subjects` | `POST .../q2/reference-to-video` (`subjects`) |
| `q2-image-pro` | `POST .../q2/image-to-video/pro` |
| `q2-image-pro-audio` | `POST .../q2/image-to-video/pro` on `viduq2-pro` with `audio` / `voice_id` / `is_rec` |
| `q2-image-turbo` | `POST .../q2/image-to-video/turbo` on `viduq2-turbo` |
| `q2-start-end-pro` | `POST .../q2/start-end-to-video/pro` |
| `call-sync` | `CallSync` + `xai.GetTask` resume + polling |

## Directory Structure

```text
examples/vidu/
├── README.md
├── shared/
│   └── service.go
├── output/
│   └── output.go
└── video/
    ├── main.go
    ├── urls.go
    ├── helpers.go
    ├── vidu_q1_text_to_video.go
    ├── vidu_q1_reference_urls.go
    ├── vidu_q1_reference_subjects.go
    ├── vidu_q1_reference_subjects_audio.go
    ├── vidu_q2_text_to_video.go
    ├── vidu_q2_reference_urls.go
    ├── vidu_q2_reference_subjects.go
    ├── vidu_q2_image_to_video_pro.go
    ├── vidu_q2_image_to_video_pro_audio.go
    ├── vidu_q2_image_to_video_turbo.go
    ├── vidu_q2_start_end_to_video_pro.go
    └── call_sync_example.go
```

## Verify

```bash
go test ./examples/vidu/... -v
```
