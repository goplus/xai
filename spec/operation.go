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

package xai

import "context"

// -----------------------------------------------------------------------------

// Action represents a specific operation that can be performed with a model, such as
// generating a video, editing an image, etc. The available actions may vary depending
// on the model and service being used. You can use the `Actions` method of a `Service`
// to get the list of supported actions for a given model, and then use the `Operation`
// method to get an `Operation` instance for a specific action.
type Action string

const (
	GenVideo       Action = "gen_video"
	GenImage       Action = "gen_image"
	EditImage      Action = "edit_image"
	RecontextImage Action = "recontext_image"
	SegmentImage   Action = "segment_image"
	UpscaleImage   Action = "upscale_image"

	// Audio actions
	Transcribe Action = "transcribe" // ASR: audio -> text
	Synthesize Action = "synthesize" // TTS: text -> audio
)

// VideoGenMode represents the video generation mode for GenVideo action.
// Different models support different modes based on input types.
//
// Mode hierarchy (higher modes imply lower modes):
//   - text_to_video: base capability, all video models support this
//   - image_to_video: single image input, implies text_to_video
//   - multi_ref_to_video: multiple reference images, implies image_to_video
//   - start_end_to_video: first + last frame, implies image_to_video
type VideoGenMode string

const (
	VideoGenModeText     VideoGenMode = "text_to_video"      // Text prompt only (base)
	VideoGenModeImage    VideoGenMode = "image_to_video"     // Single image + prompt
	VideoGenModeMultiRef VideoGenMode = "multi_ref_to_video" // Multiple reference images + prompt
	VideoGenModeStartEnd VideoGenMode = "start_end_to_video" // First frame + last frame + prompt
)

// AllVideoGenModes returns all defined video generation modes.
func AllVideoGenModes() []VideoGenMode {
	return []VideoGenMode{
		VideoGenModeText,
		VideoGenModeImage,
		VideoGenModeMultiRef,
		VideoGenModeStartEnd,
	}
}

// VideoGenModeImplies returns the modes that are implied by the given mode.
// For example, multi_ref_to_video implies image_to_video and text_to_video.
func VideoGenModeImplies(mode VideoGenMode) []VideoGenMode {
	switch mode {
	case VideoGenModeText:
		return nil
	case VideoGenModeImage:
		return []VideoGenMode{VideoGenModeText}
	case VideoGenModeMultiRef:
		return []VideoGenMode{VideoGenModeImage, VideoGenModeText}
	case VideoGenModeStartEnd:
		return []VideoGenMode{VideoGenModeImage, VideoGenModeText}
	default:
		return nil
	}
}

// ExpandVideoGenModes expands a list of modes to include all implied modes.
// For example, [multi_ref_to_video] expands to [multi_ref_to_video, image_to_video, text_to_video].
func ExpandVideoGenModes(modes []VideoGenMode) []VideoGenMode {
	seen := make(map[VideoGenMode]bool)
	var result []VideoGenMode

	var add func(m VideoGenMode)
	add = func(m VideoGenMode) {
		if seen[m] {
			return
		}
		seen[m] = true
		result = append(result, m)
		for _, implied := range VideoGenModeImplies(m) {
			add(implied)
		}
	}

	for _, m := range modes {
		add(m)
	}
	return result
}

// NormalizeVideoGenModes returns the minimal set of modes that represent the same capabilities.
// For example, [multi_ref_to_video, image_to_video, text_to_video] normalizes to [multi_ref_to_video].
func NormalizeVideoGenModes(modes []VideoGenMode) []VideoGenMode {
	if len(modes) == 0 {
		return nil
	}

	modeSet := make(map[VideoGenMode]bool)
	for _, m := range modes {
		modeSet[m] = true
	}

	var result []VideoGenMode
	for _, m := range modes {
		isImplied := false
		for other := range modeSet {
			if other == m {
				continue
			}
			for _, implied := range VideoGenModeImplies(other) {
				if implied == m {
					isImplied = true
					break
				}
			}
			if isImplied {
				break
			}
		}
		if !isImplied {
			result = append(result, m)
		}
	}
	return result
}

// -----------------------------------------------------------------------------

// VideoSchema describes the schema for video generation operations.
// Each provider (vidu, kling, gemini, etc.) implements this interface to describe
// the supported modes, fields, and their restrictions for a specific model.
type VideoSchema interface {
	// SupportedModes returns the video generation modes supported by this model.
	SupportedModes() []VideoGenMode

	// Fields returns all input fields for this model.
	Fields() []Field

	// Restrict returns the restriction for a field.
	// Returns nil if there is no restriction for the field.
	Restrict(name string) *Restriction

	// FieldModes returns the modes that a field is applicable to.
	// Returns nil if the field is applicable to all modes (common fields like prompt).
	FieldModes(name string) []VideoGenMode
}

// VideoSchemaProvider is implemented by services that support video generation.
// Use this interface to get the VideoSchema for a specific model.
type VideoSchemaProvider interface {
	// VideoSchema returns the VideoSchema for the given model.
	// Returns nil if the model does not support video generation.
	VideoSchema(model Model) VideoSchema
}

