package loop

import (
	"encoding/json"
	"fmt"

	"github.com/zred0627/ds-rescue-cc/internal/deepseek"
)

const MaxRounds = 20

// Run executes the agentic tool-use loop. Returns final assistant content + rounds used.
// Tool calls are dispatched sequentially.
// TODO: upgrade to parallel when DeepSeek confirms support for parallel_tool_calls
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

		for _, tc := range choice.Message.ToolCalls {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				args = map[string]interface{}{"_parse_error": err.Error()}
			}
			result, toolErr := DispatchTool(tc.Function.Name, args)
			if toolErr != nil {
				result = fmt.Sprintf("ERROR: %s\nOUTPUT: %s", toolErr.Error(), result)
			}
			messages = append(messages, deepseek.Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    result,
			})
		}
	}
	return "", MaxRounds, fmt.Errorf("max rounds (%d) exceeded without final answer", MaxRounds)
}
