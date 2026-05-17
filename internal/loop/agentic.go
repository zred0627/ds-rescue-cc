package loop

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/zred0627/ds-rescue-cc/internal/deepseek"
)

const MaxRounds = 20

type Options struct {
	MaxExecTimeoutSec int
}

type toolResult struct {
	callID  string
	content string
}

func Run(client *deepseek.Client, req deepseek.ChatRequest, opts Options) (string, int, error) {
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

		messages = append(messages, dispatchParallel(choice.Message.ToolCalls, opts)...)
	}
	return "", MaxRounds, fmt.Errorf("max rounds (%d) exceeded without final answer", MaxRounds)
}

func dispatchParallel(calls []deepseek.ToolCall, opts Options) []deepseek.Message {
	results := make([]toolResult, len(calls))

	if len(calls) == 1 {
		results[0] = runOne(calls[0], opts)
	} else {
		var wg sync.WaitGroup
		for i := range calls {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				results[idx] = runOne(calls[idx], opts)
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

func runOne(tc deepseek.ToolCall, opts Options) toolResult {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
		args = map[string]interface{}{"_parse_error": err.Error()}
	}
	result, toolErr := DispatchTool(tc.Function.Name, args, opts.MaxExecTimeoutSec)
	if toolErr != nil {
		result = fmt.Sprintf("ERROR: %s\nOUTPUT: %s", toolErr.Error(), result)
	}
	return toolResult{callID: tc.ID, content: result}
}
