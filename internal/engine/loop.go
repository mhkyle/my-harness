package engine

import (
	"context"
	"fmt"
	"log"
	"sync"

	contextComposer "mhkyle/my-harness/internal/context"
	"mhkyle/my-harness/internal/provider"
	"mhkyle/my-harness/internal/schema"
	"mhkyle/my-harness/internal/tools"
)

type AgentEngine struct {
	provider       provider.LLMProvider
	registry       tools.Registry
	EnableThinking bool
	composer       *contextComposer.PromptComposer
}

func NewAgentEngine(p provider.LLMProvider, r tools.Registry, workDir string, enableThinking bool) *AgentEngine {
	return &AgentEngine{
		provider:       p,
		registry:       r,
		EnableThinking: enableThinking,
		composer:       contextComposer.NewPromptComposer(workDir),
	}
}

func (e *AgentEngine) Run(ctx context.Context, session *Session, reporter Reporter) error {
	log.Printf("[Engine] Start AgentEngine with DIR %s\n", session.WorkDir)

	// adding skills dynamical promots
	systemPrompt := e.composer.Build()

	turn := 0
	for {
		turn++
		log.Printf("[Engine] Turn %d\n", turn)
		var contextHistory []schema.Message
		workingMem := session.GetWorkingMemory(6)
		contextHistory = append(contextHistory, systemPrompt)
		contextHistory = append(contextHistory, workingMem...)
		log.Printf("[Engine] Current context history length: %d\n", len(contextHistory))

		// Two-Stage ReAct
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
		actionMsg, err := e.provider.Generate(ctx, contextHistory, availableTools)
		if err != nil {
			return fmt.Errorf("failed to generate response: %v", err)
		}

		log.Printf("[Engine] Model Response: %v\n", actionMsg)
		contextHistory = append(contextHistory, *actionMsg)

		if actionMsg.Content != "" && reporter != nil {
			fmt.Printf("🤖 Model: %s\n", actionMsg.Content)
			reporter.OnMessage(ctx, actionMsg.Content)
		}

		if len(actionMsg.ToolCalls) == 0 {
			log.Println("[Engine] Task completed, exiting loop.")
			break
		}

		log.Printf("[Engine] Model requested to call %d tools...\n", len(actionMsg.ToolCalls))

		var wg sync.WaitGroup
		wg.Add(len(actionMsg.ToolCalls))
		var tempContextHistory = make([]schema.Message, len(actionMsg.ToolCalls))
		for i, toolCall := range actionMsg.ToolCalls {
			go func(tc schema.ToolCall, index int) {
				defer wg.Done()
				log.Printf("  -> 🛠️ Executing tool: %s, Arguments: %s\n", tc.Name, string(tc.Arguments))
				result := e.registry.Execute(ctx, tc)

				if reporter != nil {
					displayOutput := result.Output
					if len(displayOutput) > 200 {
						displayOutput = displayOutput[:200] + "...(truncated)"
					}
					reporter.OnToolResult(ctx, tc.Name, displayOutput, result.IsError)
				}

				tempContextHistory[index] = schema.Message{
					Role:       schema.RoleUser,
					Content:    result.Output,
					ToolCallID: tc.ID,
				}
			}(toolCall, i)
		}
		wg.Wait()
		session.Append(tempContextHistory...)
	}

	return nil
}
