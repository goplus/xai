/*
 * Copyright (c) 2026 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gemini

import (
	"fmt"
	"testing"
	"unsafe"

	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

func TestSizeofImage(t *testing.T) {
	if unsafe.Sizeof(genai.ProductImage{}) != unsafe.Sizeof((*genai.Image)(nil)) {
		t.Fatal("size of genai.ProductImage is not equal to size of *genai.Image")
	}
	if unsafe.Sizeof(genai.ScribbleImage{}) != unsafe.Sizeof((*genai.Image)(nil)) {
		t.Fatal("size of genai.ScribbleImage is not equal to size of *genai.Image")
	}
}

// -----------------------------------------------------------------------------

const (
	genVideoSchema       = "[{Prompt 4} {Image 5} {Video 7} {NumberOfVideos 2} {OutputStgUri 4} {FPS 2} {DurationSeconds 2} {Seed 2} {AspectRatio 4} {Resolution 4} {PersonGeneration 4} {PubsubTopic 4} {NegativePrompt 4} {EnhancePrompt 1} {GenerateAudio 1} {LastFrame 5} {ReferenceImages 32776} {Mask 9} {CompressionQuality 4}]"
	genImageSchema       = "[{Prompt 4} {OutputStgUri 4} {NegativePrompt 4} {NumberOfImages 2} {AspectRatio 4} {GuidanceScale 3} {Seed 2} {SafetyFilterLevel 4} {PersonGeneration 4} {IncludeSafetyAttributes 1} {IncludeRAIReason 1} {Language 4} {OutputMIMEType 4} {OutputCompressionQuality 2} {AddWatermark 1} {ImageSize 4} {EnhancePrompt 1}]"
	editImageSchema      = "[{Prompt 4} {References 32774} {OutputStgUri 4} {NegativePrompt 4} {NumberOfImages 2} {AspectRatio 4} {GuidanceScale 3} {Seed 2} {SafetyFilterLevel 4} {PersonGeneration 4} {IncludeSafetyAttributes 1} {IncludeRAIReason 1} {Language 4} {OutputMIMEType 4} {OutputCompressionQuality 2} {AddWatermark 1} {EditMode 4} {BaseSteps 2}]"
	recontextImageSchema = "[{Prompt 4} {PersonImage 5} {ProductImages 32773} {NumberOfImages 2} {BaseSteps 2} {OutputStgUri 4} {Seed 2} {SafetyFilterLevel 4} {PersonGeneration 4} {AddWatermark 1} {OutputMIMEType 4} {OutputCompressionQuality 2} {EnhancePrompt 1}]"
	segmentImageSchema   = "[{Prompt 4} {Image 5} {ScribbleImage 5} {Mode 4} {MaxPredictions 2} {ConfidenceThreshold 3} {MaskDilation 3} {BinaryColorThreshold 3}]"
	upscaleImageSchema   = "[{Image 5} {Factor 4} {OutputStgUri 4} {SafetyFilterLevel 4} {PersonGeneration 4} {IncludeRAIReason 1} {OutputMIMEType 4} {OutputCompressionQuality 2} {EnhanceInputImage 1} {ImagePreservationFactor 3}]"
)

type schemaTestCase struct {
	v    any
	want string
}

func TestInputSchema(t *testing.T) {
	cases := []schemaTestCase{
		{new(genVideo), genVideoSchema},
		{new(genImage), genImageSchema},
		{new(editImage), editImageSchema},
		{new(recontextImage), recontextImageSchema},
		{new(segmentImage), segmentImageSchema},
		{new(upscaleImage), upscaleImageSchema},
	}
	for _, c := range cases {
		flds := fmt.Sprint(newInputSchema(c.v, nil).Fields())
		if flds != c.want {
			t.Fatalf("TestInputSchema failed: %T - %v\n", c.v, flds)
		}
	}
}

const (
	genVideoRespSchema       = "[{GeneratedVideos 32781} {RAIMediaFilteredCount 2} {RAIMediaFilteredReasons 32772}]"
	genImageRespSchema       = "[{GeneratedImages 32779} {PositivePromptSafetyAttributes 10}]"
	editImageRespSchema      = "[{GeneratedImages 32779}]"
	recontextImageRespSchema = "[{GeneratedImages 32779}]"
	segmentImageRespSchema   = "[{GeneratedMasks 32780}]"
	upscaleImageRespSchema   = "[{GeneratedImages 32779}]"
)

func TestOutputSchema(t *testing.T) {
	cases := []schemaTestCase{
		{new(genai.GenerateVideosResponse), genVideoRespSchema},
		{new(genai.GenerateImagesResponse), genImageRespSchema},
		{new(genai.EditImageResponse), editImageRespSchema},
		{new(genai.RecontextImageResponse), recontextImageRespSchema},
		{new(genai.SegmentImageResponse), segmentImageRespSchema},
		{new(genai.UpscaleImageResponse), upscaleImageRespSchema},
	}
	for _, c := range cases {
		flds := fmt.Sprint(newInputSchema(c.v, nil).Fields())
		if flds != c.want {
			t.Fatalf("TestOutputSchema failed: %T - %v\n", c.v, flds)
		}
	}
}

// -----------------------------------------------------------------------------

const (
	maskSchema    = "[{MaskMode 4} {SegmentationClasses 32770} {MaskDilation 3}]"
	controlSchema = "[{ControlType 4} {EnableControlImageComputation 1}]"
	styleSchema   = "[{StyleDescription 4}]"
	subjectSchema = "[{SubjectType 4} {SubjectDescription 4}]"
)

func TestConfSchema(t *testing.T) {
	cases := []schemaTestCase{
		{new(genai.MaskReferenceConfig), maskSchema},
		{new(genai.ControlReferenceConfig), controlSchema},
		{new(genai.StyleReferenceConfig), styleSchema},
		{new(genai.SubjectReferenceConfig), subjectSchema},
	}
	for _, c := range cases {
		flds := fmt.Sprint(newInputSchema(c.v, nil).Fields())
		if flds != c.want {
			t.Fatalf("TestConfSchema failed: %T - %v\n", c.v, flds)
		}
	}
}

// -----------------------------------------------------------------------------
