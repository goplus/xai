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

package geno

import (
	"iter"

	"github.com/goplus/xai"
)

// -----------------------------------------------------------------------------

func (p *Service[T]) Features() xai.Feature {
	return xai.FeatureOperation
}

func (p *Service[T]) Gen(params xai.GenParams) (xai.GenResponse, error) {
	panic("unsupported")
}

func (p *Service[T]) GenStream(params xai.GenParams) iter.Seq2[xai.GenResponse, error] {
	panic("unsupported")
}

func (p *Service[T]) GenParams() xai.GenParams {
	panic("unsupported")
}

func (p *Service[T]) Images() xai.ImageBuilder {
	panic("unsupported")
}

func (p *Service[T]) Docs() xai.DocumentBuilder {
	panic("unsupported")
}

func (p *Service[T]) UserMsg() xai.MsgBuilder {
	panic("unsupported")
}

func (p *Service[T]) AssistantMsg() xai.MsgBuilder {
	panic("unsupported")
}

func (p *Service[T]) WebSearchTool() xai.WebSearchTool {
	panic("unsupported")
}

func (p *Service[T]) ToolDef(name string) xai.Tool {
	panic("unsupported")
}

func (p *Service[T]) Tool(name string) xai.Tool {
	panic("unsupported")
}

// -----------------------------------------------------------------------------
