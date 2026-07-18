package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"mhkyle/my-harness/internal/schema"
)

const (
	maxLengthBytes = 8000
)

type ReadFileTool struct {
	workDir string
}

type readFileArgs struct {
	Path string `json:"path"`
}

func NewReadFileTool(workDir string) *ReadFileTool {
	return &ReadFileTool{
		workDir: workDir,
	}
}

func (m *ReadFileTool) Name() string {
	return "read_file"
}

func (m *ReadFileTool) Definition() schema.ToolDefinition {
	return schema.ToolDefinition{
		Name:        m.Name(),
		Description: "Read the content of a file in the current directory.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to read.",
				},
			},
			"required": []string{"path"},
		},
	}
}

func (m *ReadFileTool) Execute(ctx context.Context, args json.RawMessage) (string, error) {
	var input readFileArgs
	if err := json.Unmarshal(args, &input); err != nil {
		return "", fmt.Errorf("invalid arguments: %v", err)
	}

	// Prevent path traversal by ensuring the resolved path stays within workDir.
	cleanPath := filepath.Clean(input.Path)
	if filepath.IsAbs(cleanPath) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}

	baseAbs, err := filepath.Abs(m.workDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base dir: %v", err)
	}
	candidateAbs, err := filepath.Abs(filepath.Join(baseAbs, cleanPath))
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %v", err)
	}
	rel, err := filepath.Rel(baseAbs, candidateAbs)
	if err != nil {
		return "", fmt.Errorf("failed to compute relative path: %v", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes working directory")
	}

	file, err := os.Open(candidateAbs)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	if len(content) > maxLengthBytes {
		truncatedContent := content[:maxLengthBytes]
		fmt.Printf("content truncated to %d bytes\n", maxLengthBytes)
		return string(truncatedContent), nil
	}

	return string(content), nil
}
