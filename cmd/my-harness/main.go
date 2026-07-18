package main

import (
	"context"
	"log"
	"os"

	"mhkyle/my-harness/internal/engine"
	"mhkyle/my-harness/internal/provider"
	"mhkyle/my-harness/internal/schema"
	"mhkyle/my-harness/internal/tools"
)

func main() {
	workDir, _ := os.Getwd()

	// apiKey, err := os.ReadFile("/Users/minghyuan/.ebay-claude-code.txt")
	// if err != nil {
	// 	log.Fatalf("failed to read API key: %v", err)
	// }
	// provider := provider.NewZhipuOpenAIProvider("https://hubgptgatewaysvc.vip.qa.ebay.com/gateway/v1/",
	// 	string(apiKey), "hubgpt-chat-completions-dedicated")
	// if provider == nil {
	// 	log.Fatalf("failed to initialize Zhipu OpenAI provider")
	// }

	// local ollama server
	provider := provider.NewZhipuOpenAIProvider("http://localhost:11434/v1",
		"fake", "qwen3.6:latest")
	if provider == nil {
		log.Fatalf("failed to initialize Zhipu OpenAI provider")
	}

	registry := tools.NewRegistry()
	registry.Register(tools.NewReadFileTool(workDir))
	registry.Register(tools.NewWriteFileTool(workDir))
	registry.Register(tools.NewEditFileTool(workDir))
	registry.Register(tools.NewBashTool())

	reporter := engine.NewTerminalReporter()
	// nil session
	userQuery := `
	What this project is about in current directory? Tell me more details about the project, and give me a list of all the files in the project with their file sizes.
	`
	s := engine.GlobalSessionMgr.GetOrCreate("chat1", workDir)
	s.Append(schema.Message{
		Role:    schema.RoleUser,
		Content: userQuery,
	})

	EnableThinking := false
	eng := engine.NewAgentEngine(provider, registry, workDir, EnableThinking)
	err := eng.Run(context.Background(), s, reporter)
	if err != nil {
		panic(err)
	}
}
