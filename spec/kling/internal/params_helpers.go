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
 * WITHOUT WARRANTIES OR CONDITIONS OF KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package internal

import "strings"

// GetBool returns the bool value for the given param name, or false if missing/invalid.
func GetBool(p ParamsReader, name string) bool {
	v, ok := p.Get(name)
	if !ok {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	}
	return false
}

// GetInt returns the int value for the given param name, or 1 if missing/invalid.
func GetInt(p ParamsReader, name string) int {
	v, ok := p.Get(name)
	if !ok {
		return 1
	}
	switch x := v.(type) {
	case int:
		if x > 0 {
			return x
		}
	case int64:
		if x > 0 {
			return int(x)
		}
	case float64:
		if x > 0 {
			return int(x)
		}
	}
	return 1
}

// GetFloat64Ptr returns a pointer to the float64 value, or nil if missing/invalid.
func GetFloat64Ptr(p ParamsReader, name string) *float64 {
	v, ok := p.Get(name)
	if !ok {
		return nil
	}
	switch x := v.(type) {
	case float64:
		return &x
	case int:
		f := float64(x)
		return &f
	case int64:
		f := float64(x)
		return &f
	}
	return nil
}

// GetStringSlice returns the string slice for the given param name.
func GetStringSlice(p ParamsReader, name string) []string {
	v, ok := p.Get(name)
	if !ok {
		return nil
	}
	switch x := v.(type) {
	case string:
		if s := strings.TrimSpace(x); s != "" {
			return []string{s}
		}
		return nil
	case []string:
		return x
	case []interface{}:
		var out []string
		for _, item := range x {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

// GetSubjectImageList returns the subject image URLs for the given param name.
// Supports []string or []interface{} with elements as string or map[string]string
// containing "subject_image" key.
func GetSubjectImageList(p ParamsReader, name string) []string {
	v, ok := p.Get(name)
	if !ok {
		return nil
	}
	switch x := v.(type) {
	case []string:
		return x
	case []interface{}:
		var out []string
		for _, item := range x {
			switch elem := item.(type) {
			case string:
				if s := strings.TrimSpace(elem); s != "" {
					out = append(out, s)
				}
			case map[string]interface{}:
				if sub, ok := elem[ParamSubjectImage].(string); ok && strings.TrimSpace(sub) != "" {
					out = append(out, sub)
				}
			case map[string]string:
				if sub := elem[ParamSubjectImage]; strings.TrimSpace(sub) != "" {
					out = append(out, sub)
				}
			}
		}
		return out
	}
	return nil
}
