package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"mhkyle/my-harness/internal/engine"
	"mhkyle/my-harness/internal/provider"
	"mhkyle/my-harness/internal/schema"
	"mhkyle/my-harness/internal/tools"
)

func main() {
	workDir, _ := os.Getwd()

	// models
	apiKey, err := os.ReadFile("/Users/minghyuan/.ebay-claude-code.txt")
	if err != nil {
		log.Fatalf("failed to read API key: %v", err)
	}
	provider := provider.NewZhipuOpenAIProvider("https://hubgptgatewaysvc.vip.qa.ebay.com/gateway/v1/",
		string(apiKey), "hubgpt-chat-completions-dedicated")
	if provider == nil {
		log.Fatalf("failed to initialize Zhipu OpenAI provider")
	}

	// go run cmd/my-harness/main.go -prompt="搭建一个极简的 Go 语言 Web Server 项目在 server.go 中"
	promptPtr := flag.String("prompt", "", "Prompt to send to the AI")
	flag.Parse()
	if *promptPtr == "" {
		fmt.Println("Please provide a prompt using the -prompt flag.")
		os.Exit(1)
	}

	registry := tools.NewRegistry()
	registry.Register(tools.NewReadFileTool(workDir))
	registry.Register(tools.NewWriteFileTool(workDir))
	registry.Register(tools.NewEditFileTool(workDir))
	registry.Register(tools.NewBashTool())

	reporter := engine.NewTerminalReporter()
	// nil session
	s := engine.GlobalSessionMgr.GetOrCreate("chat1", workDir)
	s.Append(schema.Message{
		Role:    schema.RoleUser,
		Content: *promptPtr,
	})

	EnableThinking := true
	enablePlanMode := true

	eng := engine.NewAgentEngine(provider, registry, workDir, EnableThinking, enablePlanMode)
	if err := eng.Run(context.Background(), s, reporter); err != nil {
		panic(err)
	}
}
