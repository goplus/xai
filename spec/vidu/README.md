# Vidu

`spec/vidu` provides a Vidu video generation specification implementation in the same style as `spec/kling`, supporting `vidu-q1`, `vidu-q2`, `viduq2-turbo` and `viduq2-pro` via `xai.GenVideo`.

## Supported Models

- `vidu-q1`
- `vidu-q2`
- `viduq2-turbo`
- `viduq2-pro`

## Supported Routes (Auto-detected)

- Text-to-video: `text-to-video`
- Reference-to-video: `reference-to-video`
- Image-to-video (q2): `image-to-video/pro`
- Start-end-to-video (q2): `start-end-to-video/pro`

## Core Capabilities

- Parameter normalization and validation (conflict detection, route legality check)
- Async task handling (create task + poll status)
- Unified result encapsulation (`xai.Results` / `xai.OutputVideo`)
- Backend abstraction (like `spec/gemini`) for pluggable providers
- Qiniu provider (`spec/vidu/provider/qiniu`)

## Quick Start

```go
ctx := context.Background()

svc := qiniu.NewService(os.Getenv("QINIU_API_KEY"))
op, _ := svc.Operation(xai.Model(vidu.ModelViduQ2), xai.GenVideo)
op.Params().
    Set(vidu.ParamPrompt, "A woman walking through a vibrant city street at night.").
    Set(vidu.ParamImageURL, "https://example.com/ref.jpg").
    Set(vidu.ParamDuration, 4).
    Set(vidu.ParamResolution, vidu.Resolution720p)

results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
if err != nil {
    panic(err)
}

videoURL := results.At(0).(*xai.OutputVideo).URL()
fmt.Println(videoURL)
```

## File Structure

```
spec/vidu/
├── vidu.go           # Register, Scheme, model constants
├── backend.go        # Backend interface (Submit, GetTaskStatus)
├── params.go         # Params, param constants, VideoParams, validation
├── operation.go      # genVideo, inputSchema
├── service.go        # Service, NewWithBackend
├── results.go        # SyncOperationResponse, AsyncOperationResponse, NewOutputVideos
├── media.go          # image helpers (delegates video to video/)
├── video/
│   ├── impl.go       # videoByURI, videoByBytes, NewVideoFromURI, NewVideoFromBytes
│   └── results.go    # outputVideos, NewOutputVideos
├── provider/
│   └── qiniu/
│       ├── backend.go   # Backend implementation
│       ├── client.go    # HTTP client
│       ├── service.go   # NewService, Register
│       ├── qiniu_test.go
│       ├── vidu_video.md
│       └── README.md
└── README.md
```

## Adding a New Provider

Implement `vidu.Backend` and pass it to `vidu.NewWithBackend`:

```go
type myBackend struct{ /* ... */ }

func (b *myBackend) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) { /* ... */ }
func (b *myBackend) GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) { /* ... */ }

svc := vidu.NewWithBackend(&myBackend{})
```

## Documentation

- [Vidu Video API Reference (Qiniu)](./provider/qiniu/vidu_video.md)
- [Qiniu Provider Guide](./provider/qiniu/README.md)
