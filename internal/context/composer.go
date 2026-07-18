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
	planMode    bool
	skillLoader *SkillLoader
}

func NewPromptComposer(workDir string, planMode bool) *PromptComposer {
	return &PromptComposer{
		workDir:     workDir,
		planMode:    planMode,
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

	if c.planMode {
		promptBuilder.WriteString(`
		# plan mode: On
		In this mode, you cannot execute the commands directly, you must provide a detailed plan of the steps and actions in the physical local files.
		When receiving a user request, you must do the following:
		** STEP 1: Bootstrapping **
		- Use base tool (ls -la) check current workspace files. Check if the files "PLAN.md" and "TODO.md" are exist.
		-- if both files not exist, create them with "write_file" tool. Create "PLAN.md" first, write your understanding, constructure plans and technical details in it. 
		   Then create "TODO.md", split the plans into different doable tasks with standard "MARKDOWN Checklist" format, like "- [ ] Step 1: Do something".
		-- if both files exist, DO NOT modify or replace them!!! Which means system may get restarted, and human may have already modified the files. You must read the files with read_file tool, and analyze the content, then update your plans and tasks in your mind. You can add new tasks to "TODO.md" if necessary, but DO NOT remove or modify existing tasks.

		** STEP 2: Strictly Execute the single step and update the checks **
		- Start to execute the tasks in "TODO.md" one by one, strictly follow the order of the tasks.
		- **Strict requirements **: When you complete a task, you must update the TODO.md file, mark the task as completed with "[x]", and then read the updated TODO.md file to check if there are any new tasks added by human. If there are new tasks, you must add them to your mind and continue to execute them in order.
		- MUST NOT skip any tasks.
		- MUST execute the tasks one by one, and only after completing a task, you check the TODO.md and continue to the next task.

		** STEP 3 **
		When you encounter any problems that the tools reports, and do not know what to do next, you must use the read_file tool to read the "TODO.md" to locate your current task, and then analyze the content of the task to find out what is the next step you should do. If you still cannot figure out what to do, you must ask human for help, and wait for human's response. You must not continue to execute any tasks until you get a clear instruction from human.
		
		`)
	}

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
