package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"unicode/utf8"

	xai "github.com/goplus/xai/spec"
)

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------

// AgentConfig holds everything the agent needs. No hardcoded phases or filenames.
type AgentConfig struct {
	Service         xai.Service
	Model           string
	Task            string // user instruction: what to do, what files to produce, content requirements
	TargetDir       string // absolute path to the codebase to examine
	OutputDir       string // absolute path where deliverables are written
	MaxTurns        int    // upper bound on LLM round-trips (default 30)
	MaxOutputTokens int64  // max tokens per LLM response (default 8192)
	Tools           []LocalTool
}

// ---------------------------------------------------------------------------
// System Prompt — role and working method only, no task specifics
// ---------------------------------------------------------------------------

const systemPromptTemplate = `You are a senior software engineer. You will receive a task from the user.

## Repository

Root: {{TARGET_DIR}}

## Working Method

- Use ReAct: explain your reasoning before making tool calls.
- Start by exploring the directory structure, then read key files to build understanding.
- Use search_code to find patterns, inconsistencies, or specific constructs.
- Produce well-structured Markdown documents with clear headings.
- Be specific: cite file paths and line numbers when discussing findings.
- Focus on substance — not cosmetic nitpicks.

## Constraints

- You may ONLY write files into the output directory: {{OUTPUT_DIR}}
- Do NOT modify any source code in the repository.
- When you have completed all requested deliverables, stop calling tools.
`

func buildSystemPrompt(targetDir, outputDir string) string {
	r := strings.NewReplacer(
		"{{TARGET_DIR}}", targetDir,
		"{{OUTPUT_DIR}}", outputDir,
	)
	return r.Replace(systemPromptTemplate)
}

// ---------------------------------------------------------------------------
// Agent Loop — no phases, LLM drives its own workflow
// ---------------------------------------------------------------------------

func runAgent(ctx context.Context, cfg AgentConfig) error {
	if cfg.MaxTurns <= 0 {
		cfg.MaxTurns = 30
	}
	if cfg.MaxOutputTokens <= 0 {
		cfg.MaxOutputTokens = 8192
	}

	toolRefs := make([]xai.ToolBase, 0, len(cfg.Tools))
	for _, t := range cfg.Tools {
		toolRefs = append(toolRefs, cfg.Service.Tool(t.Name()))
	}
	toolIndex := buildToolIndex(cfg.Tools)

	// Always use the absolute output path in the system prompt.
	// Computing a relative path from TargetDir to OutputDir is unreliable when
	// they live on completely different directory trees (e.g. separate repos).
	absOutput, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		absOutput = cfg.OutputDir
	}

	sysPrompt := buildSystemPrompt(cfg.TargetDir, absOutput)

	// The Task is the complete user instruction — filenames, content structure, everything.
	history := []xai.MsgBuilder{cfg.Service.UserMsg().Text(cfg.Task)}

	for turn := 0; turn < cfg.MaxTurns; turn++ {
		fmt.Printf("\n-- Turn %d/%d --\n", turn+1, cfg.MaxTurns)

		resp, err := cfg.Service.Gen(ctx,
			cfg.Service.Params().
				Model(xai.Model(cfg.Model)).
				MaxOutputTokens(cfg.MaxOutputTokens).
				System(cfg.Service.Texts(sysPrompt)).
				Messages(history...).
				Tools(toolRefs...),
			cfg.Service.Options(),
		)
		if err != nil {
			return fmt.Errorf("gen: %w", err)
		}
		if resp.Len() == 0 {
			return fmt.Errorf("gen: empty response")
		}

		printTextParts(resp)
		history = append(history, resp.At(0).ToMsg())

		toolCalls := extractAllToolUses(resp)

		// No tool calls → LLM thinks it's done; check for actual output.
		if len(toolCalls) == 0 {
			has, err := hasOutputFiles(absOutput)
			if err != nil {
				fmt.Printf("[Agent] Warning: could not check output dir: %v\n", err)
			}
			if has {
				fmt.Println("\n[Agent] Deliverables produced. Finishing.")
				break
			}
			// No output yet — nudge the agent to keep working.
			history = append(history, cfg.Service.UserMsg().Text(
				"You haven't produced any deliverables yet. Please continue working.",
			))
			continue
		}

		// Execute tool calls concurrently.
		type result struct {
			id, name string
			out      string
			isError  bool
		}
		results := make([]result, len(toolCalls))
		var wg sync.WaitGroup
		for i, call := range toolCalls {
			wg.Add(1)
			go func(idx int, c xai.ToolUse) {
				defer wg.Done()
				fmt.Printf("\n[Tool] %s(%s)\n", c.Name, shortJSON(c.Input))

				out, err := dispatchTool(ctx, toolIndex, c)
				if err != nil {
					fmt.Printf("[Result] ERROR: %s\n", err)
					results[idx] = result{id: c.ID, name: c.Name, out: err.Error(), isError: true}
					return
				}
				fmt.Printf("[Result] %s\n", truncate(out, 300))
				results[idx] = result{id: c.ID, name: c.Name, out: out}
			}(i, call)
		}
		wg.Wait()

		trMsg := cfg.Service.UserMsg()
		for _, r := range results {
			val := any(r.out)
			if r.isError {
				val = errors.New(r.out)
			}
			trMsg = trMsg.ToolResult(xai.ToolResult{
				ID: r.id, Name: r.name, Result: val, IsError: r.isError,
			})
		}
		history = append(history, trMsg)
	}

	return reportResults(absOutput)
}

