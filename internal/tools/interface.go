package tools

import (
	"context"
	"encoding/json"

	"mhkyle/my-harness/internal/schema"
)

type BaseTool interface {
	Name() string
	Definition() schema.ToolDefinition
	Execute(ctx context.Context, args json.RawMessage) (string, error)
}

type Registry interface {
	Register(tool BaseTool)

	GetAvailableTools() []schema.ToolDefinition

	Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult
}
