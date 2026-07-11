package main

import (
	"context"
	"log"
	"os"

	"mhkyle/my-harness/internal/engine"
	"mhkyle/my-harness/internal/provider"
	"mhkyle/my-harness/internal/tools"
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

	registry := tools.NewRegistry()
	registry.Register(tools.NewReadFileTool(workDir))
	registry.Register(tools.NewWriteFileTool(workDir))
	registry.Register(tools.NewBashTool())

	query := `请帮我执行以下操作： 
	1. 用 bash 查看一下我当前电脑的 Go 版本。 
	2. 帮我写一个简单的 helloworld.go 文件，输出 "Hello, harness!"。 
	3. 用 bash 编译并运行这个 go 文件，确认它能正常工作。`
	EnableThinking := false

	eng := engine.NewAgentEngine(provider, registry, workDir, EnableThinking)
	err = eng.Run(context.Background(), query)
	if err != nil {
		panic(err)
	}
}
