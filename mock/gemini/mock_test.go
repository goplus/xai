package gemini

import (
	"testing"

	"github.com/goplus/xai/mock/testutil"
)

var geminiMockEnv = map[string]string{
	"QINIU_API_KEY":   "test-key",
	"QINIU_MOCK_CURL": "1",
}

func TestGeminiExamplesMock(t *testing.T) {
	tests := []struct {
		demo  string
		wants []string
	}{
		{
			demo: "chat-text",
			wants: []string{
				`--- chat-text ---`,
				`https://api.qnaigc.com/v1/chat/completions`,
				`response { candidates: 1 }`,
				`text: "Gemini mock response."`,
			},
		},
		{
			demo: "chat-image",
			wants: []string{
				`--- chat-image ---`,
				`"text":"Change this image to red."`,
				`image-to-image-1.jpg`,
				`text: "I changed the image to a red style."`,
				`data_base64: "aGVsbG8="`,
			},
		},
		{
			demo: "chat-tool",
			wants: []string{
				`--- chat-tool ---`,
				`first_response`,
				`type: "tool_use"`,
				`id: "call_mock_weather"`,
				`name: "get_weather"`,
				`final_response`,
				`text: "Shanghai is sunny and 26C."`,
			},
		},
		{
			demo: "image-generate",
			wants: []string{
				`--- image-generate ---`,
				`https://api.qnaigc.com/v1/images/generations`,
				`"aspect_ratio":"16:9","image_size":"1K"`,
				`results { images: 1 }`,
				`https://example.com/mock/qiniu/images/generations/0.png`,
			},
		},
		{
			demo: "image-generate-simple",
			wants: []string{
				`--- image-generate-simple ---`,
				`https://api.qnaigc.com/v1/images/generations`,
				`"prompt":"一只可爱的橘猫坐在窗台上看着夕阳，照片风格，高清画质"`,
				`results { images: 1 }`,
				`https://example.com/mock/qiniu/images/generations/0.png`,
			},
		},
		{
			demo: "image-generate-portrait",
			wants: []string{
				`--- image-generate-portrait ---`,
				`"model":"gemini-3.1-flash-image-preview"`,
				`"aspect_ratio":"9:16","image_size":"1K"`,
				`results { images: 1 }`,
				`https://example.com/mock/qiniu/images/generations/0.png`,
			},
		},
		{
			demo: "image-edit",
			wants: []string{
				`--- image-edit ---`,
				`https://api.qnaigc.com/v1/images/edits`,
				`"image":["https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg","https://aitoken-public.qnaigc.com/example/generate-video/lawn.jpg"]`,
				`results { images: 1 }`,
				`https://example.com/mock/qiniu/images/edits/0.png`,
			},
		},
		{
			demo: "image-edit-single",
			wants: []string{
				`--- image-edit-single ---`,
				`"image":"https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg"`,
				`"image_size":"1K"`,
				`results { images: 1 }`,
				`https://example.com/mock/qiniu/images/edits/0.png`,
			},
		},
		{
			demo: "image-edit-mask",
			wants: []string{
				`--- image-edit-mask ---`,
				`image-to-image-with-mask-1.jpg`,
				`image-to-image-with-mask-2.png`,
				`results { images: 1 }`,
				`https://example.com/mock/qiniu/images/edits/0.png`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.demo, func(t *testing.T) {
			out := testutil.RunExampleWithEnv(t, "./examples/gemini", geminiMockEnv, tt.demo)
			testutil.RequireContainsAll(t, out, tt.wants...)
		})
	}
}
