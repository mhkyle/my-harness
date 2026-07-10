package tools

import (
	"context"
	"fmt"
	"log"

	"mhkyle/my-harness/internal/schema"
)

type RegisterImpl struct {
	tools map[string]BaseTool
}

func NewRegistry() Registry {
	return &RegisterImpl{
		tools: make(map[string]BaseTool),
	}
}

func (r *RegisterImpl) Register(tool BaseTool) {
	name := tool.Name()
	if _, exists := r.tools[name]; exists {
		log.Printf("tool already registered by %s, will replace it", name)
	}
	r.tools[name] = tool
	log.Printf("tool registered: %s", name)
}

func (r *RegisterImpl) GetAvailableTools() []schema.ToolDefinition {
	var all []schema.ToolDefinition
	for _, tool := range r.tools {
		all = append(all, tool.Definition())
	}
	return all
}

func (r *RegisterImpl) Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult {
	tool, exists := r.tools[call.Name]
	if !exists {
		return schema.ToolResult{
			ToolCallID: call.ID,
			Output:     fmt.Sprintf("tool not found: %s", call.Name),
			IsError:    true,
		}
	}

	output, err := tool.Execute(ctx, call.Arguments)
	if err != nil {
		log.Printf("tool execution error: %v", err)
		return schema.ToolResult{
			ToolCallID: call.ID,
			Output:     fmt.Sprintf("tool execution error: %v", err),
			IsError:    true,
		}
	}

	return schema.ToolResult{
		ToolCallID: call.ID,
		Output:     output,
		IsError:    false,
	}
}
