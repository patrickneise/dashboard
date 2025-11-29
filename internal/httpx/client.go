package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type Client struct {
	HTTP      *http.Client
	UserAgent string

	Retries int
	Backoff time.Duration
}

func New(userAgent string) *Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &Client{
		HTTP: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
		UserAgent: userAgent,
		Retries:   2,
		Backoff:   250 * time.Millisecond,
	}
}

func (c *Client) GetJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	req.Header.Set("Accept", "application/json")

	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		if attempt > 0 {
			time.Sleep(c.Backoff * time.Duration(attempt))
		}

		resp, err := c.HTTP.Do(req)
		if err != nil {
			lastErr = err
			if isTransient(err) && attempt < c.Retries {
				continue
			}
			return err
		}
		// Always close body
		body := resp.Body
		defer body.Close()

		if resp.StatusCode >= 500 && attempt < c.Retries {
			// drain body to allow connection reuse
			_, _ = io.Copy(io.Discard, body)
			lastErr = fmt.Errorf("server error %d", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 400 {
			b, _ := io.ReadAll(io.LimitReader(body, 4<<10))
			return fmt.Errorf("http %d: %s", resp.StatusCode, string(b))
		}

		return json.NewDecoder(body).Decode(out)
	}

	return lastErr
}

func isTransient(err error) bool {
	// net.Error covers timeouts and temporary erros for many transports
	var ne net.Error
	if errors.As(err, &ne) {
		return ne.Timeout()
	}
	return false
}
