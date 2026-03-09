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

package qiniu

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/url"
	"path"
	"strconv"
	"strings"

	xai "github.com/goplus/xai/spec"
	"google.golang.org/genai"
)

const (
	endpointChatCompletions = "chat/completions"
	endpointImagesGenerate  = "images/generations"
	endpointImagesEdit      = "images/edits"
)

var errStreamStopped = errors.New("qiniu: stream stopped")

type backend struct {
	client *client
}

func newBackend(cli *client) *backend {
	return &backend{client: cli}
}

func (p *backend) Actions(model xai.Model) []xai.Action {
	if !isGeminiModel(model) {
		return nil
	}
	return []xai.Action{xai.GenImage, xai.EditImage}
}

func isGeminiModel(model xai.Model) bool {
	return strings.Contains(strings.ToLower(string(model)), "gemini")
}

func (p *backend) GenerateContent(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	body, err := buildChatRequest(model, contents, config, false)
	if err != nil {
		return nil, err
	}
	var resp chatCompletionResponse
	if err := p.client.postJSONAt(ctx, p.baseURL(genContentHTTPOptions(config)), endpointChatCompletions, body, &resp); err != nil {
		return nil, err
	}
	return resp.toGenerateContent()
}

func (p *backend) GenerateContentStream(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) iter.Seq2[*genai.GenerateContentResponse, error] {
	body, err := buildChatRequest(model, contents, config, true)
	if err != nil {
		return func(yield func(*genai.GenerateContentResponse, error) bool) {
			yield(nil, err)
		}
	}
	base := p.baseURL(genContentHTTPOptions(config))
	return func(yield func(*genai.GenerateContentResponse, error) bool) {
		stream, err := p.client.postStreamAt(ctx, base, endpointChatCompletions, body)
		if err != nil {
			yield(nil, err)
			return
		}
		defer stream.Close()

		err = readSSE(stream, func(data []byte) error {
			var chunk chatStreamChunk
			if err := json.Unmarshal(data, &chunk); err != nil {
				return err
			}
			resp, err := chunk.toGenerateContent()
			if err != nil {
				return err
			}
			if resp == nil {
				return nil
			}
			if !yield(resp, nil) {
				return errStreamStopped
			}
			return nil
		})
		if err != nil && !errors.Is(err, errStreamStopped) {
			yield(nil, err)
		}
	}
}

func (p *backend) GenerateVideosFromSource(ctx context.Context, model string, source *genai.GenerateVideosSource, config *genai.GenerateVideosConfig) (*genai.GenerateVideosOperation, error) {
	return nil, xai.ErrNotSupported
}

func (p *backend) GetVideosOperation(ctx context.Context, op *genai.GenerateVideosOperation, config *genai.GetOperationConfig) (*genai.GenerateVideosOperation, error) {
	return nil, xai.ErrNotSupported
}

func (p *backend) GenerateImages(ctx context.Context, model string, prompt string, config *genai.GenerateImagesConfig) (*genai.GenerateImagesResponse, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("qiniu: Prompt is required")
	}
	body := map[string]any{
		"model":  model,
		"prompt": prompt,
	}
	if config != nil {
		if config.NumberOfImages > 0 {
			body["n"] = config.NumberOfImages
		}
		if config.AspectRatio != "" {
			body["image_config"] = map[string]any{"aspect_ratio": config.AspectRatio}
		}
	}

	var resp imagesResponse
	if err := p.client.postJSONAt(ctx, p.baseURL(genImageHTTPOptions(config)), endpointImagesGenerate, body, &resp); err != nil {
		return nil, err
	}
	return resp.toGenerateImagesResponse(), nil
}

