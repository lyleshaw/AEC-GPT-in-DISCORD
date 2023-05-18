package main

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

var MESSAGE_QUEUE_OPENAI map[string][]openai.ChatCompletionMessage

func CompletionWithoutSessionByOpenAI(ctx context.Context, client *openai.Client, prompt string) (string, error) {
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func CompletionWithoutSessionWithStreamByOpenAI(ctx context.Context, client *openai.Client, prompt string) (*openai.ChatCompletionStream, error) {
	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 2048,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Stream: true,
	}
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return nil, err
	}
	return stream, nil
}

func CompletionWithSessionByOpenAI(ctx context.Context, client *openai.Client, conversationID string, prompt string) (string, error) {
	messages := AddMessageToOpenAI(conversationID, prompt)
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: messages,
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	AddMessageToOpenAI(conversationID, resp.Choices[0].Message.Content)
	return resp.Choices[0].Message.Content, nil
}

// CompletionWithSessionWithStreamByOpenAI should call AddMessageToOpenAI after using stream
func CompletionWithSessionWithStreamByOpenAI(ctx context.Context, client *openai.Client, conversationID string, prompt string) (*openai.ChatCompletionStream, error) {
	messages := AddMessageToOpenAI(conversationID, prompt)
	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 2048,
		Messages:  messages,
		Stream:    true,
	}
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		fmt.Printf("ChatCompletionStream error: %v\n", err)
		return nil, err
	}
	return stream, nil
}

func AddMessageToOpenAI(conversationID string, prompt string) []openai.ChatCompletionMessage {
	if MESSAGE_QUEUE_OPENAI == nil {
		MESSAGE_QUEUE_OPENAI = make(map[string][]openai.ChatCompletionMessage)
	}
	if _, ok := MESSAGE_QUEUE_OPENAI[conversationID]; !ok {
		MESSAGE_QUEUE_OPENAI[conversationID] = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: InitialPrompt,
			},
		}
	}
	if len(MESSAGE_QUEUE_OPENAI[conversationID]) > 4 {
		MESSAGE_QUEUE_OPENAI[conversationID] = MESSAGE_QUEUE_OPENAI[conversationID][1:]
	}

	MESSAGE_QUEUE_OPENAI[conversationID] = append(MESSAGE_QUEUE_OPENAI[conversationID], openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})
	return MESSAGE_QUEUE_OPENAI[conversationID]
}

func GetClient() (context.Context, *openai.Client, error) {
	ctx := context.Background()
	config := openai.DefaultConfig(ApiKey)
	config.BaseURL = BaseUrl
	return ctx, openai.NewClientWithConfig(config), nil
}
