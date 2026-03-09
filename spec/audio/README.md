# spec/audio

Audio specification for ASR (Speech Recognition) and TTS (Text-to-Speech), following the same design as `spec/kling`.

## Design

- **Actions**: `Transcribe` (ASR), `Synthesize` (TTS)
- **Models**: `asr` (ASR), `tts-v1` (TTS)
- **Operation**: Both ASR and TTS are synchronous (return immediately with results)

## Usage

### With Qiniu Provider

```go
import (
    "context"
    xai "github.com/goplus/xai/spec"
    "github.com/goplus/xai/spec/audio"
    "github.com/goplus/xai/spec/audio/provider/qiniu"
)

func main() {
    // Register audio service with Qiniu
    qiniu.Register("your-api-token")

    ctx := context.Background()
    svc, err := xai.New(ctx, "audio://")
    if err != nil {
        log.Fatal(err)
    }

    // ASR (Transcribe): audio URL -> text
    op, _ := svc.Operation(xai.Model("asr"), xai.Transcribe)
    op.Params().Set(audio.ParamAudio, "https://example.com/audio.mp3")
    resp, err := op.Call(ctx, svc, nil)
    if err != nil {
        log.Fatal(err)
    }
    if resp.Done() {
        text := resp.Results().At(0).(*xai.OutputText)
        fmt.Println("Transcribed:", text.Text)
    }

    // TTS (Synthesize): text -> audio
    op, _ = svc.Operation(xai.Model("tts-v1"), xai.Synthesize)
    op.Params().Set(audio.ParamInput, "Hello, world")
    op.Params().Set(audio.ParamVoice, "qiniu_zh_female_wwxkjx")
    resp, _ = op.Call(ctx, svc, nil)
    if resp.Done() {
        out := resp.Results().At(0).(*xai.OutputAudio)
        fmt.Println("Audio URL:", out.Audio, "Format:", out.Format)
    }
}
```

### ListVoices (Voice List)

When using the Qiniu provider, `ListVoices` returns available TTS voices. Use `VoiceType` as `ParamVoice` when calling Synthesize.

```go
if audioSvc, ok := svc.(*audio.Service); ok {
    voices, _ := audioSvc.ListVoices(ctx)
    for _, v := range voices {
        fmt.Printf("%s: %s\n", v.VoiceName, v.VoiceType)
    }
}
```

### Params

| Action     | Required | Optional                    |
|------------|----------|-----------------------------|
| Transcribe | audio (URL or {format, url}) | model, format |
| Synthesize | input (text) | voice, format, speed |

### Custom Service

```go
asrExec := myASRExecutor{}
ttsExec := myTTSExecutor{}
svc := audio.NewService(asrExec, ttsExec)
audio.Register(svc)
```

## Provider: Qiniu

- **ASR**: POST `/voice/asr` — body: `{model: "asr", audio: {format, url}}`
- **TTS**: POST `/voice/tts` — body: `{audio: {voice_type, encoding, speed_ratio}, request: {text}}`
- **Voice list**: GET `/voice/list` — returns `[{voice_name, voice_type, url, category, updatetime}]`

Base URL: `https://api.qnaigc.com/v1` (configurable via `WithBaseURL`).
