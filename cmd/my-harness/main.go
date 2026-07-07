package main

import (
	"context"
	"log"
	"os"

	"mhkyle/my-harness/internal/engine"
	"mhkyle/my-harness/internal/provider"
	bashTool "mhkyle/my-harness/internal/tools/bash"
)

func main() {
	workDir, _ := os.Getwd()

	apiKey, err := os.ReadFile("/Users/minghyuan/.ebay-claude-code.txt")
	if err != nil {
		log.Fatalf("failed to read API key: %v", err)
	}

	provider := provider.NewZhipuOpenAIProvider("https://hubgptgatewaysvc.vip.qa.ebay.com/gateway/v1/",
		string(apiKey), "hubgpt-chat-completions-dedicated")
	if provider == nil {
		log.Fatalf("failed to initialize Zhipu OpenAI provider")
	}

	registry := bashTool.NewBashTool()

	query := "Please help to review the codes in current directory, and tell me what it is mainly doing."

	eng := engine.NewAgentEngine(provider, registry, workDir, true)
	err = eng.Run(context.Background(), query)
	if err != nil {
		panic(err)
	}
}
