package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"

	_ "github.com/goplus/xai/claude"
	_ "github.com/goplus/xai/openai"
	xai "github.com/goplus/xai/spec"
)

const (
	defaultURI   = "claude:"
	defaultModel = "claude-sonnet-4-6"
)

// defaultTask is the built-in instruction used when -task is not set.
// Task specifics (filenames, content structure) belong here, NOT in the system prompt.
const defaultTask = `Review the codebase and produce two deliverables:

1. **_analysis.md** — A comprehensive analysis:
   - Architecture overview and module structure
   - Code quality issues (bugs, anti-patterns, dead code, inconsistencies)
   - Test coverage gaps and API design concerns
   - Performance or security considerations
   - Prioritized improvement candidates with impact assessment

2. **_plan.md** — A concrete improvement plan:
   - Chosen targets with rationale
   - Detailed change list by file
   - Implementation approach and design decisions
   - Validation steps, risk assessment, and rollback strategy
`

func main() {
	uri := flag.String("uri", defaultURI, "provider URI (e.g. claude:, openai:)")
	model := flag.String("model", defaultModel, "model name")
	task := flag.String("task", defaultTask, "user instruction: what to do and what files to produce")
	targetDir := flag.String("target", ".", "directory to examine (default: current directory)")
	outputDir := flag.String("output", "./output", "directory to write deliverables (default: ./output)")
	maxTurns := flag.Int("maxturns", 30, "maximum number of LLM round-trips")
	flag.Parse()

	absTarget, err := filepath.Abs(*targetDir)
	if err != nil {
		log.Fatalf("resolve target dir: %v", err)
	}

	absOutput, err := filepath.Abs(*outputDir)
	if err != nil {
		log.Fatalf("resolve output dir: %v", err)
	}
	if err := os.MkdirAll(absOutput, 0o755); err != nil {
		log.Fatalf("create output dir: %v", err)
	}

	ctx := context.Background()
	provider, err := xai.New(ctx, *uri)
	if err != nil {
		log.Fatalf("xai.New: %v", err)
	}

	tools := buildTools(absTarget, absOutput)
	registerTools(provider, absTarget, tools)

	if err := runAgent(ctx, AgentConfig{
		Provider:  provider,
		Model:     *model,
		Task:      *task,
		TargetDir: absTarget,
		OutputDir: absOutput,
		MaxTurns:  *maxTurns,
		Tools:     tools,
	}); err != nil {
		log.Fatal(err)
	}
}
