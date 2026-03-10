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
)

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
	Retry(ctx context.Context, svc Service, opts OptionBuilder) (OperationResponse, error)

	// Results returns the result from the operation.
	Results() Results
}

// Wait is a helper function that waits for an `OperationResponse` to be done by
// repeatedly calling `Retry` with appropriate sleeping in between. Once the operation
// is done, it returns the results of the operation.
func Wait(ctx context.Context, svc Service, resp OperationResponse, progress func(OperationResponse), opts OptionBuilder) (ret Results, err error) {
	for !resp.Done() {
		if progress != nil {
			progress(resp)
		}
		resp.Sleep()
		resp, err = resp.Retry(ctx, svc, opts)
		if err != nil {
			return
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

// Call is a helper function that calls an `Operation` with the given options, and then
// waits for the operation to be done. It returns the results of the operation once it's
// completed.
func Call(ctx context.Context, svc Service, op Operation, opts OptionBuilder, progress func(OperationResponse)) (ret Results, err error) {
	resp, err := op.Call(ctx, svc, opts)
	if err != nil {
		return
	}
	return Wait(ctx, svc, resp, progress, opts)
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
