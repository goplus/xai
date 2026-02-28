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
	"github.com/goplus/xai"
	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

type response struct {
	*genai.GenerateContentResponse
}

func (p response) Len() int {
	return len(p.Candidates)
}

func (p response) At(i int) xai.Candidate {
	return candidate{p.Candidates[i]}
}

// -----------------------------------------------------------------------------

type candidate struct {
	*genai.Candidate
}

func (p candidate) AsContent() xai.ContentBuilder {
	var parts []*genai.Part
	if c := p.Content; c != nil {
		parts = c.Parts
	}
	return &contentBuilder{parts}
}

// -----------------------------------------------------------------------------
