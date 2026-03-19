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
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/goplus/dql/klingai/model"
)

// -----------------------------------------------------------------------------

type modelParam struct {
	Model       string
	Required    string
	Defval      string   `json:",omitempty"`
	Description string   `json:",omitempty"`
	Notes       string   `json:",omitempty"`
	Enumvals    []string `json:",omitempty"`
}

type paramInfo struct {
	Name   string
	Type   string
	Models []*modelParam
}

type none struct{}

var imageModelSels = map[string]none{
	"imageGeneration":   {},
	"OmniImage":         {},
	"multiImageToImage": {},
}

var knownParams = map[string]string{
	"model_name":   "",
	"prompt":       "Prompt",
	"n":            "NumberOfImages",
	"aspect_ratio": "AspectRatio",
	"resolution":   "ImageSize",
}

var knownParamTypes = map[string]string{
	"image_list": "[]xai.Image",
}

func main() {
	b, err := os.ReadFile("../../../spec/kling/klingai.json")
	check(err)

	var ret []*model.Result
	err = json.Unmarshal(b, &ret)
	check(err)

	var paramExists = make(map[string]*paramInfo)
	var params []*paramInfo
	for _, r := range ret {
		if _, ok := imageModelSels[r.Model]; ok {
			for _, item := range r.APIs[0].Req.Body {
				if _, ok := knownParams[item.Name]; ok {
					continue
				}
				model := &modelParam{
					Model:       r.Model,
					Required:    item.Required,
					Defval:      item.Defval,
					Description: item.Description,
					Notes:       item.Notes,
					Enumvals:    item.Enumvals,
				}
				param, ok := paramExists[item.Name]
				if ok {
					if param.Type != item.Type {
						log.Panicf("param %s type mismatch: %s vs %s\n", item.Name, param.Type, item.Type)
					}
					param.Models = append(param.Models, model)
				} else {
					param = &paramInfo{
						Name:   item.Name,
						Type:   item.Type,
						Models: []*modelParam{model},
					}
					paramExists[item.Name] = param
					params = append(params, param)
				}
			}
		}
	}
	for _, param := range params {
		name := param.Name
		typ := param.Type
		if typ == "array" || typ == "object" {
			if t, ok := knownParamTypes[name]; ok {
				typ = t
			} else {
				log.Panicf("type of param %s is unknown\n", name)
			}
		}
		fmt.Print(name, " ", typ, ":")
		for _, model := range param.Models {
			fmt.Print(" ", model.Model)
		}
		fmt.Println("  //", param.Models[0].Description)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// -----------------------------------------------------------------------------
