module github.com/goplus/xai

go 1.24.5

require github.com/goplus/xai/spec/openai v0.0.0

require (
	github.com/openai/openai-go/v3 v3.23.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
)

replace github.com/goplus/xai/spec/openai => ./spec/openai
