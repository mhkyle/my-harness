package tools

import (
	"context"

	"mhkyle/my-harness/internal/schema"
)

type mockRegistry struct {
}

func NewMockRegistry() Registry {
	return &mockRegistry{}
}

func (m *mockRegistry) GetAvailableTools() []schema.ToolDefinition {
	return []schema.ToolDefinition{
		{
			Name: "bash",
		}}
}

func (m *mockRegistry) Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult {
	return schema.ToolResult{
		ToolCallID: call.ID,
		Output:     "-rw-r--r-- 1 user group 234 Oct 24 10:00 main.go\n",
		IsError:    false,
	}
}
