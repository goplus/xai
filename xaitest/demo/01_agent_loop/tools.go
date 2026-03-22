package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	pathpkg "path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	xai "github.com/goplus/xai/spec"
)

type LocalTool interface {
	Name() string
	Description() string
	InputSchema() string
	Execute(ctx context.Context, input json.RawMessage) (string, error)
}

func buildTools(root, outputDir string) []LocalTool {
	return []LocalTool{
		&ListDirTool{root: root},
		&ReadFileTool{root: root},
		&SearchCodeTool{root: root},
		&WriteFileTool{root: root, outputDir: outputDir},
	}
}

// toolWithInputSchema is an optional capability that provider tools may support.
// When present, the JSON schema is set as a structured parameter definition
// instead of being embedded in the description text.
type toolWithInputSchema interface {
	InputSchema(string) xai.Tool
}

func registerTools(provider xai.Service, root string, tools []LocalTool) {
	for _, tool := range tools {
		t := provider.ToolDef(tool.Name())
		t.Description(tool.Description())
		if ts, ok := t.(toolWithInputSchema); ok {
			ts.InputSchema(tool.InputSchema())
		}
	}
}

func buildToolIndex(tools []LocalTool) map[string]LocalTool {
	index := make(map[string]LocalTool, len(tools))
	for _, tool := range tools {
		index[tool.Name()] = tool
	}
	return index
}

// isBinary reports whether the file at path appears to be a binary file.
// It reads up to the first 512 bytes and checks for NUL bytes.
func isBinary(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	for _, b := range buf[:n] {
		if b == 0 {
			return true
		}
	}
	return false
}

func resolvePath(root, rel string) (string, error) {
	trimmed := strings.TrimSpace(rel)
	if trimmed == "" {
		trimmed = "."
	}
	var full string
	if filepath.IsAbs(trimmed) {
		full = filepath.Clean(trimmed)
	} else {
		full = filepath.Clean(filepath.Join(root, trimmed))
	}

	// Resolve symlinks to prevent traversal via symbolic links.
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve root symlinks: %w", err)
	}
	realFull, err := filepath.EvalSymlinks(full)
	if err != nil {
		// Path may not exist yet (e.g. for write targets).
		// Resolve the nearest existing parent to prevent traversal via
		// an ancestor symlink that points outside root.
		parentReal, pErr := filepath.EvalSymlinks(filepath.Dir(full))
		if pErr != nil {
			realFull = full // parent doesn't exist either; caller will get a fs error
		} else {
			realFull = filepath.Join(parentReal, filepath.Base(full))
		}
	}

	relToRoot, err := filepath.Rel(realRoot, realFull)
	if err != nil {
		return "", err
	}
	if relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root: %s", rel)
	}
	return realFull, nil
}

type ListDirTool struct {
	root string
}

func (t *ListDirTool) Name() string { return "list_dir" }

func (t *ListDirTool) Description() string {
	return "List direct children in a directory. Non-recursive and hidden entries are omitted."
}

func (t *ListDirTool) InputSchema() string {
	return `{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`
}

func (t *ListDirTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	full, err := resolvePath(t.root, req.Path)
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(full)
	if err != nil {
		return "", err
	}

	items := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if entry.IsDir() {
			items = append(items, fmt.Sprintf("[dir] %s", name))
			continue
		}
		info, err := entry.Info()
		if err != nil {
			items = append(items, fmt.Sprintf("[file] %s", name))
			continue
		}
		items = append(items, fmt.Sprintf("[file] %s (%d bytes)", name, info.Size()))
	}
	sort.Strings(items)
	if len(items) == 0 {
		return "(empty)", nil
	}
	return strings.Join(items, "\n"), nil
}

type ReadFileTool struct {
	root string
}

func (t *ReadFileTool) Name() string { return "read_file" }

func (t *ReadFileTool) Description() string {
	return "Read a file with line numbers. Supports pagination using offset and limit in line units."
}

func (t *ReadFileTool) InputSchema() string {
	return `{"type":"object","properties":{"path":{"type":"string"},"offset":{"type":"integer","minimum":0},"limit":{"type":"integer","minimum":1,"maximum":300}},"required":["path"]}`
}

