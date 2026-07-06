package engine

import (
	"context"
	"fmt"
	"log"

	"mhkyle/my-harness/internal/provider"
	"mhkyle/my-harness/internal/schema"
	"mhkyle/my-harness/internal/tools"
)

type AgentEngine struct {
	provider provider.LLMProvider
	registry tools.Registry

	WorkDir string
}

func NewAgentEngine(p provider.LLMProvider, r tools.Registry, workDir string) *AgentEngine {
	return &AgentEngine{
		provider: p,
		registry: r,
		WorkDir:  workDir,
	}
}

func (e *AgentEngine) Run(ctx context.Context, userPrompt string) error {
	log.Printf("[Engine] Start AgentEngine with DIR %s\n", e.WorkDir)

	contextHistory := []schema.Message{
		{
			Role:    schema.RoleSystem,
			Content: "You are a harness project, an expert coding assistant. You have full access to tools in the workspace.",
		},
		{
			Role:    schema.RoleUser,
			Content: userPrompt,
		},
	}

	turnCount := 0

	for {
		turnCount++
		log.Printf("========== [Turn %d] Start ==========\n", turnCount)

		availableTools := e.registry.GetAvailableTools()

		log.Println("[Engine] Start reasoning...")
		responseMsg, err := e.provider.Generate(ctx, contextHistory, availableTools)
		if err != nil {
			return fmt.Errorf("failed to generate response: %v", err)
		}

		contextHistory = append(contextHistory, *responseMsg)

		if responseMsg.Content != "" {
			fmt.Printf("🤖 Model: %s\n", responseMsg.Content)
		}

		if len(responseMsg.ToolCalls) == 0 {
			log.Println("[Engine] Task completed, exiting loop.")
			break
		}

		log.Printf("[Engine] Model requested to call %d tools...\n", len(responseMsg.ToolCalls))

		for _, toolCall := range responseMsg.ToolCalls {
			log.Printf("  -> 🛠️ Executing tool: %s, Arguments: %s\n", toolCall.Name, string(toolCall.Arguments))

			// Route and execute the underlying tool through the Registry
			result := e.registry.Execute(ctx, toolCall)

			if result.IsError {
				log.Printf("  -> ❌ Tool execution error: %s\n", result.Output)
			} else {
				log.Printf("  -> ✅ Tool executed successfully (returned %d bytes)\n", len(result.Output))
			}

			// Encapsulate the tool execution observation as a User Message and append it to the context
			// Note: ToolCallID must be included! This is crucial for maintaining the reasoning chain of the LLM
			observationMsg := schema.Message{
				Role:       schema.RoleUser,
				Content:    result.Output,
				ToolCallID: toolCall.ID,
			}
			contextHistory = append(contextHistory, observationMsg)
		}
	}

	return nil
}
