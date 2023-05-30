package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bytedance/gopkg/util/logger"
	"net/http"
)

// Message represents a single message in a conversation.
type Message struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

// MessagePrompt contains the messages to use as prompt for the model.
type MessagePrompt struct {
	Messages []Message `json:"messages"`
}

// GenerateResponse contains the model-generated responses.
type GenerateResponse struct {
	Candidates []Message `json:"candidates"`
	Messages   []Message `json:"messages"`
}

// Request represents the request to generate a message.
type Request struct {
	Prompt         MessagePrompt `json:"prompt"`
	Temperature    float64       `json:"temperature"`
	CandidateCount int           `json:"candidateCount"`
}

// MESSAGE_QUEUE_PALM stores the conversation state.
var MESSAGE_QUEUE_PALM map[string][]Message

// CompletionWithSessionByPaLM generates a single response message using the PaLM API.
func CompletionWithSessionByPaLM(conversationID string, prompt string) (string, error) {
	// Build the request
	messages := AddMessageToPaLM(conversationID, prompt, "user")
	request := Request{
		Prompt:         messages,
		Temperature:    0.75,
		CandidateCount: 1,
	}
	reqBody, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	logger.Infof("reqBody: %v", string(reqBody))

	// Send request
	resp, err := http.Post("https://generativelanguage.googleapis.com/v1beta2/models/chat-bison-001:generateMessage?key="+PalmApiKey,
		"application/json",
		bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Errorf("resp: %v", resp)
		return "", err
	}

	// Parse response
	var response GenerateResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	respStr := response.Candidates[0].Content
	AddMessageToPaLM(conversationID, respStr, "bot")
	return respStr, nil
}

// AddMessageToPaLM adds a message to the chat session.
func AddMessageToPaLM(conversationID string, prompt string, author string) MessagePrompt {
	if MESSAGE_QUEUE_PALM == nil {
		MESSAGE_QUEUE_PALM = make(map[string][]Message)
	}

	messages := MESSAGE_QUEUE_PALM[conversationID]
	if len(messages) > 8 {
		messages = messages[1:]
	}

	messages = append(messages, Message{
		Author:  author,
		Content: prompt,
	})
	MESSAGE_QUEUE_PALM[conversationID] = messages

	resp := MessagePrompt{
		Messages: messages,
	}
	return resp
}

func main__() {
	Init()
	resp, err := CompletionWithSessionByPaLM("test", "Hello")
	if err != nil {
		return
	}
	fmt.Println(resp)
}
