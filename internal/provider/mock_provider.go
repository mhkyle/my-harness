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
	if len(availableTools) == 0 {
		return &schema.Message{Role: schema.RoleAssistant, Content: "I have no tools available to use. Need to use bash tool to check the current directory with `ls` command."}, nil
	}
	m.turn++

	if m.turn == 1 {
		return &schema.Message{Role: schema.RoleAssistant,
			Content: "Need to execute the plans that generated last turn.", ToolCalls: []schema.ToolCall{{ID: "call_123", Name: "bash", Arguments: []byte(`{"command": "ls -lah"}`)}}}, nil
	}

	return &schema.Message{Role: schema.RoleAssistant, Content: "I have seen the file list, which includes main.go. Task completed!"}, nil
}
