package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"mhkyle/my-harness/internal/schema"
	"mhkyle/my-harness/internal/tools"
)

type BashTool struct {
}

type bashInput struct {
	Command string `json:"command"`
	WorkDir string `json:"workdir,omitempty"`
}

func NewBashTool() tools.Registry {
	return &BashTool{}
}

func (m *BashTool) GetAvailableTools() []schema.ToolDefinition {
	return []schema.ToolDefinition{
		{
			Name:        "bash",
			Description: "Execute bash commands in the current directory.",
			InputSchema: string(`{"type":"object","properties":{"command":{"type":"string","description":"Bash command to execute."},"workdir":{"type":"string","description":"Optional working directory for command execution."}},"required":["command"]}`),
		},
	}
}

func (m *BashTool) Execute(ctx context.Context, call schema.ToolCall) schema.ToolResult {
	if call.Name != "bash" {
		return schema.ToolResult{
			ToolCallID: call.ID,
			Output:     "Unknown tool call",
			IsError:    true,
		}
	}

	var input bashInput
	if err := json.Unmarshal(call.Arguments, &input); err != nil {
		return schema.ToolResult{
			ToolCallID: call.ID,
			Output:     fmt.Sprintf("invalid arguments: %v", err),
			IsError:    true,
		}
	}

	if strings.TrimSpace(input.Command) == "" {
		return schema.ToolResult{
			ToolCallID: call.ID,
			Output:     "command is required",
			IsError:    true,
		}
	}

	workDir := input.WorkDir
	if workDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return schema.ToolResult{
				ToolCallID: call.ID,
				Output:     fmt.Sprintf("failed to get current directory: %v", err),
				IsError:    true,
			}
		}
		workDir = cwd
	}

	cmd := exec.CommandContext(ctx, "bash", "-lc", input.Command)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		if len(output) == 0 {
			return schema.ToolResult{
				ToolCallID: call.ID,
				Output:     fmt.Sprintf("command failed: %v", err),
				IsError:    true,
			}
		}

		return schema.ToolResult{
			ToolCallID: call.ID,
			Output:     strings.TrimRight(string(output), "\n"),
			IsError:    true,
		}
	}

	return schema.ToolResult{
		ToolCallID: call.ID,
		Output:     strings.TrimRight(string(output), "\n"),
		IsError:    false,
	}
}
