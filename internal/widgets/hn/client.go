package hn

import (
	"context"
	"fmt"

	"github.com/patrickneise/dashboard/internal/httpx"
)

const baseURL = "https://hacker-news.firebaseio.com/v0"

type Client struct {
	http *httpx.Client
}

func NewClient(h *httpx.Client) *Client {
	if h == nil {
		h = httpx.New("dashboard/0.1")
	}
	return &Client{http: h}
}

// TopStories returns a list of item IDs in rank order.
func (c *Client) TopStories(ctx context.Context) ([]int64, error) {
	var ids []int64
	if err := c.http.GetJSON(ctx, fmt.Sprintf("%s/topstories.json", baseURL), &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func (c *Client) Item(ctx context.Context, id int64) (*Item, error) {
	var it Item
	if err := c.http.GetJSON(ctx, fmt.Sprintf("%s/item/%d.json", baseURL, id), &it); err != nil {
		return nil, err
	}
	return &it, nil
}
