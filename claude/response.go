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

package claude

import (
	"iter"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
	"github.com/goplus/xai"
)

// -----------------------------------------------------------------------------

type message struct {
	msg *anthropic.BetaMessage
}

func (p message) AsContent() xai.ContentBuilder {
	content := make([]anthropic.BetaContentBlockParamUnion, len(p.msg.Content))
	for i, c := range p.msg.Content {
		content[i] = c.ToParam()
	}
	return &contentBuilder{content}
}

// -----------------------------------------------------------------------------

func buildMsgIter(stream *ssestream.Stream[anthropic.BetaRawMessageStreamEventUnion]) iter.Seq2[xai.Message, error] {
	panic("todo")
}

// -----------------------------------------------------------------------------
