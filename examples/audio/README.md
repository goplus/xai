# Audio Examples

ASR (speech-to-text) and TTS (text-to-speech) examples via `spec/audio` and Qiniu provider.

## Quick Start

```bash
# List available demos
go run ./examples/audio

# Run a single demo (mock mode, no API key)
go run ./examples/audio list-voices
go run ./examples/audio asr
go run ./examples/audio tts
go run ./examples/audio tts-voice

# Run all demos
go run ./examples/audio all
```

## Backend Mode

- **Mock** (default): No API key needed. Returns placeholder results. Works in CI.
- **Real**: Set `QINIU_API_KEY` to use the Qiniu API for actual ASR/TTS.

```bash
export QINIU_API_KEY=your-key
go run ./examples/audio asr
go run ./examples/audio tts
go run ./examples/audio list-voices
```

## Demos

| Demo | Description |
|------|-------------|
| list-voices | List available TTS voices (voice_type, category, sample URL) |
| asr | ASR: Transcribe audio URL to text |
| tts | TTS: Synthesize text to audio (default voice) |
| tts-voice | TTS with specific voice (qiniu_zh_female_wwxkjx) |

## API Mapping

| xai | Qiniu API |
|-----|-----------|
| `Operation(asr, Transcribe)` | POST /v1/voice/asr |
| `Operation(tts-v1, Synthesize)` | POST /v1/voice/tts |
| `Service.ListVoices(ctx)` | GET /v1/voice/list |

## See Also

- [spec/audio/README.md](../../spec/audio/README.md)
- [spec/audio/provider/qiniu/audio.md](../../spec/audio/provider/qiniu/audio.md)
