# my-harness

`my-harness` is a small Go agent harness for experimenting with LLM tool calling. It keeps the core loop, provider adapters, message schema, and tools separated so new models and tools can be added with minimal wiring.

## What It Does

- Runs an agent loop that maintains message history and repeatedly asks a model for the next assistant response.
- Exposes available tools to the model through a registry.
- Executes model-requested tool calls and feeds observations back into the conversation.
- Supports concurrent execution when the model returns multiple tool calls in one turn.
- Supports an optional two-stage ReAct-style thinking pass before the tool-enabled response.
- Includes provider adapters for OpenAI-compatible chat-completion APIs and Anthropic Claude-compatible message APIs.
- Includes a mock provider for local experiments that do not need an external model.
- Includes a `read_file` tool that lets the model inspect files under the current workspace.

## Current Flow

The current entrypoint in `cmd/my-harness/main.go` does the following:

1. Uses the current process directory as the workspace root.
2. Reads an API key from a local file.
3. Creates an OpenAI-compatible provider with a configured base URL and model name.
4. Creates a tool registry.
5. Registers the `read_file` tool against the workspace root.
6. Runs the agent with the prompt: `Please tell me the project in current directory is mainly for.`

The current `main.go` is configured for a local environment. Update the API key source, base URL, model, prompt, and registered tools before using it in a different environment.

## Project Structure

```text
.
├── cmd/my-harness/main.go          # Local runnable entrypoint
├── go.mod                          # Module and dependency definition
├── internal/engine/loop.go         # Agent loop and tool-call orchestration
├── internal/provider/
│   ├── interface.go                # LLMProvider interface
│   ├── openai.go                   # OpenAI-compatible chat-completion provider
│   ├── claude.go                   # Claude-compatible messages provider
│   └── mock_provider.go            # Mock provider for local experiments
├── internal/schema/message.go      # Provider-neutral message and tool schema
└── internal/tools/
    ├── interface.go                # BaseTool and Registry interfaces
    ├── registry.go                 # Tool registry implementation
    ├── read_file.go                # Workspace-scoped file reader tool
    └── bash/bash_tool.go           # Experimental bash tool implementation
```

The repository also includes a `vendor/` directory with vendored dependencies for the OpenAI and Anthropic Go SDKs.

## Core Concepts

### Agent Engine

`internal/engine.AgentEngine` owns the run loop. Each turn:

1. Sends conversation history to the configured `LLMProvider`.
2. Passes registered tool definitions when tool use is enabled for that call.
3. Appends the assistant response to history.
4. Executes any returned tool calls through the registry.
5. Appends tool observations as user messages tied to the original tool call IDs.
6. Stops when the assistant response contains no tool calls.

When `EnableThinking` is `true`, the engine first calls the provider without tools so the model can produce a planning/thinking response, then calls again with tools available.

### Providers

All model adapters implement:

```go
type LLMProvider interface {
  Generate(ctx context.Context, messages []schema.Message, availableTools []schema.ToolDefinition) (*schema.Message, error)
}
```

Available provider implementations:

- `OpenAIProvider`: adapts the internal schema to OpenAI-compatible chat completions and function tools.
- `ClaudeProvider`: adapts the internal schema to Anthropic-compatible messages and tool-use blocks.
- `MockProvider`: returns deterministic responses for simple local harness testing.

### Tool Registry

Tools implement `BaseTool` and are registered by name:

```go
type BaseTool interface {
  Name() string
  Definition() schema.ToolDefinition
  Execute(ctx context.Context, args json.RawMessage) (string, error)
}
```

The registry exposes tool definitions to the provider and dispatches returned tool calls by name.

## Tools

### `read_file`

Reads a file from the configured workspace root.

Input JSON:

```json
{
  "path": "README.md"
}
```

Behavior:

- `path` is required.
- The path is joined with the configured workspace directory.
- File content is capped at `8000` bytes.
- Errors are returned to the model as tool execution errors.

### `bash`

There is an experimental `bash` tool under `internal/tools/bash`. It is not currently registered by `main.go`.

Input JSON:

```json
{
  "command": "pwd",
  "workdir": "/optional/path"
}
```

Behavior:

- `command` is required.
- `workdir` is optional and defaults to the process working directory.
- Output is combined `stdout` and `stderr`.
- Commands are executed through `bash -lc`.

Security note: this tool executes arbitrary shell commands. Add validation, allowlisting, sandboxing, or human approval before enabling it in any sensitive workspace.

## Quick Start

### Prerequisites

- Go `1.26.4` or the version declared in `go.mod`.
- Network/API access for the provider configured in `cmd/my-harness/main.go`.
- A local API key source matching the entrypoint configuration.

### Install Dependencies

Dependencies are vendored, so standard Go commands can use the checked-in `vendor/` directory:

```bash
go mod vendor
```

Run this only when dependency versions change.

### Run

```bash
go run ./cmd/my-harness
```

Expected behavior with the current entrypoint:

1. The engine starts in the current directory.
2. The model receives the user prompt and the `read_file` tool definition.
3. The model may call `read_file` to inspect project files.
4. The engine prints model responses and logs tool execution progress.
5. The run exits when the model returns a response with no tool calls.

## Development

### Add a Provider

1. Create a new implementation under `internal/provider`.
2. Convert `schema.Message` values into the target provider's message format.
3. Convert `schema.ToolDefinition` values into the provider's tool/function schema.
4. Convert provider responses back into `schema.Message`, including tool call IDs, names, and JSON arguments.
5. Wire the provider in `cmd/my-harness/main.go` or a new command.

### Add a Tool

1. Implement `tools.BaseTool`.
2. Return a JSON-schema-like input definition from `Definition()`.
3. Parse `json.RawMessage` arguments in `Execute()`.
4. Register the tool with `registry.Register(...)` before calling `engine.Run(...)`.

Example registration:

```go
registry := tools.NewRegistry()
registry.Register(tools.NewReadFileTool(workDir))
```

## Test and Lint

```bash
go test ./...
golangci-lint run
```

`golangci-lint` is optional unless your local workflow requires it.

## Current Limitations

- `cmd/my-harness/main.go` is currently a local experiment entrypoint, not a configurable CLI.
- API credentials are read from a local file path in `main.go`; prefer environment variables or a config file before sharing or deploying.
- `read_file` joins paths with the workspace root but does not currently reject path traversal such as `../`.
- The experimental `bash` tool is not registered by default and should be hardened before use.
- There are no first-party tests currently checked in for the active engine/provider/tool paths.
