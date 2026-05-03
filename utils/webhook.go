package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var client = &http.Client{Timeout: 10 * time.Second}

func SendWebhook(ctx context.Context, url string, data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling webhook body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	// Set content type to JSON
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		r := io.LimitReader(resp.Body, 1024*1024) // limit to 1MiB
		body, _ := io.ReadAll(r)
		return fmt.Errorf("webhook returned status code %d with data %q", resp.StatusCode, string(body))
	}

	return nil
}