func (p *backend) EditImage(ctx context.Context, model string, prompt string, references []genai.ReferenceImage, config *genai.EditImageConfig) (*genai.EditImageResponse, error) {
	if strings.TrimSpace(prompt) == "" {
		return nil, fmt.Errorf("qiniu: Prompt is required")
	}
	images, err := collectReferenceImages(references)
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("qiniu: Image or Images is required")
	}
	body := map[string]any{
		"model":  model,
		"prompt": prompt,
	}
	if len(images) == 1 {
		body["image"] = images[0]
	} else {
		body["image"] = images
	}
	if config != nil {
		if config.NumberOfImages > 0 {
			body["n"] = config.NumberOfImages
		}
		if config.AspectRatio != "" {
			body["image_config"] = map[string]any{"aspect_ratio": config.AspectRatio}
		}
	}

	var resp imagesResponse
	if err := p.client.postJSONAt(ctx, p.baseURL(editImageHTTPOptions(config)), endpointImagesEdit, body, &resp); err != nil {
		return nil, err
	}
	return resp.toEditImageResponse(), nil
}

func (p *backend) RecontextImage(ctx context.Context, model string, source *genai.RecontextImageSource, config *genai.RecontextImageConfig) (*genai.RecontextImageResponse, error) {
	return nil, xai.ErrNotSupported
}

func (p *backend) UpscaleImage(ctx context.Context, model string, image *genai.Image, factor string, config *genai.UpscaleImageConfig) (*genai.UpscaleImageResponse, error) {
	return nil, xai.ErrNotSupported
}

func (p *backend) SegmentImage(ctx context.Context, model string, source *genai.SegmentImageSource, config *genai.SegmentImageConfig) (*genai.SegmentImageResponse, error) {
	return nil, xai.ErrNotSupported
}

func (p *backend) baseURL(opts *genai.HTTPOptions) string {
	if opts != nil && opts.BaseURL != "" {
		return normalizeBaseURL(opts.BaseURL)
	}
	return p.client.base
}

func genContentHTTPOptions(cfg *genai.GenerateContentConfig) *genai.HTTPOptions {
	if cfg == nil {
		return nil
	}
	return cfg.HTTPOptions
}

func genImageHTTPOptions(cfg *genai.GenerateImagesConfig) *genai.HTTPOptions {
	if cfg == nil {
		return nil
	}
	return cfg.HTTPOptions
}

func editImageHTTPOptions(cfg *genai.EditImageConfig) *genai.HTTPOptions {
	if cfg == nil {
		return nil
	}
	return cfg.HTTPOptions
}

func readSSE(r io.Reader, handle func([]byte) error) error {
	reader := bufio.NewReader(r)
	lines := make([]string, 0, 4)

	flush := func() error {
		if len(lines) == 0 {
			return nil
		}
		data := strings.Join(lines, "\n")
		lines = lines[:0]
		if data == "[DONE]" {
			return io.EOF
		}
		return handle([]byte(data))
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if err := flush(); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}
		} else if strings.HasPrefix(line, "data:") {
			lines = append(lines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}

	if err := flush(); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

type chatUsage struct {
	PromptTokens     int32 `json:"prompt_tokens"`
	CompletionTokens int32 `json:"completion_tokens"`
	TotalTokens      int32 `json:"total_tokens"`
}

type chatImage struct {
	Type     string `json:"type"`
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url"`
	Index int32 `json:"index"`
}

type chatToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type chatMessage struct {
	Role             string          `json:"role"`
	Content          json.RawMessage `json:"content"`
	ReasoningContent string          `json:"reasoning_content"`
	Images           []chatImage     `json:"images"`
	ToolCalls        []chatToolCall  `json:"tool_calls"`
	ToolCallID       string          `json:"tool_call_id"`
	Name             string          `json:"name"`
}

type chatCompletionChoice struct {
	Index        int32       `json:"index"`
	Message      chatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type chatCompletionResponse struct {
	Choices []chatCompletionChoice `json:"choices"`
	Usage   chatUsage              `json:"usage"`
}

func (p *chatCompletionResponse) toGenerateContent() (*genai.GenerateContentResponse, error) {
	candidates := make([]*genai.Candidate, 0, len(p.Choices))
	for _, choice := range p.Choices {
		parts, err := choice.Message.toParts()
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, &genai.Candidate{
			Index:        choice.Index,
			FinishReason: mapFinishReason(choice.FinishReason),
			Content: &genai.Content{
				Role:  mapRoleFromChat(choice.Message.Role),
				Parts: parts,
			},
		})
	}
	return &genai.GenerateContentResponse{
		Candidates:    candidates,
		UsageMetadata: p.Usage.toUsageMetadata(),
	}, nil
}

type chatStreamChoice struct {
	Index        int32       `json:"index"`
	Delta        chatMessage `json:"delta"`
	FinishReason string      `json:"finish_reason"`
}

type chatStreamChunk struct {
	Choices []chatStreamChoice `json:"choices"`
	Usage   chatUsage          `json:"usage"`
}

func (p *chatStreamChunk) toGenerateContent() (*genai.GenerateContentResponse, error) {
	if len(p.Choices) == 0 {
		if usage := p.Usage.toUsageMetadata(); usage != nil {
			return &genai.GenerateContentResponse{UsageMetadata: usage}, nil
		}
		return nil, nil
	}
	candidates := make([]*genai.Candidate, 0, len(p.Choices))
	for _, choice := range p.Choices {
		parts, err := choice.Delta.toParts()
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, &genai.Candidate{
			Index:        choice.Index,
			FinishReason: mapFinishReason(choice.FinishReason),
			Content: &genai.Content{
				Role:  mapRoleFromChat(choice.Delta.Role),
				Parts: parts,
			},
		})
	}
	return &genai.GenerateContentResponse{
		Candidates:    candidates,
		UsageMetadata: p.Usage.toUsageMetadata(),
	}, nil
}

