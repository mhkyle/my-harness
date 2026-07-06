package engine

import (
	"context"
	"fmt"
	"log"
	"sync"

	"mhkyle/my-harness/internal/provider"
	"mhkyle/my-harness/internal/schema"
	"mhkyle/my-harness/internal/tools"
)

type AgentEngine struct {
	provider       provider.LLMProvider
	registry       tools.Registry
	WorkDir        string
	EnableThinking bool
}

func NewAgentEngine(p provider.LLMProvider, r tools.Registry, workDir string, enableThinking bool) *AgentEngine {
	return &AgentEngine{
		provider:       p,
		registry:       r,
		WorkDir:        workDir,
		EnableThinking: enableThinking,
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

		// ReAct
		if e.EnableThinking {
			log.Println("[Engine] Thinking mode enabled. Start thinking...")
			thinkResp, err := e.provider.Generate(ctx, contextHistory, nil)
			if err != nil {
				return fmt.Errorf("failed to generate thinking response: %v", err)
			}

			if thinkResp.Content != "" {
				log.Printf("💭 Model Thinking: %s\n", thinkResp.Content)
				contextHistory = append(contextHistory, *thinkResp)
			}
		}

		availableTools := e.registry.GetAvailableTools()
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

		var wg sync.WaitGroup
		wg.Add(len(responseMsg.ToolCalls))
		var tempContextHistory = make([]schema.Message, len(responseMsg.ToolCalls))
		for i, toolCall := range responseMsg.ToolCalls {
			go func(tc schema.ToolCall, index int) {
				defer wg.Done()
				log.Printf("  -> 🛠️ Executing tool: %s, Arguments: %s\n", tc.Name, string(tc.Arguments))
				result := e.registry.Execute(ctx, tc)

				if result.IsError {
					log.Printf("  -> ❌ Tool execution error: %s\n", result.Output)
				} else {
					log.Printf("  -> ✅ Tool executed successfully (returned %d bytes)\n", len(result.Output))
				}
				observationMsg := schema.Message{
					Role:       schema.RoleUser,
					Content:    result.Output,
					ToolCallID: tc.ID,
				}
				tempContextHistory[index] = observationMsg
			}(toolCall, i)
		}
		wg.Wait()
		contextHistory = append(contextHistory, tempContextHistory...)
	}

	return nil
}
