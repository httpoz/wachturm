// Package notification provides functionality for sending notifications about package updates.
package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// TelegramNotifier sends notifications via Telegram.
type TelegramNotifier struct {
	botToken   string
	chatID     string
	httpClient *http.Client
	baseURL    string
}

// NewTelegramNotifier creates a new TelegramNotifier with the provided token and chat ID.
func NewTelegramNotifier(token, chatID string) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: token,
		chatID:   chatID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: "https://api.telegram.org",
	}
}

// DefaultTelegramNotifier creates a new TelegramNotifier using environment variables.
func DefaultTelegramNotifier() (*TelegramNotifier, error) {
	token := os.Getenv("WATCHTURM_TELEGRAM_BOT_KEY")
	if token == "" {
		return nil, fmt.Errorf("WATCHTURM_TELEGRAM_BOT_KEY environment variable not set")
	}

	chatID := os.Getenv("WATCHTURM_TELEGRAM_CHAT_ID")
	if chatID == "" {
		return nil, fmt.Errorf("WATCHTURM_TELEGRAM_CHAT_ID environment variable not set")
	}

	return NewTelegramNotifier(token, chatID), nil
}

// WithHTTPClient sets a custom HTTP client for the notifier.
func (n *TelegramNotifier) WithHTTPClient(client *http.Client) *TelegramNotifier {
	n.httpClient = client
	return n
}

// Send sends a notification with the provided message.
func (n *TelegramNotifier) Send(message string) error {
	url := fmt.Sprintf("%s/bot%s/sendMessage", n.baseURL, n.botToken)

	payload := map[string]string{
		"chat_id": n.chatID,
		"text":    message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := n.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send Telegram notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// SendSummary sends a notification with the contents of a summary file.
func (n *TelegramNotifier) SendSummary(summaryFilePath string) error {
	content, err := os.ReadFile(summaryFilePath)
	if err != nil {
		return fmt.Errorf("failed to read summary file: %w", err)
	}

	return n.Send(string(content))
}