// ---------------------------------------------------------------------------
// Completion check — no hardcoded filenames
// ---------------------------------------------------------------------------

// hasOutputFiles reports whether any non-empty, non-directory file exists
// anywhere under outputDir (including subdirectories).
func hasOutputFiles(outputDir string) (bool, error) {
	found := false
	err := filepath.WalkDir(outputDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil // skip unreadable files
		}
		if info.Size() > 0 {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found, err
}

// ---------------------------------------------------------------------------
// Final report — list absolute paths of all produced files
// ---------------------------------------------------------------------------

func reportResults(outputDir string) error {
	absDir, err := filepath.Abs(outputDir)
	if err != nil {
		return fmt.Errorf("resolve output dir: %w", err)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  Task completed! Deliverables:")
	fmt.Println(strings.Repeat("=", 60))

	found := 0
	err = filepath.WalkDir(absDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() > 0 {
			fmt.Printf("  ✓ %s  (%d bytes)\n", path, info.Size())
			found++
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk output dir: %w", err)
	}

	if found == 0 {
		fmt.Println("  (no files produced)")
		return fmt.Errorf("no deliverables were produced")
	}
	fmt.Println(strings.Repeat("=", 60))
	return nil
}

func extractAllToolUses(resp xai.GenResponse) []xai.ToolUse {
	if resp.Len() == 0 {
		return nil
	}
	candidate := resp.At(0)
	var uses []xai.ToolUse
	for i := 0; i < candidate.Parts(); i++ {
		if toolUse, ok := candidate.Part(i).AsToolUse(); ok {
			uses = append(uses, toolUse)
		}
	}
	return uses
}

func printTextParts(resp xai.GenResponse) {
	if resp.Len() == 0 {
		return
	}
	candidate := resp.At(0)
	for i := 0; i < candidate.Parts(); i++ {
		text := strings.TrimSpace(candidate.Part(i).Text())
		if text == "" {
			continue
		}
		fmt.Printf("\n[Agent] %s\n", text)
	}
}

func dispatchTool(ctx context.Context, index map[string]LocalTool, toolUse xai.ToolUse) (string, error) {
	tool, ok := index[toolUse.Name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", toolUse.Name)
	}
	input, err := json.Marshal(toolUse.Input)
	if err != nil {
		return "", fmt.Errorf("marshal tool input: %w", err)
	}
	return tool.Execute(ctx, input)
}

func shortJSON(value any) string {
	b, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return truncate(string(b), 120)
}

func truncate(text string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(text) <= max {
		return text
	}
	// Walk back to a valid UTF-8 boundary to avoid splitting multi-byte runes.
	i := max
	for i > 0 && !utf8.RuneStart(text[i]) {
		i--
	}
	return text[:i] + "..."
}
