package main

import (
	"context"
	"os"

	"mhkyle/my-harness/internal/engine"
	"mhkyle/my-harness/internal/provider"
	bashtool "mhkyle/my-harness/internal/tools/bash"
)

func main() {
	workDir, _ := os.Getwd()

	provider := provider.NewMockProvider()
	registry := bashtool.NewBashTool()

	eng := engine.NewAgentEngine(provider, registry, workDir)
	err := eng.Run(context.Background(), "Please help to check the files in current directory.")
	if err != nil {
		panic(err)
	}
}