// -----------------------------------------------------------------------------

// Results represents the results of an `Operation`.
type Results interface {
	// XGo_Attr ($name) retrieves a property value from the results by name.
	XGo_Attr(name string) any

	// Len returns the number of generated images or videos.
	Len() int

	// At retrieves a generated image or video from the results by index.
	// For GenVideo, returns *OutputVideo;
	// For SegmentImage, returns *OutputImageMask;
	// For GenImage, EditImage, RecontextImage, UpscaleImage, returns *OutputImage.
	At(i int) Generated
}

// OperationResponse represents the response from an `Operation`. It provides methods
// to check the status of the operation, retrieve results when it's done.
type OperationResponse interface {
	// Done returns true if the operation is completed.
	Done() bool

	// Sleep sleeps a suggested amount of time before the next retry.
	Sleep()

	// Retry retries the operation. It returns a new `OperationResponse` that can be
	// used to check the status of the operation and retrieve results when it's done.
	// You can call this method multiple times to keep retrying until the operation
	// is done.
	Retry(ctx context.Context, svc Service) (OperationResponse, error)

	// Results returns the result from the operation.
	Results() Results

	// TaskID returns the task ID for async operations. Empty for sync (immediate) responses.
	// Callers can persist this to DB and later use GetTask to resume polling.
	TaskID() string
}

// OperationResponseWithError is an optional interface. When implemented, GetError
// returns the error when the operation failed (e.g. content policy violation).
type OperationResponseWithError interface {
	OperationResponse
	GetError() error
}

// Wait is a helper function that waits for an `OperationResponse` to be done by
// repeatedly calling `Retry` with appropriate sleeping in between. Once the operation
// is done, it returns the results of the operation.
func Wait(ctx context.Context, svc Service, resp OperationResponse, progress func(OperationResponse)) (ret Results, err error) {
	for !resp.Done() {
		if progress != nil {
			progress(resp)
		}
		resp.Sleep()
		resp, err = resp.Retry(ctx, svc)
		if err != nil {
			return
		}
	}
	if errResp, ok := resp.(OperationResponseWithError); ok {
		if opErr := errResp.GetError(); opErr != nil {
			return resp.Results(), opErr
		}
	}
	return resp.Results(), nil
}

// Operation represents a long-running task that may take some time to complete, such as
// generating a video or editing an image. You can use an `Operation` to set parameters
// for the action and then call it with a prompt to start the operation.
type Operation interface {
	// InputSchema returns the schema for the input parameters of this operation. This
	// schema defines the parameters that can be set for this operation, such as the
	// type and name of each parameter. You can use this schema to understand what
	// parameters are required or optional for this operation, and to set them correctly
	// before calling the operation.
	InputSchema() InputSchema

	// Params returns a `Params` that can be used to set parameters for the operation.
	Params() Params

	// Call starts the operation with the given options. It returns an `OperationResponse`
	// that can be used to check the status of the operation and retrieve results when
	// it's done.
	Call(ctx context.Context, svc Service, opts OptionBuilder) (OperationResponse, error)
}

// CallSync starts the operation and returns the OperationResponse. Callers can then
// use Wait to poll until the operation is done.
func CallSync(ctx context.Context, svc Service, op Operation, opts OptionBuilder) (OperationResponse, error) {
	return op.Call(ctx, svc, opts)
}

// Call is a helper function that calls an `Operation` with the given options, and then
// waits for the operation to be done. It returns the results of the operation once it's
// completed.
func Call(ctx context.Context, svc Service, op Operation, opts OptionBuilder, progress func(OperationResponse)) (ret Results, err error) {
	resp, err := CallSync(ctx, svc, op, opts)
	if err != nil {
		return
	}
	return Wait(ctx, svc, resp, progress)
}

// GetTask returns an OperationResponse for an existing task by taskID.
// Use when resuming from DB. Returns error if the service does not support it.
func GetTask(ctx context.Context, svc Service, model Model, action Action, taskID string) (OperationResponse, error) {
	if tr, ok := svc.(interface {
		GetTask(ctx context.Context, model Model, action Action, taskID string) (OperationResponse, error)
	}); ok {
		return tr.GetTask(ctx, model, action, taskID)
	}
	return nil, ErrNotSupported
}

// -----------------------------------------------------------------------------

type operationService interface {
	// Actions returns the list of supported actions for the given model.
	Actions(model Model) []Action

	// Operation returns an `Operation` that can be used to perform the specified action
	// with the given model. An `Operation` represents a long-running task that may take
	// some time to complete, such as generating a video or editing an image. You can
	// use the returned `Operation` to set parameters for the action and then call it
	// with a prompt to start the operation. The `OperationResponse` can then be used
	// to check the status of the operation and retrieve results when it's done.
	Operation(model Model, action Action) (Operation, error)
}

// -----------------------------------------------------------------------------
