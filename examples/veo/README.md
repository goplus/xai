# Veo Examples

Runnable demos for Veo video generation through `spec/gemini/provider/qiniu`.

Reference: [spec/gemini/provider/qiniu/veo.md](../../spec/gemini/provider/qiniu/veo.md)

## Quick Start

```bash
# 1) set API key
export QINIU_API_KEY=your-key

# 2) list demos
go run ./examples/veo

# 3) run one model
go run ./examples/veo veo-3.0-generate-preview

# 4) run image-to-video (uses DemoURLs.Image)
go run ./examples/veo veo-image-to-video

# 5) run first+last frame (uses DemoURLs.FirstFrame, DemoURLs.LastFrame)
go run ./examples/veo veo-first-last-frame

# 6) run video input (引用视频, uses DemoURLs.Video)
go run ./examples/veo veo-video-input

# 7) run reference images (多参考图, veo-3.1 only)
go run ./examples/veo veo-reference-images

# 8) run all models (full coverage)
go run ./examples/veo all
```

## Model Coverage

Each model is implemented in its own file. Examples use URLs from `urls.go` (no local files). Align with [veo.md](../../spec/gemini/provider/qiniu/veo.md) curl samples:

| File | veo.md curl | Params |
|------|-------------|--------|
| `veo_3_0_generate_preview.go` | 3.1 文生视频 | 全参数：Prompt, AspectRatio, DurationSeconds, NumberOfVideos, Seed, NegativePrompt, PersonGeneration, GenerateAudio, Resolution, CompressionQuality |
| `veo_callback.go` | 3.3 回调模式 | 全参数 + PubsubTopic (callback_url) |
| `veo_image_to_video.go` | 3.2 图生视频 | Image + Prompt，使用 `DemoURLs.Image` |
| `veo_first_last_frame.go` | 3.4 首尾帧 | Image（首帧）+ LastFrame（尾帧）+ Prompt，使用 `DemoURLs.FirstFrame`、`DemoURLs.LastFrame` |
| `veo_video_input.go` | 3.4.1 引用视频 | Video + Prompt，使用 `DemoURLs.Video` |
| `veo_reference_images.go` | 多参考图 | ReferenceImages + Prompt，仅 veo-3.1/veo-2.0-exp，最多 3 asset 或 1 style，时长 8s |
| `veo_2_0_*.go` | - | aspectRatio, durationSeconds, personGeneration（Veo 2 无 resolution） |
| `veo_3_0_fast_*.go`, `veo_3_1_*.go` | - | 同上 |

- `veo_2_0_generate_001.go`
- `veo_2_0_generate_exp.go`
- `veo_2_0_generate_preview.go`
- `veo_3_0_generate_preview.go`
- `veo_3_0_fast_generate_preview.go`
- `veo_3_1_generate_preview.go`
- `veo_3_1_fast_generate_preview.go`
- `veo_callback.go` - text-to-video with callback_url (PubsubTopic)
- `veo_image_to_video.go` - image-to-video (veo.md 3.2)
- `veo_first_last_frame.go` - first+last frame (veo.md 3.4)
- `veo_video_input.go` - video as input (引用视频)
- `veo_reference_images.go` - multi reference images (多参考图)

## Notes

- **文生视频**：`Prompt` 必填。**图生视频**：`Image` + `Prompt`。**首尾帧**：`Image` + `LastFrame` + `Prompt`。**引用视频**：`Video` + `Prompt`。**多参考图**：`ReferenceImages` + `Prompt`（最多 3 asset 或 1 style，时长 8s）。图/视频输入均使用 URL。
- 其余参数可选，使用 `gemini.Aspect16x9`、`gemini.Duration8` 等常量。
- Examples use `Operation(xai.GenVideo)` + `xai.Wait` polling.

## Verify

```bash
go test ./examples/veo
```
