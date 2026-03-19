package video

import (
	"testing"

	"github.com/goplus/xai/mock/kling/testutil"
)

func TestKlingVideoExamplesMockCurl(t *testing.T) {
	tests := []struct {
		demo  string
		wants []string
	}{
		{
			demo: "kling-v2-1",
			wants: []string{
				`"model":"kling-v2-1","prompt":"镜头缓慢右移，人物开始奔跑","seconds":"5","size":"1280x720"`,
				`"negative_prompt":"blurry, low quality, jittery, unstable"`,
				`"image_tail":"https://picsum.photos/1280/720","input_reference":"https://picsum.photos/1280/720"`,
				`"seconds":"10","size":"1920x1080"`,
			},
		},
		{
			demo: "kling-v2-5-turbo",
			wants: []string{
				`"model":"kling-v2-5-turbo","prompt":"一只可爱的小猫在草地上玩耍"`,
				`"negative_prompt":"blurry, low quality, distorted"`,
				`"image_tail":"https://picsum.photos/1280/720","input_reference":"https://picsum.photos/1280/720"`,
				`"seconds":"10","size":"1920x1080"`,
				`"size":"1080x1080"`,
			},
		},
		{
			demo: "kling-v2-6",
			wants: []string{
				`"negative_prompt":"blurry, low quality, distorted, ugly"`,
				`"character_orientation":"image"`,
				`"keep_original_sound":"yes"`,
				`"character_orientation":"video"`,
				`"keep_original_sound":"no"`,
				`"sound":"on"`,
				`"size":"1080x1920"`,
				`"seconds":"10","size":"1920x1080"`,
			},
		},
		{
			demo: "kling-video-o1",
			wants: []string{
				`"model":"kling-video-o1","prompt":"一只可爱的橘猫在阳光下奔跑，慢镜头，电影质感"`,
				`"image_list":[{"image":"https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg"}]`,
				`"image_list":[{"image":"https://picsum.photos/1280/720","type":"first_frame"}]`,
				`"image_list":[{"image":"https://picsum.photos/1280/720","type":"first_frame"},{"image":"https://picsum.photos/1280/720","type":"end_frame"}]`,
				`"video_list":[{"keep_original_sound":"yes","refer_type":"feature","video_url":"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"}]`,
				`"video_list":[{"refer_type":"base","video_url":"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"}]`,
			},
		},
		{
			demo: "kling-v3",
			wants: []string{
				`"model":"kling-v3","prompt":"一只可爱的小猫在阳光下玩耍"`,
				`"input_reference":"https://picsum.photos/1280/720"`,
				`"sound":"on"`,
				`"multi_prompt":[{"duration":"3","index":1,"prompt":"清晨，阳光照进窗台，城市逐渐苏醒"},{"duration":"4","index":2,"prompt":"午后，街道人流与车流交织，节奏加快"},{"duration":"3","index":3,"prompt":"夜晚，霓虹点亮天际线，镜头拉远收束"}],"multi_shot":true`,
				`"shot_type":"customize"`,
			},
		},
		{
			demo: "kling-v3-omni",
			wants: []string{
				`"model":"kling-v3-omni","multi_prompt":[{"duration":"4","index":1,"prompt":"产品全景展示，旋转"},{"duration":"3","index":2,"prompt":"产品细节特写"},{"duration":"3","index":3,"prompt":"产品使用场景"}],"multi_shot":true`,
				`"image_list":[{"image":"https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg"}]`,
				`"image_list":[{"image":"https://picsum.photos/1280/720","type":"first_frame"},{"image":"https://picsum.photos/1280/720","type":"end_frame"}]`,
				`"video_list":[{"keep_original_sound":"yes","refer_type":"feature","video_url":"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"}]`,
				`"sound":"on"`,
				`"shot_type":"auto"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.demo, func(t *testing.T) {
			out := testutil.RunExample(t, "./examples/kling/video", tt.demo)
			testutil.RequireContainsAll(t, out, tt.wants...)
		})
	}
}