func (p chatUsage) toUsageMetadata() *genai.GenerateContentResponseUsageMetadata {
	if p.PromptTokens == 0 && p.CompletionTokens == 0 && p.TotalTokens == 0 {
		return nil
	}
	return &genai.GenerateContentResponseUsageMetadata{
		PromptTokenCount:     p.PromptTokens,
		CandidatesTokenCount: p.CompletionTokens,
		TotalTokenCount:      p.TotalTokens,
	}
}

func (p chatMessage) toParts() ([]*genai.Part, error) {
	var parts []*genai.Part

	contentParts, err := parseContentRaw(p.Content)
	if err != nil {
		return nil, err
	}
	parts = append(parts, contentParts...)

	if p.ReasoningContent != "" {
		parts = append(parts, &genai.Part{
			Text:    p.ReasoningContent,
			Thought: true,
		})
	}
	for _, item := range p.Images {
		part, err := imagePartFromURL(item.ImageURL.URL)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}
	for _, call := range p.ToolCalls {
		args := make(map[string]any)
		if call.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				return nil, err
			}
		}
		part := genai.NewPartFromFunctionCall(call.Function.Name, args)
		part.FunctionCall.ID = call.ID
		parts = append(parts, part)
	}
	return parts, nil
}

func parseContentRaw(raw json.RawMessage) ([]*genai.Part, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	text := strings.TrimSpace(string(raw))
	if text == "" || text == "null" {
		return nil, nil
	}
	switch text[0] {
	case '"':
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			return nil, err
		}
		if s == "" {
			return nil, nil
		}
		return []*genai.Part{genai.NewPartFromText(s)}, nil
	case '[':
		var items []map[string]any
		if err := json.Unmarshal(raw, &items); err != nil {
			return nil, err
		}
		parts := make([]*genai.Part, 0, len(items))
		for _, item := range items {
			typ, _ := item["type"].(string)
			switch typ {
			case "text", "output_text", "input_text":
				if v, _ := item["text"].(string); v != "" {
					parts = append(parts, genai.NewPartFromText(v))
				}
			case "image_url":
				urlValue := ""
				if v, ok := item["image_url"].(map[string]any); ok {
					urlValue, _ = v["url"].(string)
				}
				if urlValue == "" {
					continue
				}
				part, err := imagePartFromURL(urlValue)
				if err != nil {
					return nil, err
				}
				parts = append(parts, part)
			}
		}
		return parts, nil
	default:
		var s string
		if err := json.Unmarshal(raw, &s); err == nil && s != "" {
			return []*genai.Part{genai.NewPartFromText(s)}, nil
		}
	}
	return nil, nil
}

func mapRoleFromChat(role string) string {
	switch role {
	case "assistant":
		return genai.RoleModel
	case "user":
		return genai.RoleUser
	default:
		return role
	}
}

func mapRoleToChat(role string) string {
	switch role {
	case genai.RoleModel:
		return "assistant"
	case genai.RoleUser, "":
		return "user"
	default:
		return role
	}
}

