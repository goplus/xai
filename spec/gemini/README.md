# spec/gemini

`spec/gemini` is the XAI framework's Gemini protocol implementation, providing full integration with Google Gemini API and Vertex AI.

## Architecture Overview

```
spec/gemini/
├── gemini.go      # Service definition, New/NewWithBackend constructors
├── backend.go     # Backend interface definition, genAIBackend implementation
├── operation.go   # Image/video Operation implementations
├── schema.go      # Reflection-based InputSchema/Params, Image/Video types
├── params.go      # ParamBuilder implementation
├── options.go     # OptionBuilder implementation
├── message.go     # MsgBuilder/TextBuilder implementation
├── response.go    # GenResponse/Candidate/Part implementation
├── data.go        # ImageBuilder/DocumentBuilder implementation
├── tool.go        # Tool/WebSearchTool implementation
└── provider/      # Third-party Provider implementations
    └── qiniu/     # Qiniu Cloud Gemini Provider
```

## Core Interfaces

### Service

```go
type Service struct {
    backend Backend
    tools   tools
}

// Create standard Gemini Service (uses google.golang.org/genai)
func New(ctx context.Context, uri string) (xai.Service, error)

// Create Service with custom Backend
func NewWithBackend(backend Backend) *Service

// Get underlying Backend
func (p *Service) Backend() Backend
```

**URI format**:
- Gemini API: `gemini:base=https://generativelanguage.googleapis.com/&key=YOUR_API_KEY`
- Vertex AI: `gemini:project=PROJECT_ID&location=us-central1`

### Backend Interface

```go
type Backend interface {
    Actions(model xai.Model) []xai.Action

    // Text generation
    GenerateContent(ctx, model, contents, config) (*genai.GenerateContentResponse, error)
    GenerateContentStream(ctx, model, contents, config) iter.Seq2[*genai.GenerateContentResponse, error]

    // Video generation
    GenerateVideosFromSource(ctx, model, source, config) (*genai.GenerateVideosOperation, error)
    GetVideosOperation(ctx, op, config) (*genai.GenerateVideosOperation, error)

    // Image operations
    GenerateImages(ctx, model, prompt, config) (*genai.GenerateImagesResponse, error)
    EditImage(ctx, model, prompt, references, config) (*genai.EditImageResponse, error)
    RecontextImage(ctx, model, source, config) (*genai.RecontextImageResponse, error)
    UpscaleImage(ctx, model, image, factor, config) (*genai.UpscaleImageResponse, error)
    SegmentImage(ctx, model, source, config) (*genai.SegmentImageResponse, error)
}

// BackendService is implemented by services that have a gemini Backend
type BackendService interface {
    xai.Service
    Backend() Backend
}
```

## Exported Utilities

### Schema Reflection

```go
// Auto-generate InputSchema from struct pointer
func NewInputSchema(params any) xai.InputSchema

// InputSchema with field restrictions
func NewInputSchemaEx(params any, restriction map[string]*xai.Restriction) xai.InputSchema

// Reflection-based Params setter
func NewParams(params any) xai.Params
```

### Operation Response

```go
// Create synchronous OperationResponse (Done() == true)
func NewSyncResponse(ret xai.Results) xai.OperationResponse
```

## Supported Operations

| Action | Description | Async |
|--------|-------------|-------|
| `xai.GenVideo` | Video generation | ✓ |
| `xai.GenImage` | Image generation | |
| `xai.EditImage` | Image editing | |
| `xai.RecontextImage` | Image recontext | |
| `xai.UpscaleImage` | Image upscaling | |
| `xai.SegmentImage` | Image segmentation | |

## Usage Examples

### Basic Text Generation

```go
import (
    "context"
    xai "github.com/goplus/xai/spec"
    _ "github.com/goplus/xai/spec/gemini" // register "gemini" scheme
)

func main() {
    ctx := context.Background()
    svc, _ := xai.New(ctx, "gemini:key=YOUR_API_KEY")

    params := svc.Params().
        Model("gemini-2.0-flash").
        System(svc.Texts("You are a helpful assistant.")).
        Messages(svc.UserMsg().Text("Hello!"))

    resp, _ := svc.Gen(ctx, params, nil)
    fmt.Println(resp.At(0).Part(0).Text())
}
```

### Image Generation Operation

```go
op, _ := svc.Operation("gemini-2.0-flash-exp", xai.GenImage)
op.Params().
    Set("Prompt", "A cute orange cat").
    Set("AspectRatio", "16:9")

resp, _ := op.Call(ctx, svc, nil)
for !resp.Done() {
    resp.Sleep()
    resp, _ = resp.Retry(ctx, svc)
}
results := resp.Results()
img := results.At(0).(*xai.OutputImage).Image
```

### Custom Backend

```go
import "github.com/goplus/xai/spec/gemini"

type myBackend struct {
    // implement gemini.Backend interface
}

func main() {
    backend := &myBackend{}
    svc := gemini.NewWithBackend(backend)
    // use svc...
}
```

## Provider Development Guide

To create a custom Provider (e.g. `provider/qiniu`):

1. **Implement the `gemini.Backend` interface**
   ```go
   type backend struct {
       client *client
   }

   func (p *backend) Actions(model xai.Model) []xai.Action { ... }
   func (p *backend) GenerateContent(...) { ... }
   // ... other methods
   ```

2. **Create Service with `gemini.NewWithBackend`**
   ```go
   type Service struct {
       *gemini.Service
   }

   func NewService(token string) *Service {
       backend := newBackend(token)
       return &Service{
           Service: gemini.NewWithBackend(backend),
       }
   }
   ```

3. **Register URI Scheme**
   ```go
   func Register(token string) {
       xai.Register("my-provider", func(ctx context.Context, uri string) (xai.Service, error) {
           return NewService(token), nil
       })
   }
   ```

By embedding `*gemini.Service`, a Provider automatically gets:
- `Gen` / `GenStream` text generation
- `Actions` / `Operation` image/video operations
- `ImageFrom*` / `VideoFrom*` media factory methods
- `ReferenceImage` / `GenVideoMask` and other helpers
- `Params` / `Options` / `UserMsg` / `AssistantMsg` builders

## Related Documentation

- [provider/qiniu/gemini.md](provider/qiniu/gemini.md) - Qiniu Provider detailed documentation
