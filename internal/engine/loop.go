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

const (
	maxHistorySize = 100
	maxContextSize = 36000
)

type AgentEngine struct {
	provider       provider.LLMProvider
	registry       tools.Registry
	EnableThinking bool
	PlanMode       bool
	composer       *contextComposer.PromptComposer
	compactor      contextComposer.Compactor
	recovery       contextComposer.RecoveryManager
}

func NewAgentEngine(p provider.LLMProvider, r tools.Registry, workDir string, enableThinking, planMode bool) *AgentEngine {
	return &AgentEngine{
		provider:       p,
		registry:       r,
		EnableThinking: enableThinking,
		PlanMode:       planMode,
		composer:       contextComposer.NewPromptComposer(workDir, planMode),
		compactor:      contextComposer.NewStaticCompactor(maxContextSize, maxHistorySize),
		recovery:       contextComposer.NewSimpleRecoveryManager(),
	}
}

func (e *AgentEngine) Run(ctx context.Context, session *Session, reporter Reporter) error {
	log.Printf("[Engine] Start AgentEngine with DIR %s\n", session.WorkDir)

	// adding skills dynamical promots
	e.composer = contextComposer.NewPromptComposer(session.WorkDir, e.PlanMode)
	systemPrompt := e.composer.Build()

	turn := 0
	for {
		turn++
		log.Printf("[Engine] Turn %d\n", turn)
		var contextHistory []schema.Message
		workingMem := session.GetWorkingMemory(maxHistorySize)
		contextHistory = append(contextHistory, systemPrompt)
		contextHistory = append(contextHistory, workingMem...)

		// compact the context history if it exceeds the limit
		compactedContext := e.compactor.Compact(contextHistory)
		log.Printf("[Engine] Current context history length: %d\n", len(compactedContext))

		// Two-Stage ReAct
		if e.EnableThinking {
			log.Println("[Engine] Thinking mode enabled. Start thinking...")
			thinkResp, err := e.provider.Generate(ctx, compactedContext, nil)
			if err != nil {
				return fmt.Errorf("failed to generate thinking response: %v", err)
			}

			if thinkResp.Content != "" {
				log.Printf("💭 Model Thinking: %s\n", thinkResp.Content)
				compactedContext = append(compactedContext, *thinkResp)
			}
		}

		availableTools := e.registry.GetAvailableTools()
		actionMsg, err := e.provider.Generate(ctx, compactedContext, availableTools)
		if err != nil {
			return fmt.Errorf("failed to generate response: %v", err)
		}

		session.Append(*actionMsg)
		log.Printf("[Engine] Model Response: %v\n", actionMsg)
		compactedContext = append(compactedContext, *actionMsg)

		if actionMsg.Content != "" && reporter != nil {
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

				finalOutput := result.Output
				if result.IsError {
					finalOutput = e.recovery.AnalyzeAndInject(toolCall.Name, result.Output)
					log.Printf("  -> ⚠️ Tool %s returned an error: %s\n", tc.Name, finalOutput)
				} else {
					log.Printf("  -> ✅ Tool %s executed successfully. Output: %s\n", tc.Name, finalOutput)
				}
				if reporter != nil {
					displayOutput := result.Output
					if len(displayOutput) > maxContextSize {
						displayOutput = displayOutput[:maxContextSize] + "...(truncated)"
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
