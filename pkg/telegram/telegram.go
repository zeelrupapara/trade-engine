package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gitlab.com/zeelrupapara/trade-engine/config"
)

// Client handles Telegram messaging
type Client struct {
	BotToken string
	ChatID   string
	Client   *http.Client
}

// NewClient initializes Telegram client from env vars

func NewClient(config *config.AppConfig) *Client {
	return &Client{
		BotToken: config.Telegram.BotToken,
		ChatID:   config.Telegram.ChatID,
		Client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// Send sends message text to the configured chat
func (c *Client) Send(text string) error {
	fmt.Println("Sending Telegram message:", text)
	fmt.Println("Chat ID:", c.ChatID)
	fmt.Println("Bot Token:", c.BotToken)
	url := "https://api.telegram.org/bot" + c.BotToken + "/sendMessage"
	payload := map[string]string{
		"chat_id": c.ChatID,
		"text":    text,
	}
	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Telegram Response:", string(body)) // Add this for debugging

	return nil
}
