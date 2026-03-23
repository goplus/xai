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

import (
	"context"
	"time"
)

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

// WaitParams represents the parameters that can be set when waiting for an `Operation`
// to complete.
type WaitParams interface {
	// BaseURL sets the base URL for the API endpoint.
	BaseURL(string) WaitParams

	// Timeout sets a timeout for the API request. If the request takes longer than
	// the specified duration, it will be aborted and an error will be returned.
	Timeout(time.Duration) WaitParams

	// Progress sets a callback function that will be called with the current
	// `OperationResponse` each time the operation status is checked. This can be
	// used to provide progress updates to the user while waiting for the operation
	// to complete.
	Progress(func(OperationResponse)) WaitParams
}

// OperationResponse represents the response from an `Operation`. It provides methods
// to check the status of the operation, retrieve results when it's done.
type OperationResponse interface {
	// Done returns true if the operation is completed.
	Done() bool

	// Results returns the result from the operation.
	Results() Results

	// WaitParams returns a `WaitParams` that can be used to set parameters for waiting
	// on the operation.
	WaitParams() WaitParams

	// Wait waits for the operation to be completed. It repeatedly checks the status
	// of the operation and calls the provided progress function with the current
	// operation response. Once the operation is done, it returns the results.
	Wait(ctx context.Context, __xgo_optional_params WaitParams) (Results, error)
}

type CallParams interface {
	// Set sets a parameter for the operation. You can call this method multiple
	// times to set multiple parameters.
	Set(name string, val any) CallParams

	// BaseURL sets the base URL for the API endpoint.
	BaseURL(string) CallParams

	// Timeout sets a timeout for the API request. If the request takes longer than
	// the specified duration, it will be aborted and an error will be returned.
	Timeout(time.Duration) CallParams
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

	// CallParams creates a `CallParams` instance that can be used to set parameters for
	// this operation. You can use the `Set` method of `CallParams` to set parameters by
	// name and value, and then pass the `CallParams` to the `Call` method to start the
	// operation.
	CallParams() CallParams

	// Call starts the operation with the given options. It returns an `OperationResponse`
	// that can be used to check the status of the operation and retrieve results when
	// it's done.
	Call(ctx context.Context, params CallParams) (OperationResponse, error)
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
