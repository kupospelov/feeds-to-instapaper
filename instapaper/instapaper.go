package instapaper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Instapaper struct {
	username string
	password string
	client   *http.Client
}

func New(username, password string) *Instapaper {
	return &Instapaper{
		username: username,
		password: password,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (api *Instapaper) Add(link, title string) error {
	apiURL := "https://www.instapaper.com/api/add"
	formData := fmt.Sprintf("username=%s&password=%s&url=%s",
		url.QueryEscape(api.username),
		url.QueryEscape(api.password),
		url.QueryEscape(link))
	if title != "" {
		formData += "&title=" + url.QueryEscape(title)
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(formData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := api.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Instapaper API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
