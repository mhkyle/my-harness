package provider

import (
	"context"

	"mhkyle/my-harness/internal/schema"
)

type LLMProvider interface {
	Generate(ctx context.Context, messages []schema.Message, availableTools []schema.ToolDefinition) (*schema.Message, error)
}
