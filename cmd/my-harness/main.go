package main

import (
	"context"
	"os"

	"mhkyle/my-harness/internal/engine"
	"mhkyle/my-harness/internal/provider"
	"mhkyle/my-harness/internal/tools"
)

func main() {
	workDir, _ := os.Getwd()

	provider := provider.NewMockProvider()
	registry := tools.NewMockRegistry()

	eng := engine.NewAgentEngine(provider, registry, workDir, true)
	err := eng.Run(context.Background(), "Please help to check the files in current directory.")
	if err != nil {
		panic(err)
	}
}
