package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sse_example/config"
)

// PipeConnectionManager stores input channel in case of /chat reconnect. It allows to continue data processing with a correct channel close
type PipeConnectionManager struct {
	InputCh     chan string
	InputClosed chan struct{}
	OutputCh    chan string
}

// Process takes data from input channel, send them to llm and write response in output channel
func Process(cm *PipeConnectionManager, httpClient http.Client, cfg config.Config) {
	defer close(cm.OutputCh)
	for {
		for v := range cm.InputCh {
			res, err := requestLLMChatAutocomplete(v, httpClient, cfg.OpenAIApiKey)
			if err != nil {
				log.Printf("LLM request failed: %v", err)
				res = "your request can't be processed. Try again later"
			}
			cm.OutputCh <- res
		}
	}
}

type Payload struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

var CHAT_AUTOCOMPLETE_URL = "https://api.openai.com/v1/chat/completions"

func requestLLMChatAutocomplete(message string, httpClient http.Client, apiKey string) (string, error) {
	p := Payload{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant",
			},
			{
				Role:    "user",
				Content: message,
			},
		},
	}
	jsonP, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("can't generate JSON payload for request: %w", err)
	}
	req, err := http.NewRequest("POST", CHAT_AUTOCOMPLETE_URL, bytes.NewReader(jsonP))
	if err != nil {
		return "", fmt.Errorf("failed to form http request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", apiKey))
	res, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making http request: %w", err)
	}
	var r Response
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&r); err != nil {
		return "", fmt.Errorf("can't decode server response: %w ", err)
	}
	return r.Choices[0].Message.Content, nil
}
