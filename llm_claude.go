package main

import (
	"fmt"
	"strings"

	"github.com/madebywelch/anthropic-go/pkg/anthropic"
)

var MESSAGE_QUEUE_CLAUDE map[string]string

func CompletionWithoutSessionByClaude(client *anthropic.Client, prompt string) (string, error) {
	resp, err := client.Complete(&anthropic.CompletionRequest{
		Prompt:            fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", prompt),
		Model:             anthropic.ClaudeV1_3_100k,
		MaxTokensToSample: 1000000,
		StopSequences:     []string{"\r", "Human:"},
	}, nil)

	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return "", err
	}

	return resp.Completion, nil
}

func CompletionWithoutSessionWithStreamByClaude(client *anthropic.Client, prompt string, callBack anthropic.StreamCallback) error {
	_, err := client.Complete(&anthropic.CompletionRequest{
		Prompt:            fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", prompt),
		Model:             anthropic.ClaudeV1_3_100k,
		MaxTokensToSample: 1000000,
		StopSequences:     []string{"\r", "Human:"},
		Stream:            true,
	}, callBack)
	if err != nil {
		fmt.Printf("CompletionStream error: %v\n", err)
		return err
	}
	return nil
}

func CompletionWithSessionByClaude(client *anthropic.Client, conversationID string, prompt string) (string, error) {
	messages := MESSAGE_QUEUE_CLAUDE[conversationID]
	resp, err := client.Complete(&anthropic.CompletionRequest{
		Prompt:            messages + fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", prompt),
		Model:             anthropic.ClaudeV1_3_100k,
		MaxTokensToSample: 1000000,
		StopSequences:     []string{"\r", "Human:"},
	}, nil)

	if err != nil {
		fmt.Printf("Completion error: %v\n", err)
		return "", err
	}

	AddMessageToClaude(conversationID, fmt.Sprintf("\n\nHuman: %s\n\nAssistant: %s", prompt, resp.Completion))
	return resp.Completion, nil
}

func CompletionWithSessionWithStreamByClaude(client *anthropic.Client, conversationID string, prompt string, callBack anthropic.StreamCallback) error {
	messages := MESSAGE_QUEUE_CLAUDE[conversationID]
	_, err := client.Complete(&anthropic.CompletionRequest{
		Prompt:            messages + fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", prompt),
		Model:             anthropic.ClaudeV1_3_100k,
		MaxTokensToSample: 1000000,
		StopSequences:     []string{"\r", "Human:"},
		Stream:            true,
	}, callBack)
	if err != nil {
		fmt.Printf("CompletionStream error: %v\n", err)
		return err
	}
	return nil
}

func AddMessageToClaude(conversationID string, message string) string {
	if MESSAGE_QUEUE_CLAUDE == nil {
		MESSAGE_QUEUE_CLAUDE = make(map[string]string)
	}
	if _, ok := MESSAGE_QUEUE_CLAUDE[conversationID]; !ok {
		MESSAGE_QUEUE_CLAUDE[conversationID] = ""
	}
	splitValue := strings.Split(MESSAGE_QUEUE_CLAUDE[conversationID], "\\n\\nHuman: ")
	if len(splitValue) > 10 {
		newValue := strings.Join(splitValue[1:], "\\n\\nHuman: ")
		MESSAGE_QUEUE_CLAUDE[conversationID] = newValue
	}
	MESSAGE_QUEUE_CLAUDE[conversationID] += message
	return MESSAGE_QUEUE_CLAUDE[conversationID]
}

func GetClaudeClient() (*anthropic.Client, error) {
	return anthropic.NewClient(AnthropicApiKey)
}

func main___() {
	Init()
	c, err := GetClaudeClient()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	resp1, err := CompletionWithoutSessionByClaude(c, "Hello, I am Claude. I am a chatbot. Who are you")
	if err != nil {
		return
	}
	fmt.Printf("Response: %v\n", resp1)
	prompt := ""
	respp := ""
	var callback anthropic.StreamCallback = func(resp *anthropic.CompletionResponse) error {
		fmt.Printf("Completion: %v\n", resp.Completion)
		respp = resp.Completion
		return nil
	}
	err = CompletionWithoutSessionWithStreamByClaude(c, "Hello, I am Claude. I am a chatbot. Who are you", callback)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	resp2, _ := CompletionWithSessionByClaude(c, "1", "a=1, b=2, c=a+b, c=?")
	fmt.Printf("resp2: %s", resp2)
	resp3, _ := CompletionWithSessionByClaude(c, "1", "a=?")
	fmt.Printf("resp3: %s", resp3)

	prompt = "a=1, b=2, c=a+b, c=?"
	_ = CompletionWithSessionWithStreamByClaude(c, "2", prompt, callback)
	AddMessageToClaude("2", fmt.Sprintf("\n\nHuman: %s\n\nAssistant: %s", prompt, respp))
	prompt = "a=?"
	_ = CompletionWithSessionWithStreamByClaude(c, "2", prompt, callback)
	AddMessageToClaude("2", fmt.Sprintf("\n\nHuman: %s\n\nAssistant: %s", prompt, respp))

	fmt.Println("Done")
}
