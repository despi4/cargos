// Package client implements authenticated service-to-service HTTP transport.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/depsi4/cargos/pkg/fault"
)

type Client struct {
	Key  string
	URLs map[string]string
	HTTP *http.Client
}

func New(key string, urls map[string]string) *Client {
	return &Client{Key: key, URLs: urls, HTTP: &http.Client{Timeout: 8 * time.Second}}
}
func (s *Client) Do(ctx context.Context, owner, role, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, s.URLs[role]+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("X-Internal-Key", s.Key)
	req.Header.Set("X-User-ID", owner)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return &fault.Unavailable{Message: role + " service unavailable"}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var e struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&e)
		if resp.StatusCode == 404 {
			return fault.ErrNotFound
		}
		if resp.StatusCode == 409 {
			return fault.ErrConflict
		}
		return &fault.Unavailable{Message: role + " service request failed"}
	}
	if out != nil {
		if err = json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(out); err != nil {
			return err
		}
	}
	return nil
}
func All[T any](ctx context.Context, s *Client, owner, role, path string) ([]T, error) {
	result := []T{}
	for offset := 0; ; offset += 500 {
		var page struct {
			Items []T `json:"items"`
		}
		if err := s.Do(ctx, owner, role, "GET", fmt.Sprintf("%s?limit=500&offset=%d", path, offset), nil, &page); err != nil {
			return nil, err
		}
		result = append(result, page.Items...)
		if len(page.Items) < 500 {
			return result, nil
		}
	}
}
