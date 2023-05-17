package main

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

var MESSAGE_QUEUE map[string][]openai.ChatCompletionMessage

func CompletionWithoutSession(ctx context.Context, client *openai.Client, prompt string) (string, error) {
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

func CompletionWithoutSessionWithStream(ctx context.Context, client *openai.Client, prompt string) (*openai.ChatCompletionStream, error) {
	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 20,
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

func CompletionWithSession(ctx context.Context, client *openai.Client, conversationID string, prompt string) (string, error) {
	messages := AddMessage(conversationID, prompt)
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

	AddMessage(conversationID, resp.Choices[0].Message.Content)
	return resp.Choices[0].Message.Content, nil
}

// CompletionWithSessionWithStream should call AddMessage after using stream
func CompletionWithSessionWithStream(ctx context.Context, client *openai.Client, conversationID string, prompt string) (*openai.ChatCompletionStream, error) {
	messages := AddMessage(conversationID, prompt)
	req := openai.ChatCompletionRequest{
		Model:     openai.GPT3Dot5Turbo,
		MaxTokens: 20,
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

func AddMessage(conversationID string, prompt string) []openai.ChatCompletionMessage {
	if MESSAGE_QUEUE == nil {
		MESSAGE_QUEUE = make(map[string][]openai.ChatCompletionMessage)
	}
	if _, ok := MESSAGE_QUEUE[conversationID]; !ok {
		MESSAGE_QUEUE[conversationID] = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: InitialPrompt,
			},
		}
	}
	MESSAGE_QUEUE[conversationID] = append(MESSAGE_QUEUE[conversationID], openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})
	return MESSAGE_QUEUE[conversationID]
}

func GetClient() (context.Context, *openai.Client, error) {
	ctx := context.Background()
	config := openai.DefaultConfig(ApiKey)
	config.BaseURL = BaseUrl
	return ctx, openai.NewClientWithConfig(config), nil
}
