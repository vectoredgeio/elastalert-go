package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type GoogleChatAlert struct {
	WebhookURL string
}

// MessagePayload struct to format the message
type MessagePayload struct {
	Text string `json:"text"`
}

// Send method to send alerts to Google Chat
func (g *GoogleChatAlert) Send(message string) error {
	payload := MessagePayload{Text: message}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(g.WebhookURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return nil
}