func mapFinishReason(reason string) genai.FinishReason {
	switch reason {
	case "stop":
		return genai.FinishReasonStop
	case "length":
		return genai.FinishReasonMaxTokens
	case "content_filter":
		return genai.FinishReasonSafety
	case "tool_calls":
		return genai.FinishReasonStop
	default:
		return genai.FinishReasonUnspecified
	}
}

func buildChatRequest(model string, contents []*genai.Content, config *genai.GenerateContentConfig, stream bool) (map[string]any, error) {
	messages, err := buildChatMessages(contents, config)
	if err != nil {
		return nil, err
	}
	body := map[string]any{
		"model":    model,
		"messages": messages,
		"stream":   stream,
	}
	if config == nil {
		return body, nil
	}
	if config.MaxOutputTokens > 0 {
		body["max_tokens"] = config.MaxOutputTokens
	}
	if config.Temperature != nil {
		body["temperature"] = *config.Temperature
	}
	if config.TopP != nil {
		body["top_p"] = *config.TopP
	}
	if config.TopK != nil {
		body["top_k"] = *config.TopK
	}
	if config.ImageConfig != nil && config.ImageConfig.AspectRatio != "" {
		body["image_config"] = map[string]any{
			"aspect_ratio": config.ImageConfig.AspectRatio,
		}
	}
	if tools := buildChatTools(config.Tools); len(tools) > 0 {
		body["tools"] = tools
	}
	return body, nil
}

func buildChatMessages(contents []*genai.Content, config *genai.GenerateContentConfig) ([]map[string]any, error) {
	messages := make([]map[string]any, 0, len(contents)+1)
	if config != nil && config.SystemInstruction != nil {
		if text := contentText(config.SystemInstruction); text != "" {
			messages = append(messages, map[string]any{
				"role":    "system",
				"content": text,
			})
		}
	}
	for _, content := range contents {
		items, err := buildMessage(content)
		if err != nil {
			return nil, err
		}
		messages = append(messages, items...)
	}
	return messages, nil
}

func buildMessage(content *genai.Content) ([]map[string]any, error) {
	if content == nil {
		return nil, nil
	}

	role := mapRoleToChat(content.Role)
	contentItems := make([]map[string]any, 0, len(content.Parts))
	toolCalls := make([]map[string]any, 0, 1)
	toolMessages := make([]map[string]any, 0, 1)

	for i, part := range content.Parts {
		if part == nil {
			continue
		}
		if part.FunctionResponse != nil {
			raw, err := json.Marshal(part.FunctionResponse.Response)
			if err != nil {
				return nil, err
			}
			callID := part.FunctionResponse.ID
			if callID == "" {
				callID = part.FunctionResponse.Name
			}
			if callID == "" {
				callID = "tool_" + strconv.Itoa(i)
			}
			msg := map[string]any{
				"role":         "tool",
				"tool_call_id": callID,
				"content":      string(raw),
			}
			if part.FunctionResponse.Name != "" {
				msg["name"] = part.FunctionResponse.Name
			}
			toolMessages = append(toolMessages, msg)
			continue
		}
		if part.FunctionCall != nil {
			argsRaw := "{}"
			if part.FunctionCall.Args != nil {
				raw, err := json.Marshal(part.FunctionCall.Args)
				if err != nil {
					return nil, err
				}
				argsRaw = string(raw)
			}
			callID := part.FunctionCall.ID
			if callID == "" {
				callID = "call_" + strconv.Itoa(i)
			}
			toolCalls = append(toolCalls, map[string]any{
				"id":   callID,
				"type": "function",
				"function": map[string]any{
					"name":      part.FunctionCall.Name,
					"arguments": argsRaw,
				},
			})
			continue
		}
		if part.Text != "" {
			contentItems = append(contentItems, map[string]any{
				"type": "text",
				"text": part.Text,
			})
			continue
		}
		u, ok, err := partToImageURL(part)
		if err != nil {
			return nil, err
		}
		if ok {
			contentItems = append(contentItems, map[string]any{
				"type": "image_url",
				"image_url": map[string]any{
					"url": u,
				},
			})
		}
	}

	ret := make([]map[string]any, 0, 1+len(toolMessages))
	if len(contentItems) > 0 || len(toolCalls) > 0 {
		msg := map[string]any{
			"role": role,
		}
		if len(contentItems) == 1 {
			if typ, _ := contentItems[0]["type"].(string); typ == "text" {
				msg["content"] = contentItems[0]["text"]
			} else {
				msg["content"] = contentItems
			}
		} else if len(contentItems) > 1 {
			msg["content"] = contentItems
		} else {
			msg["content"] = ""
		}
		if len(toolCalls) > 0 {
			msg["tool_calls"] = toolCalls
		}
		ret = append(ret, msg)
	}
	ret = append(ret, toolMessages...)
	return ret, nil
}

