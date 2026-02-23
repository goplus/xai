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

// Role is the type of a message role, which can be system, assistant, user or tool.
type Role string

const (
	// RoleSystem is the role of a system, means the message is a system message.
	RoleSystem Role = "system"

	// RoleAssistant is the role of an assistant, means the message is returned by AI.
	RoleAssistant Role = "assistant"

	// RoleUser is the role of a user, means the message is a user message.
	RoleUser Role = "user"

	// RoleTool is the role of a tool, means the message is a tool call output.
	RoleTool Role = "tool"
)

// -----------------------------------------------------------------------------

type Message struct {
	Role    Role
	Content Content
}

// -----------------------------------------------------------------------------

type Option struct {
}

// -----------------------------------------------------------------------------

// Chatter is the interface for chatting with AI. It takes a list of messages as
// input and returns a message as output.
type Chatter interface {
	Chat(ctx context.Context, in []Message, opts ...Option) (out Message, err error)
}

// -----------------------------------------------------------------------------
