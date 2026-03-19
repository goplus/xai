package image

import (
	"fmt"
	"strings"
	"testing"

	"github.com/goplus/xai/spec/kling/internal"
)

type testParams map[string]any

func (p testParams) Get(name string) (any, bool) {
	v, ok := p[name]
	return v, ok
}

func (p testParams) GetString(name string) string {
	v, ok := p[name]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func TestBuildImageParams_O1RejectsTooManyReferenceImages(t *testing.T) {
	refs := make([]string, 11)
	for i := range refs {
		refs[i] = "https://example.com/ref.jpg"
	}
	_, err := BuildImageParams(internal.ModelKlingImageO1, testParams{
		internal.ParamPrompt:          "an illustration",
		internal.ParamReferenceImages: refs,
	})
	if err == nil || !strings.Contains(err.Error(), "at most 10") {
		t.Fatalf("expected reference_images limit error, got %v", err)
	}
}

func TestBuildImageParams_O1RejectsNAboveDocumentedLimit(t *testing.T) {
	_, err := BuildImageParams(internal.ModelKlingImageO1, testParams{
		internal.ParamPrompt: "an illustration",
		internal.ParamN:      10,
	})
	if err == nil || !strings.Contains(err.Error(), "between 1 and 9") {
		t.Fatalf("expected o1 n limit error, got %v", err)
	}
}

func TestBuildImageParams_RejectsInvalidSubjectImageCount(t *testing.T) {
	tests := []struct {
		model string
		count int
	}{
		{model: internal.ModelKlingV2, count: 1},
		{model: internal.ModelKlingV2, count: 5},
		{model: internal.ModelKlingV21, count: 1},
		{model: internal.ModelKlingV21, count: 5},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s-%d", tt.model, tt.count), func(t *testing.T) {
			subjects := make([]string, tt.count)
			for i := range subjects {
				subjects[i] = "https://example.com/ref.jpg"
			}
			_, err := BuildImageParams(tt.model, testParams{
				internal.ParamPrompt:           "combine subjects",
				internal.ParamSubjectImageList: subjects,
			})
			if err == nil || !strings.Contains(err.Error(), "requires 2 to 4 images") {
				t.Fatalf("expected subject_image_list count error, got %v", err)
			}
		})
	}
}
