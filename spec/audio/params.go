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

package audio

import (
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/audio/internal"
)

// Param name constants (re-exported from internal).
const (
	ParamAudio  = internal.ParamAudio
	ParamModel  = internal.ParamModel
	ParamFormat = internal.ParamFormat
	ParamInput  = internal.ParamInput
	ParamVoice  = internal.ParamVoice
	ParamSpeed  = internal.ParamSpeed
)

// Params implements xai.Params with map-based storage.
type Params struct {
	m map[string]any
}

// NewParams creates a new Params instance.
func NewParams() *Params {
	return &Params{m: make(map[string]any)}
}

// Set sets a parameter. Implements xai.Params.
func (p *Params) Set(name string, val any) xai.Params {
	p.m[name] = val
	return p
}

// Export returns the params as a map for Executor implementations.
func (p *Params) Export() map[string]any {
	out := make(map[string]any, len(p.m))
	for k, v := range p.m {
		out[k] = v
	}
	return out
}

// Get returns the raw value for the given param name and whether it exists.
func (p *Params) Get(name string) (any, bool) {
	v, ok := p.m[name]
	return v, ok
}

// HasNonEmptyString returns true if the param exists and is a non-empty string.
func (p *Params) HasNonEmptyString(name string) bool {
	v, ok := p.m[name]
	if !ok {
		return false
	}
	s, ok := v.(string)
	return ok && strings.TrimSpace(s) != ""
}

// GetString returns the trimmed string value of the param, or "" if missing or not a string.
func (p *Params) GetString(name string) string {
	v, ok := p.m[name]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}
