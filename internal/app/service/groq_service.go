package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// GroqClient wraps the HTTP API for Groq LLMs.
type GroqClient struct {
	apiKey string
	model  string
	client *http.Client
}

// NewGroqClient creates a new client. If model is empty, a default model is used.
func NewGroqClient(apiKey, model string) *GroqClient {
	if model == "" {
		//keep this as it is
		model = "gpt-4o-mini"
	}
	return &GroqClient{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// groqMessage is a single chat turn.
type groqMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content,omitempty"`
	ToolCalls  []groqToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	Name       string         `json:"name,omitempty"`
}

type groqToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function groqFunctionCall `json:"function"`
}

type groqFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type groqTool struct {
	Type     string           `json:"type"`
	Function groqToolFunction `json:"function"`
}

type groqToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type groqChatRequest struct {
	Model      string        `json:"model"`
	Messages   []groqMessage `json:"messages"`
	Tools      []groqTool    `json:"tools,omitempty"`
	ToolChoice any           `json:"tool_choice,omitempty"`
	MaxTokens  int           `json:"max_tokens,omitempty"`
	Temperature float64      `json:"temperature,omitempty"`
}

type groqChoice struct {
	Message      groqMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type groqChatResponse struct {
	Choices []groqChoice `json:"choices"`
}

// ToolCallResult holds the tool chosen by the LLM and its raw JSON arguments.
type ToolCallResult struct {
	ToolName  string
	Arguments json.RawMessage
}

// Completion sends a user prompt to Groq and returns the assistant's reply.
func (c *GroqClient) Completion(ctx context.Context, prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("groq API key not configured")
	}
	reqBody := groqChatRequest{
		Model:    c.model,
		Messages: []groqMessage{{Role: "user", Content: prompt}},
	}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var body bytes.Buffer
		body.ReadFrom(resp.Body)
		return "", fmt.Errorf("groq request failed: %s – %s", resp.Status, body.String())
	}
	var payload groqChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if len(payload.Choices) == 0 {
		return "", fmt.Errorf("groq response missing choices")
	}
	return payload.Choices[0].Message.Content, nil
}

// PurchaseToolCall runs one agentic turn: the LLM is given the system prompt,
// the user prompt, and a purchase tool schema. It returns the tool call the LLM
// decided to make, or (nil, textReply, nil) if it chose not to call a tool.
func (c *GroqClient) PurchaseToolCall(ctx context.Context, systemPrompt, userPrompt string) (*ToolCallResult, string, error) {
	if c.apiKey == "" {
		return nil, "", fmt.Errorf("groq API key not configured")
	}

	tools := []groqTool{
		{
			Type: "function",
			Function: groqToolFunction{
				Name:        "purchase",
				Description: "Execute a purchase through the Aegis policy gateway. Use this to buy a product for the user.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"sku": map[string]any{
							"type":        "string",
							"description": "The exact offer SKU of the item to purchase",
						},
						"quantity": map[string]any{
							"type":        "integer",
							"description": "Number of units to buy (must be a positive integer)",
						},
						"amount_paisa": map[string]any{
							"type":        "integer",
							"description": "The exact total price to pay in paisa, if the user explicitly specifies a tampered or custom price",
						},
					},
					"required": []string{"sku", "quantity"},
				},
			},
		},
	}

	messages := []groqMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	reqBody := groqChatRequest{
		Model:      c.model,
		Messages:   messages,
		Tools:      tools,
		ToolChoice: "required",
		MaxTokens:  500,
		Temperature: 0.1,
	}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var body bytes.Buffer
		body.ReadFrom(resp.Body)
		return nil, "", fmt.Errorf("groq request failed: %s – %s", resp.Status, body.String())
	}

	var payload groqChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, "", err
	}
	if len(payload.Choices) == 0 {
		return nil, "", fmt.Errorf("groq response missing choices")
	}

	choice := payload.Choices[0]

	// Model called a tool — return the call.
	if len(choice.Message.ToolCalls) > 0 {
		tc := choice.Message.ToolCalls[0]
		return &ToolCallResult{
			ToolName:  tc.Function.Name,
			Arguments: json.RawMessage(tc.Function.Arguments),
		}, "", nil
	}

	// Some models (e.g., gpt-oss) return tool calls as a function call to "json"
	// with a tool_calls array in arguments. Try to parse that.
	if choice.Message.Content != "" {
		var parsed struct {
			Name      string `json:"name"`
			Arguments struct {
				ToolCalls []struct {
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"arguments"`
		}
		if err := json.Unmarshal([]byte(choice.Message.Content), &parsed); err == nil {
			if (parsed.Name == "json" || parsed.Name == "commentary") && len(parsed.Arguments.ToolCalls) > 0 {
				tc := parsed.Arguments.ToolCalls[0]
				if tc.Function.Name == "purchase" {
					return &ToolCallResult{
						ToolName:  tc.Function.Name,
						Arguments: json.RawMessage(tc.Function.Arguments),
					}, "", nil
				}
			}
		}
	}

	// Model replied with text (refused, asked clarifying question, etc.).
	return nil, choice.Message.Content, nil
}