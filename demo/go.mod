module github.com/goplus/xai/demo

go 1.25.0

require (
	github.com/goplus/xai v0.1.0
	github.com/goplus/xai/spec/claude v0.0.0
	github.com/goplus/xai/spec/gemini v0.0.0
	github.com/goplus/xai/spec/openai v0.0.0
)

replace (
	github.com/goplus/xai => ../
	github.com/goplus/xai/spec/claude => ../spec/claude
	github.com/goplus/xai/spec/gemini => ../spec/gemini
	github.com/goplus/xai/spec/openai => ../spec/openai
)
