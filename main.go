package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	BotChannel          = "1107148758887710760"
	BotTestChannel      = "1108401368370249728"
	BotTestSubChannel   = "1108674815889522719"
	BotBardChannel      = "1108670964952211527"
	BotClaudeChannel    = "1108671068383744000"
	BotOpenaiTDFChannel = "1108671217688379403"
)

var (
	BaseUrl       string
	InitialPrompt string
	DiscordToken  string
	ApiKey        string
	PalmApiKey    string
)

func Init() {
	viper.SetConfigFile(".env")
	_ = viper.ReadInConfig()
	if err := viper.BindEnv("DISCORD_TOKEN"); err != nil {
		log.Fatal(err)
	}
	DiscordToken = viper.GetString("DISCORD_TOKEN")

	if err := viper.BindEnv("API_KEY"); err != nil {
		log.Fatal(err)
	}
	ApiKey = viper.GetString("API_KEY")

	if err := viper.BindEnv("BASE_URL"); err != nil {
		log.Fatal(err)
	}
	BaseUrl = viper.GetString("BASE_URL")

	if err := viper.BindEnv("PALM_API_KEY"); err != nil {
		log.Fatal(err)
	}
	PalmApiKey = viper.GetString("PALM_API_KEY")

	if err := viper.BindEnv("INITIAL_PROMPT"); err != nil {
		log.Fatal(err)
	}
	InitialPrompt = viper.GetString("INITIAL_PROMPT")
	if InitialPrompt == "" {
		InitialPrompt = "You are a professional assistant"
	}
}

func main() {
	Init()
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + DiscordToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	logger.Infof("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	_ = dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.ChannelID != BotChannel &&
		m.ChannelID != BotTestChannel &&
		m.ChannelID != BotBardChannel &&
		m.ChannelID != BotClaudeChannel &&
		m.ChannelID != BotOpenaiTDFChannel &&
		m.ChannelID != BotTestSubChannel {
		return
	}

	logger.Infof("Message Received: ID %s | Content %s | Author %s | Channel %s", m.ID, m.Content, m.Author, m.ChannelID)

	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" || m.Content == "Ping" {
		_, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
		if err != nil {
			return
		}
	}

	ctx, client, err := GetClient()
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Something went wrong with GetClient. Please try again later.")
		logger.Infof("Something went wrong with GetClient. Please try again later.")
		return
	}

	if m.ChannelID == BotChannel || m.ChannelID == BotTestChannel {
		ChatResponse(ctx, client, s, m)
	}

	if m.ChannelID == BotTestSubChannel || m.ChannelID == BotBardChannel {
		resp, err := CompletionWithSessionByPaLM(m.ChannelID, m.Content)
		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Something went wrong with CompletionWithSessionByPaLM. Please try again later.")
			logger.Infof("Something went wrong with CompletionWithSessionByPaLM. Please try again later.")
			return
		}
		_, _ = s.ChannelMessageSend(m.ChannelID, resp)
	}
}

func ChatResponse(ctx context.Context, client *openai.Client, s *discordgo.Session, m *discordgo.MessageCreate) {
	message, _ := s.ChannelMessageSend(m.ChannelID, "Typing...")

	resp, err := CompletionWithSessionWithStreamByOpenAI(ctx, client, m.ChannelID, m.Content)
	if err != nil {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Something went wrong with CompletionWithSessionWithStreamByOpenAI. Please try again later.")
		logger.Infof("Something went wrong with CompletionWithSessionWithStreamByOpenAI. Please try again later.")
		return
	}
	defer resp.Close()

	finalResp := ""
	count := 0
	for {
		response, err := resp.Recv()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			_, _ = s.ChannelMessageSend(m.ChannelID, "Something went wrong with Stream. Please try again later.")
			logger.Errorf("Stream error: %v", err)
			return
		}

		finalResp += response.Choices[0].Delta.Content
		if count%10 == 0 {
			_, err = s.ChannelMessageEdit(m.ChannelID, message.ID, finalResp+"\nTyping...")
			logger.Infof("streaming count: %d", count)
		}
		count++
		if err != nil {
			message, _ = s.ChannelMessageSend(m.ChannelID, "Something went wrong with Edit. Please try again later.")
			fmt.Printf("Edit error: %v\n", err)
			fmt.Println(m.ChannelID, message.ID, finalResp)
			return
		}
	}
	AddMessageToOpenAI(m.ChannelID, finalResp)
	logger.Infof("final response: %s", finalResp)
	logger.Infof("all count: %d", count)
	_, err = s.ChannelMessageEdit(m.ChannelID, message.ID, finalResp)
}
