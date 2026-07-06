package provider

import (
	"context"

	"mhkyle/my-harness/internal/schema"
)

type MockProvider struct {
	turn int
}

func NewMockProvider() LLMProvider {
	return &MockProvider{}
}

func (m *MockProvider) Generate(ctx context.Context, messages []schema.Message, availableTools []schema.ToolDefinition) (*schema.Message, error) {
	m.turn++

	if m.turn == 1 {
		return &schema.Message{Role: schema.RoleAssistant, Content: "I want to check the current directory for files.", ToolCalls: []schema.ToolCall{{ID: "call_123", Name: "bash", Arguments: []byte(`{"command": "ls -lah"}`)}}}, nil
	}

	return &schema.Message{Role: schema.RoleAssistant, Content: "I have seen the file list, which includes main.go. Task completed!"}, nil
}
