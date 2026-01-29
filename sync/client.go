package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles syncing notes to/from a remote server.
type Client struct {
	ServerURL  string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new sync client.
func NewClient(serverURL, apiKey string) *Client {
	if serverURL == "" {
		return nil
	}
	return &Client{
		ServerURL: serverURL,
		APIKey:    apiKey,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// notePayload is the JSON body for push/pull.
type notePayload struct {
	Date    string `json:"date"`
	Content string `json:"content"`
}

// PushNote uploads a day's note content to the server.
func (c *Client) PushNote(date time.Time, content string) error {
	if c == nil || c.ServerURL == "" {
		return nil
	}

	payload := notePayload{
		Date:    date.Format("2006-01-02"),
		Content: content,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	url := fmt.Sprintf("%s/api/notes/%s", c.ServerURL, date.Format("2006-01-02"))
	req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("sync error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// PullNote downloads a day's note content from the server.
func (c *Client) PullNote(date time.Time) (string, error) {
	if c == nil || c.ServerURL == "" {
		return "", nil
	}

	url := fmt.Sprintf("%s/api/notes/%s", c.ServerURL, date.Format("2006-01-02"))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("request error: %w", err)
	}

	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("sync error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "", nil
	}

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var payload notePayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode error: %w", err)
	}

	return payload.Content, nil
}

// PullAllDates fetches the list of available dates from the server.
func (c *Client) PullAllDates() ([]string, error) {
	if c == nil || c.ServerURL == "" {
		return nil, nil
	}

	url := fmt.Sprintf("%s/api/notes", c.ServerURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}

	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sync error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var dates []string
	if err := json.NewDecoder(resp.Body).Decode(&dates); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return dates, nil
}
