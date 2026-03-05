# 01_agent_loop

An LLM-driven agent that autonomously explores a Go codebase using tool calls and produces analysis documents.

## How It Works

The agent uses a single continuous loop — the LLM decides what to do based on the user's task instruction. Available tools: `list_dir`, `read_file`, `search_code`, `write_file`. The system prompt defines role and method; the task defines deliverables.

## Usage

```bash
# Analyze the current directory, output to ./output
go run .

# Analyze a specific codebase, output to a separate directory
go run . -target /path/to/repo -output /tmp/results

# Use a custom provider endpoint
go run . -uri "claude:base=https://api.example.com"
```

### Flags

| Flag        | Default             | Description                       |
| ----------- | ------------------- | --------------------------------- |
| `-uri`      | `claude:`           | Provider URI                      |
| `-model`    | `claude-sonnet-4-6` | Model name                        |
| `-task`     | _(built-in)_        | What to analyze and produce       |
| `-target`   | `.`                 | Directory to examine              |
| `-output`   | `./output`          | Directory for deliverables        |
| `-maxturns` | `30`                | Maximum number of LLM round-trips |

## Default Task

By default the agent produces two Markdown files: `_analysis.md` (code review) and `_plan.md` (improvement plan). Override with `-task` for custom deliverables.
