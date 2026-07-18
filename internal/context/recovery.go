// internal/context/recovery.go
package context

import (
	"fmt"
	"strings"
)

type RecoveryManager interface {
	AnalyzeAndInject(toolName string, rawError string) string
}

// SimpleRecoveryManager will analyze the raw error messages from tools and provide enhanced error messages with actionable hints.
type SimpleRecoveryManager struct{}

func NewSimpleRecoveryManager() RecoveryManager {
	return &SimpleRecoveryManager{}
}

// AnalyzeAndInject analyzes the raw error message from a tool and injects actionable hints to help the user recover from the error.
func (rm *SimpleRecoveryManager) AnalyzeAndInject(toolName string, rawError string) string {
	var hint string

	// we use relatively stable English system-level error keywords, or our own fixed error formats within the tools.
	lowerError := strings.ToLower(rawError)

	switch toolName {
	case "edit_file":
		// Match the fixed error thrown by the fuzzyReplace we wrote in Lesson 07
		if strings.Contains(rawError, "failed to find old_text") || strings.Contains(rawError, "failed to find the old_text") {
			hint = "The old_text provided does not match the current content of the file, or it lacks necessary indentation. Please use the `read_file` tool to read the file again to get the latest and accurate content before re-initiating the edit."
		} else if strings.Contains(rawError, "match more than one") || strings.Contains(rawError, "provide more context") {
			hint = "The old_text is not specific enough and matches multiple identical code blocks. Please add a few lines of code above and below the old_text to ensure the uniqueness of the replacement."
		}

	case "read_file", "write_file":
		// Match POSIX standard errors thrown by the native Go os package
		if strings.Contains(lowerError, "no such file or directory") {
			hint = "The path seems incorrect. Do not guess blindly; use the `bash` command `ls -la` or `find . -name` to locate the correct directory structure and file name."
		} else if strings.Contains(lowerError, "permission denied") {
			hint = "You do not have permission to operate on this file. Please check workspace restrictions, or consider whether you need to modify other files."
		}

	case "bash":
		if strings.Contains(lowerError, "command not found") {
			hint = "The command is not installed on the system. Consider whether there is an alternative command, or if you need to write a script to install it first."
		} else if strings.Contains(rawError, "timeout") || strings.Contains(rawError, "DeadlineExceeded") {
			// Match the 30s context.WithTimeout error we wrote
			hint = "The command execution was killed due to a timeout. If it is a long-running service (e.g., server or watch), please run it in the background (e.g., using `nohup ... &`) to avoid blocking the main thread."
		} else if strings.Contains(lowerError, "syntax error") {
			hint = "Bash syntax error. Please check for proper quoting and special characters to ensure the command can run directly in the terminal."
		}
	}

	if hint == "" {
		return rawError
	}

	return fmt.Sprintf("%s\n\n[System Recovery Guide]: %s", rawError, hint)
}
