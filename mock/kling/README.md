# Kling Mock Tests

This directory contains mock-backed integration tests for the runnable Kling demos
under `examples/kling`.

The tests run the real example entrypoints with:

- `QINIU_API_KEY=test-key`
- `QINIU_MOCK_CURL=1`

and assert against the emitted curl payloads, so we can verify request-building
behavior without making live Qiniu API calls.

Run all Kling mock tests with:

```bash
go test ./mock/kling/...
```
