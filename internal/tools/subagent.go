package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mhkyle/my-harness/internal/schema"
)

type AgentRunner interface {
	RunSub(ctx context.Context, taskPrompt string, readOnlyRegistry Registry, reporter interface{}) (string, error)
}

type subagentArgs struct {
	TaskPrompt string `json:"task_prompt"`
}

type SubagentTool struct {
	runner           AgentRunner
	readonlyRegistry Registry
	reporter         interface{}
}

func NewSubagentTool(runner AgentRunner, readOnlyRegistry Registry, reporter interface{}) BaseTool {
	return &SubagentTool{
		runner:           runner,
		readonlyRegistry: readOnlyRegistry,
		reporter:         reporter,
	}
}

func (s *SubagentTool) Name() string {
	return "subagent"
}

func (s *SubagentTool) Definition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        s.Name(),
		Description: "Run a sub-agent with the given task prompt, when you need to delegate a task to another agent. It will run the sub-agent and return the final result.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"task_prompt": map[string]interface{}{
					"type":        "string",
					"description": "The task prompt for the sub-agent to execute.",
				},
			},
			"required": []string{"task_prompt"},
		},
	}
}

func (s *SubagentTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var input subagentArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("failed to unmarshal input args: %w", err)
	}

	log.Printf("[SubAgent] Executing sub-agent with task prompt: %s", input.TaskPrompt)
	summary, err := s.runner.RunSub(ctx, input.TaskPrompt, s.readonlyRegistry, s.reporter)
	if err != nil {
		return "", fmt.Errorf("failed to run sub-agent: %w", err)
	}

	log.Printf("[SubAgent] Sub-agent execution completed. Summary: %s", summary)

	return fmt.Sprintf("Sub-agent execution completed. Summary: %s", summary), nil
}
