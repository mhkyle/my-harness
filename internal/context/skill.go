// internal/context/skill.go
package context

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Skill struct {
	Name        string
	Description string
	Body        string
}

type SkillLoader struct {
	workDir string
}

func NewSkillLoader(workDir string) *SkillLoader {
	return &SkillLoader{workDir: workDir}
}

func (s *SkillLoader) LoadAll() string {
	// default read claude skills from .claude/skills directory
	skillBaseDir := filepath.Join(s.workDir, ".claude", "skills")
	fmt.Printf("use the base dir: %+v\n", skillBaseDir)

	if _, err := os.Stat(skillBaseDir); os.IsNotExist(err) {
		return ""
	}

	var skillsBuilder strings.Builder
	skillsBuilder.WriteString("\n### available skills (Agent Skills)\n")
	skillsBuilder.WriteString("The following are the standardized plug-in skills you have. Please strictly follow the instructions in the body according to the scenarios described in the description:\n\n")

	err := filepath.WalkDir(skillBaseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "SKILL.md" {
			fmt.Printf("Found skill file: %s\n", path)
			content, err := os.ReadFile(path)
			if err == nil {
				skill := parseSkillMD(string(content))

				skillsBuilder.WriteString(fmt.Sprintf("#### skill name: %s\n", skill.Name))
				skillsBuilder.WriteString(fmt.Sprintf("**trigger condition**: %s\n\n", skill.Description))
				skillsBuilder.WriteString("**execution guide**:\n")
				skillsBuilder.WriteString(skill.Body)
				skillsBuilder.WriteString("\n\n---\n")
			}
		}
		return nil
	})

	if err != nil || skillsBuilder.Len() < 100 {
		return ""
	}

	return skillsBuilder.String()
}

func parseSkillMD(content string) Skill {
	skill := Skill{
		Name:        "Unknown Skill",
		Description: "No description provided.",
		Body:        content,
	}

	if strings.HasPrefix(content, "---\n") || strings.HasPrefix(content, "---\r\n") {
		parts := strings.SplitN(content, "---", 3)
		if len(parts) == 3 {
			frontmatter := parts[1]
			skill.Body = strings.TrimSpace(parts[2])

			lines := strings.Split(frontmatter, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "name:") {
					skill.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
				} else if strings.HasPrefix(line, "description:") {
					skill.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
				}
			}
		}
	}

	return skill
}
