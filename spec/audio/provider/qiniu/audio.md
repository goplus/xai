# Qiniu ASR/TTS API Reference

This document is based on [Qiniu ASR/TTS API Documentation](https://developer.qiniu.com/aitokenapi/12981/asr-tts-ocr-api) and describes how `spec/audio` maps to Qiniu ASR/TTS APIs.

---

## 1. Overview

| Item | Description |
|------|--------------|
| API Base URL | `https://api.qnaigc.com/v1` |
| Authentication | `Authorization: Bearer <your Qiniu AI API KEY>` |
| Content-Type | `application/json` |

---

## 2. Supported Endpoints

| Endpoint | Description | xai Mapping |
|----------|--------------|-------------|
| POST /voice/asr | Speech-to-text recognition | `Operation(asr, Transcribe)` |
| POST /voice/tts | Text-to-speech synthesis | `Operation(tts-v1, Synthesize)` |
| GET /voice/list | List available voices | `Service.ListVoices()` |

---

## 3. Speech-to-Text (ASR)

Supports multi-language (e.g. Chinese, English) speech-to-text recognition with over 95% accuracy in noisy environments.  
Supported audio formats: raw / wav / mp3 / ogg.

### 3.1 Request

- **URL**: `POST /v1/voice/asr`
- **Content-Type**: `application/json`

#### Request Parameters

| Field | Type | Required | Description | xai Param |
|-------|------|----------|-------------|-----------|
| model | string | Yes | Model name, fixed as asr | fixed asr |
| audio | object | Yes | Audio parameters | ParamAudio |
| └─ format | string | Yes | Audio format (e.g. mp3) | ParamFormat or audio.format |
| └─ url | string | Yes | Public URL of the audio file | ParamAudio (string URL) or audio.url |

#### Request Example

```json
{
  "model": "asr",
  "audio": {
    "format": "mp3",
    "url": "https://static.qiniu.com/ai-inference/example-resources/example.mp3"
  }
}
```

#### Response Structure

| Field | Type | Description |
|-------|------|-------------|
| reqid | string | Request ID |
| operation | string | Operation type, fixed as asr |
| data | object | Recognition result |
| └─ audio_info | object | Audio metadata |
| └─ duration | int | Audio duration (milliseconds) |
| └─ result | object | Recognized text and additions |
| └─ text | string | Recognized text |
| └─ additions | object | Additional info |

#### Response Example

```json
{
  "reqid": "bdf5e1b1bcaca22c7a9248aba2804912",
  "operation": "asr",
  "data": {
    "audio_info": { "duration": 9336 },
    "result": {
      "additions": { "duration": "9336" },
      "text": "Qiniu's culture is to be a simple person, make a simple product, and build a simple company."
    }
  }
}
```

#### curl Example

```bash
curl --location "https://api.qnaigc.com/v1/voice/asr" \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $OPENAI_API_KEY" \
  --data '{
    "model": "asr",
    "audio": {
      "format": "mp3",
      "url": "https://static.qiniu.com/ai-inference/example-resources/example.mp3"
    }
  }'
```

---

## 4. Text-to-Speech (TTS)

### 4.1 List Voices

- **URL**: `GET /v1/voice/list`

#### Response Structure

| Field | Type | Description |
|-------|------|-------------|
| voice_name | string | Voice display name |
| voice_type | string | Voice ID (used as voice_type in TTS requests) |
| url | string | Sample audio URL |
| category | string | Voice category |
| updatetime | int | Last update time (milliseconds) |

#### Response Example

```json
[
  {
    "voice_name": "Sweet Teaching Xiaoyuan",
    "voice_type": "qiniu_zh_female_tmjxxy",
    "url": "https://aitoken-public.qnaigc.com/ai-voice/qiniu_zh_female_tmjxxy.mp3",
    "category": "Traditional voice",
    "updatetime": 1747812605559
  }
]
```

#### curl Example

```bash
curl --location "https://api.qnaigc.com/v1/voice/list" \
  --header "Authorization: Bearer $OPENAI_API_KEY"
```

### 4.2 Text-to-Speech

- **URL**: `POST /v1/voice/tts`
- **Content-Type**: `application/json`

#### Request Parameters

| Field | Type | Required | Description | xai Param |
|-------|------|----------|-------------|-----------|
| audio | object | Yes | Audio parameters | - |
| └─ voice_type | string | Yes | Voice type | ParamVoice |
| └─ encoding | string | Yes | Audio encoding (e.g. mp3) | ParamFormat |
| └─ speed_ratio | float | No | Speech rate, default 1.0 | ParamSpeed |
| request | object | Yes | Request parameters | - |
| └─ text | string | Yes | Text to synthesize | ParamInput |

#### Request Example

```json
{
  "audio": {
    "voice_type": "qiniu_zh_female_wwxkjx",
    "encoding": "mp3",
    "speed_ratio": 1.0
  },
  "request": {
    "text": "Hello, world!"
  }
}
```

#### Response Structure

| Field | Type | Description |
|-------|------|-------------|
| reqid | string | Request ID |
| operation | string | Operation type |
| sequence | int | Sequence number, typically -1 |
| data | string | Base64-encoded synthesized audio |
| addition | object | Additional info |
| └─ duration | string | Audio duration (milliseconds) |

#### Response Example

```json
{
  "reqid": "f3dff20d7d670df7adcb2ff0ab5ac7ea",
  "operation": "query",
  "sequence": -1,
  "data": "<base64-encoded audio data>",
  "addition": { "duration": "1673" }
}
```

#### curl Example

```bash
curl --location "https://api.qnaigc.com/v1/voice/tts" \
  --header "Content-Type: application/json" \
  --header "Authorization: Bearer $OPENAI_API_KEY" \
  --data '{
    "audio": {
      "voice_type": "qiniu_zh_female_wwxkjx",
      "encoding": "mp3",
      "speed_ratio": 1.0
    },
    "request": {
      "text": "Hello, world!"
    }
  }'
```

---

## 5. spec/audio to Qiniu API Mapping

| xai Call | Qiniu API |
|----------|-----------|
| `Operation(asr, Transcribe)` + `ParamAudio` | POST /v1/voice/asr |
| `Operation(tts-v1, Synthesize)` + `ParamInput`, `ParamVoice` | POST /v1/voice/tts |
| `Service.ListVoices(ctx)` | GET /v1/voice/list |

### Parameter Mapping

| xai Param | Qiniu ASR | Qiniu TTS |
|-----------|-----------|-----------|
| ParamAudio | audio {format, url} | - |
| ParamInput | - | request.text |
| ParamVoice | - | audio.voice_type |
| ParamFormat | audio.format | audio.encoding |
| ParamSpeed | - | audio.speed_ratio |

---

## 6. Real-time ASR (WebSocket)

Qiniu also provides WebSocket-based real-time speech recognition for streaming microphone input:

- **URL**: `wss://api.qnaigc.com/v1/voice/asr`

`spec/audio` currently implements HTTP synchronous ASR only; real-time ASR requires separate integration.

---

## References

- [Qiniu ASR/TTS API Documentation](https://developer.qiniu.com/aitokenapi/12981/asr-tts-ocr-api)
