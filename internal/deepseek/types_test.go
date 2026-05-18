package deepseek

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestAssistantWithToolCallsMustSerializeContent reproduces the v0.1.1 bug:
// DeepSeek API rejects assistant messages that omit the `content` field when
// `tool_calls` is present (HTTP 400 `missing field 'content'`). Go's
// encoding/json drops Content when its zero value is "" and the tag carries
// omitempty, so a round-tripped assistant message (Content="" + tool_calls)
// serialises without a content field.
func TestAssistantWithToolCallsMustSerializeContent(t *testing.T) {
	msg := Message{
		Role:    "assistant",
		Content: "",
		ToolCalls: []ToolCall{
			{
				ID:   "call_1",
				Type: "function",
				Function: ToolCallFunc{
					Name:      "read_file",
					Arguments: `{"path":"a.txt"}`,
				},
			},
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(body)

	if !strings.Contains(got, `"content"`) {
		t.Fatalf("assistant message with tool_calls must include content field, got: %s", got)
	}
}

// TestToolMessageStillSerializesContent guards against regressions in role=tool
// messages, which always carry a non-empty Content but should never lose the
// field after the omitempty fix.
func TestToolMessageStillSerializesContent(t *testing.T) {
	msg := Message{
		Role:       "tool",
		ToolCallID: "call_1",
		Content:    "file contents",
	}
	body, _ := json.Marshal(msg)
	got := string(body)
	if !strings.Contains(got, `"content":"file contents"`) {
		t.Fatalf("tool message content lost, got: %s", got)
	}
}

// TestReasoningContentRoundTrip guards the thinking-mode contract: DeepSeek
// requires reasoning_content to be echoed back on intermediate assistant
// messages that also carry tool_calls. The struct must (a) unmarshal the field
// from a server response, and (b) re-marshal it on the next request so the
// content is preserved across the agentic loop's append-and-resend path.
func TestReasoningContentRoundTrip(t *testing.T) {
	serverReply := `{"role":"assistant","content":"","reasoning_content":"I need to read the file first.","tool_calls":[{"id":"call_1","type":"function","function":{"name":"read_file","arguments":"{}"}}]}`

	var msg Message
	if err := json.Unmarshal([]byte(serverReply), &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if msg.ReasoningContent != "I need to read the file first." {
		t.Fatalf("reasoning_content lost during unmarshal, got: %q", msg.ReasoningContent)
	}

	body, _ := json.Marshal(msg)
	if !strings.Contains(string(body), `"reasoning_content":"I need to read the file first."`) {
		t.Fatalf("reasoning_content not echoed on re-marshal, got: %s", body)
	}
}

// TestSystemAndUserMessagesUnaffected guards against accidental schema drift
// for the common roles supplied by dispatcher.go (system + user prompts).
func TestSystemAndUserMessagesUnaffected(t *testing.T) {
	cases := []Message{
		{Role: "system", Content: "you are helpful"},
		{Role: "user", Content: "hi"},
	}
	for _, m := range cases {
		body, _ := json.Marshal(m)
		if !strings.Contains(string(body), `"content":"`) {
			t.Fatalf("role=%s lost content field: %s", m.Role, body)
		}
	}
}
