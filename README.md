# my-harness

A small Go-based agent harness for tool-calling experiments.

## What it does

- Runs a simple agent loop (`internal/engine`) with message history.
- Uses an `LLMProvider` interface (`internal/provider`) to generate assistant messages and tool calls.
- Executes tools through a registry interface (`internal/tools`).
- Includes a working `bash` tool (`internal/tools/bash`) that can run shell commands like:
  - `pwd`
  - `ls -alh`
  - `echo hello`
- Ships with a mock provider so the project can run end-to-end without external APIs.

## Project structure

```text
.
├── cmd/my-harness/main.go          # Entrypoint
├── internal/engine/loop.go         # Agent loop
├── internal/provider/
│   ├── interface.go                # LLMProvider interface
│   └── mock_provider.go            # Demo provider implementation
├── internal/schema/message.go      # Message/tool schema types
└── internal/tools/
    ├── registry.go                 # Tool registry interface
    └── bash/
        ├── bash_tool.go            # Bash tool implementation
        └── bash_tool_test.go       # Bash tool tests
```

## Quick start

### Prerequisites

- Go `1.26.4` (see `go.mod`)
- `bash` available in your environment

### Run

```bash
go run ./cmd/my-harness
```

Expected behavior:

1. Engine starts with a user prompt.
2. Mock provider asks to call `bash` with `ls -lah`.
3. Bash tool executes the command and returns output.
4. Mock provider returns a final completion message.

## Bash tool contract

Tool name: `bash`

Input JSON:

```json
{
  "command": "ls -alh",
  "workdir": "/optional/path"
}
```

- `command` is required.
- `workdir` is optional; defaults to current process working directory.
- Output is combined `stdout` + `stderr`.
- On command failure, the tool returns `is_error: true` and command output.

## Test and lint

```bash
go test ./...
golangci-lint run
```

## Notes

- The current provider is a mock implementation for local experimentation.
- The current `bash` tool executes the provided command directly; if you need stricter safety, add command validation/allowlisting in `internal/tools/bash/bash_tool.go`.
