package images

import (
	"testing"

	"github.com/goplus/xai/mock/kling/testutil"
)

func TestKlingImageExamplesMockCurl(t *testing.T) {
	tests := []struct {
		demo  string
		wants []string
	}{
		{
			demo: "call-sync",
			wants: []string{
				`--- call-sync ---`,
				`"model":"kling-v1","n":1,"prompt":"一只可爱的橘猫坐在窗台上看着夕阳,照片风格,高清画质"`,
				`Error: mock mode enabled`,
			},
		},
		{
			demo: "kling-v1",
			wants: []string{
				`"image":"https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg","model":"kling-v1"`,
				`"negative_prompt":"模糊,低质量,变形"`,
				`"model":"kling-v1","n":2,"prompt":"一只可爱的橘猫"`,
			},
		},
		{
			demo: "kling-v1-5",
			wants: []string{
				`"image_reference":"subject"`,
				`"human_fidelity":0.6`,
				`"negative_prompt":"模糊,低质量"`,
			},
		},
		{
			demo: "kling-v2",
			wants: []string{
				`"image":"https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg","model":"kling-v2"`,
				`'https://api.qnaigc.com/v1/images/edits'`,
				`"subject_image_list":[{"subject_image":"https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png"},{"subject_image":"https://aitoken-public.qnaigc.com/example/generate-image/image-to-image-with-mask-1.jpg"}]`,
				`"negative_prompt":"模糊,低质量"`,
			},
		},
		{
			demo: "kling-v2-new",
			wants: []string{
				`"model":"kling-v2-new","n":1,"prompt":"将这张图片转换为赛博朋克风格"`,
				`"prompt":"将这张图片转换为中国水墨画风格"`,
				`"prompt":"将这张图片转换为梵高星空油画风格"`,
			},
		},
		{
			demo: "kling-v2-1",
			wants: []string{
				`"model":"kling-v2-1","n":1,"prompt":"a sunset over the ocean, cinematic lighting"`,
				`"reference_images":["https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg"]`,
				`"scene_image":"https://aitoken-public.qnaigc.com/example/generate-image/image-to-image-with-mask-1.jpg"`,
				`"style_image":"https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png"`,
			},
		},
		{
			demo: "kling-image-o1",
			wants: []string{
				`'https://api.qnaigc.com/queue/fal-ai/kling-image/o1'`,
				`"resolution":"2K"`,
				`"image_urls":["https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg"]`,
				`"num_images":2`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.demo, func(t *testing.T) {
			out := testutil.RunExample(t, "./examples/kling/images", tt.demo)
			testutil.RequireContainsAll(t, out, tt.wants...)
		})
	}
}
