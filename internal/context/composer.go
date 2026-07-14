// internal/context/composer.go
package context

import (
	"os"
	"path/filepath"
	"strings"

	"mhkyle/my-harness/internal/schema"
)

type PromptComposer struct {
	workDir     string
	skillLoader *SkillLoader
}

func NewPromptComposer(workDir string) *PromptComposer {
	return &PromptComposer{
		workDir:     workDir,
		skillLoader: NewSkillLoader(workDir),
	}
}

func (c *PromptComposer) Build() schema.Message {
	var promptBuilder strings.Builder

	promptBuilder.WriteString(`# core identity and discipline
	you are a highly skilled AI coding assistant, capable of understanding and manipulating code in various programming languages. 
	You have access to a set of built-in tools that allow you to read, write, edit, and execute code within the workspace. 
	Your primary goal is to assist the user in achieving their coding objectives efficiently and effectively.

	# core discipline
	1. You must always use the tools provided to you for any file operations, including reading, writing, and editing files. 
	2. You must not attempt to access or manipulate files outside of the workspace or use any tools that are not explicitly provided to you.
	3. You must not execute any code or commands that could potentially harm the system or compromise security.
	4. You must always provide clear and concise explanations for your actions and decisions, especially when modifying code or files.
	5. You must always prioritize the user's objectives and work collaboratively with them to achieve their goals.
	6. You must always adhere to best practices in coding, including writing clean, maintainable, and efficient code.
	7. You must always respect the user's privacy and confidentiality, and never share any information about their code or projects without their explicit consent.
	8. You must use them when appropriate to enhance your capabilities and provide more effective assistance, when the skills are available in the workspace.
`)

	agentsMDPath := filepath.Join(c.workDir, "AGENTS.md")
	content, err := os.ReadFile(agentsMDPath)
	if err == nil {
		promptBuilder.WriteString("\n# Project-specific rules (from AGENTS.md)\n")
		promptBuilder.WriteString("The following are the architecture specifications and precautions specific to the current workspace, and your actions must strictly comply with the following requirements:\n")
		promptBuilder.WriteString("```markdown\n")
		promptBuilder.WriteString(string(content))
		promptBuilder.WriteString("\n```\n")
	}

	skillsContent := c.skillLoader.LoadAll()
	if skillsContent != "" {
		promptBuilder.WriteString(skillsContent)
	}

	return schema.Message{
		Role:    schema.RoleSystem,
		Content: promptBuilder.String(),
	}
}
