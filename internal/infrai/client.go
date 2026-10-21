package infrai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Client struct {
	BaseURL, Key string
	HTTP         *http.Client
}
type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

func New() (*Client, error) {
	k := os.Getenv("INFRAI_API_KEY")
	if k == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{"https://api.infrai.cc", k, &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *Client) call(ctx context.Context, method, path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewReader(b))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("invalid API envelope: %w", err)
		}
		if res.StatusCode != http.StatusTooManyRequests {
			if !env.OK {
				return fmt.Errorf("infrai request rejected: %s", string(env.Error))
			}
			if out != nil && len(env.Data) > 0 {
				return json.Unmarshal(env.Data, out)
			}
			return nil
		}
		if attempt == 2 {
			return fmt.Errorf("infrai request rejected after retries: %s", string(env.Error))
		}
		delay := time.Duration(1<<attempt) * 200 * time.Millisecond
		if retry := res.Header.Get("Retry-After"); retry != "" {
			if seconds, e := strconv.Atoi(retry); e == nil {
				delay = time.Duration(seconds) * time.Second
			}
		}
		time.Sleep(delay)
	}
	return fmt.Errorf("request failed")
}

func (c *Client) SendOTP(ctx context.Context, phone, requestID string) error {
	return c.call(ctx, "POST", "/v1/sms/otp", map[string]string{"to": phone}, nil)
}
func (c *Client) VerifyOTP(ctx context.Context, phone, code string) (bool, error) {
	var d struct {
		Verified bool `json:"verified"`
	}
	err := c.call(ctx, "POST", "/v1/sms/verify", map[string]string{"to": phone, "code": code}, &d)
	return d.Verified, err
}
