package video

import (
	"testing"

	"github.com/goplus/xai/mock/testutil"
)

var viduMockEnv = map[string]string{
	"QINIU_API_KEY":   "",
	"QINIU_MOCK_CURL": "",
}

func TestViduVideoExamplesMock(t *testing.T) {
	tests := []struct {
		demo  string
		wants []string
	}{
		{
			demo: "q1-text",
			wants: []string{
				`--- q1-text ---`,
				`[q1-text] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/vidu-q1/text_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q1-ref-urls",
			wants: []string{
				`--- q1-ref-urls ---`,
				`[q1-ref-urls] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/vidu-q1/reference_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q1-ref-subjects",
			wants: []string{
				`--- q1-ref-subjects ---`,
				`[q1-ref-subjects] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/vidu-q1/reference_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q1-ref-subjects-audio",
			wants: []string{
				`--- q1-ref-subjects-audio ---`,
				`[q1-ref-subjects-audio] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/vidu-q1/reference_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q2-text",
			wants: []string{
				`--- q2-text ---`,
				`[q2-text] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/vidu-q2/text_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q2-ref-urls",
			wants: []string{
				`--- q2-ref-urls ---`,
				`[q2-ref-urls] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/vidu-q2/reference_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q2-ref-subjects",
			wants: []string{
				`--- q2-ref-subjects ---`,
				`[q2-ref-subjects] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/vidu-q2/reference_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q2-image-pro",
			wants: []string{
				`--- q2-image-pro ---`,
				`[q2-image] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/vidu-q2/image_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q2-image-pro-audio",
			wants: []string{
				`--- q2-image-pro-audio ---`,
				`[q2-image-pro-audio] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/viduq2-pro/image_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q2-image-turbo",
			wants: []string{
				`--- q2-image-turbo ---`,
				`[q2-image-turbo] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/viduq2-turbo/image_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q2-start-end-pro",
			wants: []string{
				`--- q2-start-end-pro ---`,
				`[q2-start-end] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/vidu-q2/start_end_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q3-text-turbo",
			wants: []string{
				`--- q3-text-turbo ---`,
				`[q3-text-turbo] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/viduq3-turbo/text_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q3-image-turbo",
			wants: []string{
				`--- q3-image-turbo ---`,
				`[q3-image-turbo] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/viduq3-turbo/image_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q3-start-end-turbo",
			wants: []string{
				`--- q3-start-end-turbo ---`,
				`[q3-start-end-turbo] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/viduq3-turbo/start_end_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q3-text-pro",
			wants: []string{
				`--- q3-text-pro ---`,
				`[q3-text-pro] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/viduq3-pro/text_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q3-image-pro",
			wants: []string{
				`--- q3-image-pro ---`,
				`[q3-image-pro] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/viduq3-pro/image_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "q3-start-end-pro",
			wants: []string{
				`--- q3-start-end-pro ---`,
				`[q3-start-end-pro] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/viduq3-pro/start_end_to_video/mock-vidu-task-1.mp4`,
			},
		},
		{
			demo: "call-sync",
			wants: []string{
				`--- call-sync ---`,
				`TaskID saved to DB: mock-vidu-task-1`,
				`[call-sync] polling task: mock-vidu-task-1`,
				`https://example.com/mock/vidu/vidu-q2/text_to_video/mock-vidu-task-1.mp4`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.demo, func(t *testing.T) {
			out := testutil.RunExampleWithEnv(t, "./examples/vidu/video", viduMockEnv, tt.demo)
			testutil.RequireContainsAll(t, out, tt.wants...)
		})
	}
}
