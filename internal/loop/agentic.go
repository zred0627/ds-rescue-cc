package loop

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/zred0627/ds-rescue-cc/internal/deepseek"
)

const MaxRounds = 20

// toolResult holds one dispatched tool's outcome, keyed by tool_call_id.
// Used to preserve OpenAI-compatible ordering when dispatching concurrently.
type toolResult struct {
	callID  string
	content string
}

// Run executes the agentic tool-use loop. Returns final assistant content + rounds used.
// When the model returns multiple tool_calls in a single response, they are dispatched
// concurrently via goroutines (DeepSeek V4-Pro/Flash supports up to 128 parallel calls).
// Results are written back in the original tool_calls order to satisfy the OpenAI-compatible
// invariant that tool messages must follow the same order as the assistant's tool_calls array.
func Run(client *deepseek.Client, req deepseek.ChatRequest) (string, int, error) {
	messages := req.Messages
	for round := 1; round <= MaxRounds; round++ {
		req.Messages = messages
		resp, err := client.Call(req)
		if err != nil {
			return "", round, fmt.Errorf("round %d: %w", round, err)
		}
		if len(resp.Choices) == 0 {
			return "", round, fmt.Errorf("round %d: empty choices", round)
		}
		choice := resp.Choices[0]
		messages = append(messages, choice.Message)

		if len(choice.Message.ToolCalls) == 0 {
			return choice.Message.Content, round, nil
		}

		messages = append(messages, dispatchParallel(choice.Message.ToolCalls)...)
	}
	return "", MaxRounds, fmt.Errorf("max rounds (%d) exceeded without final answer", MaxRounds)
}

// dispatchParallel runs every tool_call concurrently and returns the resulting tool
// messages in the SAME order as the input slice (OpenAI-compatible requirement).
//
// For a single tool_call, falls back to direct dispatch to avoid goroutine overhead.
// For ≥2 tool_calls, spawns one goroutine per call and joins via sync.WaitGroup.
// Total wall-clock latency ≈ max(individual tool exec) instead of sum().
func dispatchParallel(calls []deepseek.ToolCall) []deepseek.Message {
	results := make([]toolResult, len(calls))

	if len(calls) == 1 {
		results[0] = runOne(calls[0])
	} else {
		var wg sync.WaitGroup
		for i := range calls {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				results[idx] = runOne(calls[idx])
			}(i)
		}
		wg.Wait()
	}

	out := make([]deepseek.Message, len(results))
	for i, r := range results {
		out[i] = deepseek.Message{
			Role:       "tool",
			ToolCallID: r.callID,
			Content:    r.content,
		}
	}
	return out
}

// runOne dispatches a single tool_call and returns its result (never panics on tool error).
func runOne(tc deepseek.ToolCall) toolResult {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		args = map[string]interface{}{"_parse_error": err.Error()}
	}
	result, toolErr := DispatchTool(tc.Function.Name, args)
	if toolErr != nil {
		result = fmt.Sprintf("ERROR: %s\nOUTPUT: %s", toolErr.Error(), result)
	}
	return toolResult{callID: tc.ID, content: result}
}
