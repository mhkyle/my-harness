package tools

import (
	"context"
	"encoding/json"

	"mhkyle/my-harness/internal/schema"
)

type MiddlewareFunc func(ctx context.Context, call schema.ToolCall) (allowed bool, rejectReason string)

type BaseTool interface {
	Name() string
	Definition() schema.ToolDefinition
	Execute(ctx context.Context, args json.RawMessage) (string, error)
}

type Registry interface {
	Register(tool BaseTool)

	GetAvailableTools() []schema.ToolDefinition

	Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult

	Use(mw MiddlewareFunc) // middleware for tool execution, can be used for logging, authentication, etc.
}
