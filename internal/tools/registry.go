package tools

import (
	"context"

	"mhkyle/my-harness/internal/schema"
)

type Registry interface {
	GetAvailableTools() []schema.ToolDefinition

	Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult
}
