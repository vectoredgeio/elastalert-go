package alerts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SlackAlert struct {
	WebhookURL string
}

func (s *SlackAlert) Send(message string) error {
	payload := map[string]string{
		"text": message,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", s.WebhookURL, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