func (t *ReadFileTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var req struct {
		Path   string `json:"path"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if strings.TrimSpace(req.Path) == "" {
		return "", fmt.Errorf("path is required")
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	if req.Limit <= 0 {
		req.Limit = 300
	}
	if req.Limit > 300 {
		req.Limit = 300
	}

	full, err := resolvePath(t.root, req.Path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(full)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory, not a file; use list_dir instead", req.Path)
	}
	if isBinary(full) {
		return "", fmt.Errorf("binary file, cannot read as text: %s", req.Path)
	}
	file, err := os.Open(full)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 1024*64)
	scanner.Buffer(buf, 1024*1024)

	start := req.Offset
	end := req.Offset + req.Limit
	lineNo := 0
	var out strings.Builder
	for scanner.Scan() {
		if lineNo >= start && lineNo < end {
			out.WriteString(fmt.Sprintf("%6d\t%s\n", lineNo+1, scanner.Text()))
		}
		lineNo++
		if lineNo >= end {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	if out.Len() == 0 {
		return "(no lines in requested range)", nil
	}
	if lineNo >= end {
		out.WriteString(fmt.Sprintf("\n...more lines available, use offset=%d\n", end))
	}
	return strings.TrimRight(out.String(), "\n"), nil
}

type SearchCodeTool struct {
	root string
}

func (t *SearchCodeTool) Name() string { return "search_code" }

func (t *SearchCodeTool) Description() string {
	return "Search text patterns in files under a path with optional glob filter and context lines."
}

func (t *SearchCodeTool) InputSchema() string {
	return `{"type":"object","properties":{"pattern":{"type":"string"},"path":{"type":"string"},"glob":{"type":"string"},"context":{"type":"integer","minimum":0,"maximum":10}},"required":["pattern"]}`
}

func (t *SearchCodeTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var req struct {
		Pattern string `json:"pattern"`
		Path    string `json:"path"`
		Glob    string `json:"glob"`
		Context int    `json:"context"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if strings.TrimSpace(req.Pattern) == "" {
		return "", errors.New("pattern is required")
	}
	if req.Context < 0 {
		req.Context = 0
	}
	if req.Context > 10 {
		req.Context = 10
	}
	if strings.TrimSpace(req.Glob) == "" {
		req.Glob = "*"
	}

	searchRoot, err := resolvePath(t.root, req.Path)
	if err != nil {
		return "", err
	}

	re, err := regexp.Compile(req.Pattern)
	if err != nil {
		return "", fmt.Errorf("invalid pattern: %w", err)
	}
	matchLine := func(line string) bool {
		return re.MatchString(line)
	}

	var results []string
	err = filepath.WalkDir(searchRoot, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") {
				if filePath == searchRoot {
					return nil
				}
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(entry.Name(), ".") {
			return nil
		}
		rel, relErr := filepath.Rel(searchRoot, filePath)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)

		matched, err := pathpkg.Match(req.Glob, rel)
		if err != nil || !matched {
			var nameMatched bool
			nameMatched, err = pathpkg.Match(req.Glob, entry.Name())
			matched = nameMatched
		}
		if err != nil {
			return err
		}
		if !matched {
			return nil
		}

		// Skip binary files.
		if isBinary(filePath) {
			return nil
		}
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if !matchLine(line) {
				continue
			}
			relToRepo, _ := filepath.Rel(t.root, filePath)
			relToRepo = filepath.ToSlash(relToRepo)

			start := i - req.Context
			if start < 0 {
				start = 0
			}
			end := i + req.Context + 1
			if end > len(lines) {
				end = len(lines)
			}
			for j := start; j < end; j++ {
				var marker string
				switch {
				case j < i:
					marker = fmt.Sprintf("  %s:%d- ", relToRepo, j+1)
				case j == i:
					marker = fmt.Sprintf("%s:%d: ", relToRepo, j+1)
				default:
					marker = fmt.Sprintf("  %s:%d+ ", relToRepo, j+1)
				}
				results = append(results, marker+strings.TrimSpace(lines[j]))
			}
			if len(results) >= 50 {
				return errors.New("max-results-reached")
			}
		}
		return nil
	})
	if err != nil && err.Error() != "max-results-reached" {
		return "", err
	}
	if len(results) == 0 {
		return "(no matches)", nil
	}
	if len(results) > 50 {
		results = results[:50]
	}
	return strings.Join(results, "\n"), nil
}

type WriteFileTool struct {
	root      string
	outputDir string // absolute path; writes here are allowed even if outside root
}

func (t *WriteFileTool) Name() string { return "write_file" }

func (t *WriteFileTool) Description() string {
	return "Write full content to a file in the output directory. Creates parent directories if needed."
}

func (t *WriteFileTool) InputSchema() string {
	return `{"type":"object","properties":{"path":{"type":"string"},"content":{"type":"string"}},"required":["path","content"]}`
}

func (t *WriteFileTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(input, &req); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	p := strings.TrimSpace(req.Path)
	absOut := t.outputDir // already absolute (set in buildTools)

	var full string
	if filepath.IsAbs(p) {
		full = filepath.Clean(p)
	} else {
		// Relative paths resolve against outputDir, not root.
		full = filepath.Clean(filepath.Join(absOut, p))
	}

	// Resolve symlinks to prevent traversal via symbolic links.
	realOut, err := filepath.EvalSymlinks(absOut)
	if err != nil {
		return "", fmt.Errorf("resolve output dir symlinks: %w", err)
	}
	realFull, err := filepath.EvalSymlinks(full)
	if err != nil {
		// Target may not exist yet; resolve parent and append filename.
		parentReal, pErr := filepath.EvalSymlinks(filepath.Dir(full))
		if pErr != nil {
			realFull = full // parent doesn't exist either; will fail at MkdirAll
		} else {
			realFull = filepath.Join(parentReal, filepath.Base(full))
		}
	}

	// Policy: writes must be inside outputDir only.
	if realFull != realOut && !strings.HasPrefix(realFull, realOut+string(filepath.Separator)) {
		return "", fmt.Errorf("write_file to %q is blocked: must be inside output dir %q", req.Path, absOut)
	}

	if err := os.MkdirAll(filepath.Dir(realFull), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(realFull, []byte(req.Content), 0o644); err != nil {
		return "", err
	}

	rel, err := filepath.Rel(absOut, realFull)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Sprintf("written %d bytes to %s", len(req.Content), full), nil
	}
	return fmt.Sprintf("written %d bytes to %s", len(req.Content), filepath.ToSlash(rel)), nil
}
