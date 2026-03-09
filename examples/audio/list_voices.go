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

package main

import (
	"context"
	"fmt"
)

func runListVoices() {
	svc := newService()
	ctx := context.Background()

	voices, err := svc.ListVoices(ctx)
	if err != nil {
		fmt.Println("ListVoices error:", err)
		return
	}

	fmt.Printf("Available voices (%d):\n", len(voices))
	for i, v := range voices {
		fmt.Printf("  [%d] %s (voice_type: %s, category: %s)\n", i+1, v.VoiceName, v.VoiceType, v.Category)
		if v.Url != "" {
			fmt.Printf("      Sample: %s\n", v.Url)
		}
	}
}
