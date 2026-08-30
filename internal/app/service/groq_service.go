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
        model = "gpt-4o-mini"
    }
    return &GroqClient{
        apiKey: apiKey,
        model:  model,
        client: &http.Client{Timeout: 30 * time.Second},
    }
}

// request/response payloads – match Groq's OpenAI‑compatible chat API.
type groqMessage struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type groqChatRequest struct {
    Model    string        `json:"model"`
    Messages []groqMessage `json:"messages"`
}

type groqChoice struct {
    Message groqMessage `json:"message"`
}

type groqChatResponse struct {
    Choices []groqChoice `json:"choices"`
}

// Completion sends a user prompt to Groq and returns the assistant's reply.
func (c *GroqClient) Completion(ctx context.Context, prompt string) (string, error) {
    if c.apiKey == "" {
        return "", fmt.Errorf("groq API key not configured")
    }
    reqBody := groqChatRequest{
        Model: c.model,
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
