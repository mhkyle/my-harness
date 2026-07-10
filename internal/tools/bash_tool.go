package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"mhkyle/my-harness/internal/schema"
)

type BashTool struct {
}

type bashInput struct {
	Command string `json:"command"`
	WorkDir string `json:"workdir,omitempty"`
}

func NewBashTool() BaseTool {
	return &BashTool{}
}

func (m *BashTool) Name() string {
	return "bash"
}

func (m *BashTool) Definition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        "bash",
		Description: "Execute bash commands in the current directory.",
		InputSchema: string(`{"type":"object","properties":{"command":{"type":"string","description":"Bash command to execute."},"workdir":{"type":"string","description":"Optional working directory for command execution."}},"required":["command"]}`),
	}
}

func (m *BashTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var input bashInput
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %v", err)
	}

	if strings.TrimSpace(input.Command) == "" {
		return "", fmt.Errorf("command is required")
	}

	workDir := input.WorkDir
	if workDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %v", err)
		}
		workDir = cwd
	}

	cmd := exec.CommandContext(ctx, "bash", "-lc", input.Command)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		if len(output) == 0 {
			return "", fmt.Errorf("command failed: %v", err)
		}

		return strings.TrimRight(string(output), "\n"), err
	}

	return strings.TrimRight(string(output), "\n"), nil
}