func contentText(content *genai.Content) string {
	if content == nil {
		return ""
	}
	var items []string
	for _, part := range content.Parts {
		if part != nil && part.Text != "" {
			items = append(items, part.Text)
		}
	}
	return strings.Join(items, "\n")
}

func partToImageURL(part *genai.Part) (string, bool, error) {
	if part.FileData != nil && part.FileData.FileURI != "" {
		mime := strings.TrimSpace(part.FileData.MIMEType)
		if mime != "" && !strings.HasPrefix(strings.ToLower(mime), "image/") {
			return "", false, fmt.Errorf("qiniu: unsupported file_data mime %q in chat message", mime)
		}
		return part.FileData.FileURI, true, nil
	}
	if part.InlineData != nil {
		mime := strings.TrimSpace(part.InlineData.MIMEType)
		if mime == "" {
			mime = string(xai.ImagePNG)
		}
		if !strings.HasPrefix(strings.ToLower(mime), "image/") {
			return "", false, fmt.Errorf("qiniu: unsupported inline_data mime %q in chat message", mime)
		}
		data := base64.StdEncoding.EncodeToString(part.InlineData.Data)
		return "data:" + mime + ";base64," + data, true, nil
	}
	return "", false, nil
}

func buildChatTools(tools []*genai.Tool) []map[string]any {
	items := make([]map[string]any, 0, len(tools))
	for _, t := range tools {
		if t == nil {
			continue
		}
		for _, fn := range t.FunctionDeclarations {
			if fn == nil || fn.Name == "" {
				continue
			}
			items = append(items, map[string]any{
				"type": "function",
				"function": map[string]any{
					"name":        fn.Name,
					"description": fn.Description,
					"parameters":  functionParameters(fn),
				},
			})
		}
	}
	return items
}

func functionParameters(fn *genai.FunctionDeclaration) any {
	if fn == nil {
		return map[string]any{"type": "object", "properties": map[string]any{}}
	}
	if fn.ParametersJsonSchema != nil {
		return fn.ParametersJsonSchema
	}
	if fn.Parameters != nil {
		raw, err := json.Marshal(fn.Parameters)
		if err == nil {
			var m map[string]any
			if json.Unmarshal(raw, &m) == nil {
				normalizeSchemaType(m)
				return m
			}
		}
	}
	return map[string]any{"type": "object", "properties": map[string]any{}}
}

func normalizeSchemaType(v any) {
	switch item := v.(type) {
	case map[string]any:
		for k, vv := range item {
			if k == "type" {
				if s, ok := vv.(string); ok {
					item[k] = strings.ToLower(s)
				}
			}
			normalizeSchemaType(vv)
		}
	case []any:
		for _, vv := range item {
			normalizeSchemaType(vv)
		}
	}
}

type imagesResponse struct {
	Created      int64  `json:"created"`
	OutputFormat string `json:"output_format"`
	Data         []struct {
		B64JSON string `json:"b64_json"`
		URL     string `json:"url"`
	} `json:"data"`
	Usage struct {
		TotalTokens int64 `json:"total_tokens"`
	} `json:"usage"`
}

func (p *imagesResponse) toGenerateImagesResponse() *genai.GenerateImagesResponse {
	return &genai.GenerateImagesResponse{
		GeneratedImages: p.generatedImages(),
	}
}

func (p *imagesResponse) toEditImageResponse() *genai.EditImageResponse {
	return &genai.EditImageResponse{
		GeneratedImages: p.generatedImages(),
	}
}

func (p *imagesResponse) generatedImages() []*genai.GeneratedImage {
	items := make([]*genai.GeneratedImage, 0, len(p.Data))
	for _, item := range p.Data {
		img := &genai.Image{}
		switch {
		case item.URL != "":
			img.GCSURI = item.URL
			img.MIMEType = string(guessImageMime(item.URL, p.OutputFormat))
		case item.B64JSON != "":
			format := p.OutputFormat
			if format == "" {
				format = "png"
			}
			img.GCSURI = "data:image/" + format + ";base64," + item.B64JSON
			img.MIMEType = "image/" + format
		default:
			continue
		}
		items = append(items, &genai.GeneratedImage{Image: img})
	}
	return items
}

func collectReferenceImages(refs []genai.ReferenceImage) ([]string, error) {
	images := make([]string, 0, len(refs))
	for _, ref := range refs {
		img := referenceImageOf(ref)
		if img == nil {
			continue
		}
		u, err := imageInput(img)
		if err != nil {
			return nil, err
		}
		images = append(images, u)
	}
	return images, nil
}

func referenceImageOf(ref genai.ReferenceImage) *genai.Image {
	switch v := ref.(type) {
	case *genai.RawReferenceImage:
		return v.ReferenceImage
	case *genai.MaskReferenceImage:
		return v.ReferenceImage
	case *genai.ControlReferenceImage:
		return v.ReferenceImage
	case *genai.StyleReferenceImage:
		return v.ReferenceImage
	case *genai.SubjectReferenceImage:
		return v.ReferenceImage
	case *genai.ContentReferenceImage:
		return v.ReferenceImage
	default:
		return nil
	}
}

func imageInput(img *genai.Image) (string, error) {
	if img == nil {
		return "", nil
	}
	if img.GCSURI != "" {
		return img.GCSURI, nil
	}
	if len(img.ImageBytes) > 0 {
		mime := strings.TrimSpace(img.MIMEType)
		if mime == "" {
			mime = string(xai.ImagePNG)
		}
		return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(img.ImageBytes), nil
	}
	return "", fmt.Errorf("qiniu: invalid reference image")
}

func imagePartFromURL(rawURL string) (*genai.Part, error) {
	mime, data, ok := parseDataURL(rawURL)
	if ok {
		return genai.NewPartFromBytes(data, mime), nil
	}
	return genai.NewPartFromURI(rawURL, string(guessImageMime(rawURL, ""))), nil
}

func parseDataURL(raw string) (mime string, data []byte, ok bool) {
	if !strings.HasPrefix(raw, "data:") {
		return "", nil, false
	}
	payload := strings.TrimPrefix(raw, "data:")
	pos := strings.Index(payload, ",")
	if pos < 0 {
		return "", nil, false
	}
	header := payload[:pos]
	b64 := payload[pos+1:]
	parts := strings.Split(header, ";")
	if len(parts) == 0 {
		return "", nil, false
	}
	mime = parts[0]
	if mime == "" {
		return "", nil, false
	}
	if len(parts) < 2 || parts[1] != "base64" {
		return "", nil, false
	}
	buf, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", nil, false
	}
	return mime, buf, true
}

func guessImageMime(rawURL, format string) xai.ImageType {
	if strings.HasPrefix(rawURL, "data:image/") {
		if idx := strings.Index(rawURL, ";"); idx > len("data:") {
			return xai.ImageType(rawURL[len("data:"):idx])
		}
	}
	if format != "" {
		switch strings.ToLower(strings.TrimSpace(format)) {
		case "jpeg", "jpg":
			return xai.ImageJPEG
		case "gif":
			return xai.ImageGIF
		case "webp":
			return xai.ImageWebP
		case "png":
			return xai.ImagePNG
		}
	}
	if parsed, err := url.Parse(rawURL); err == nil {
		ext := strings.ToLower(path.Ext(parsed.Path))
		switch ext {
		case ".jpg", ".jpeg":
			return xai.ImageJPEG
		case ".gif":
			return xai.ImageGIF
		case ".webp":
			return xai.ImageWebP
		case ".png":
			return xai.ImagePNG
		}
	}
	return xai.ImagePNG
}
