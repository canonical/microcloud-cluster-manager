package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type operation struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	Err        string `json:"err"`
}

// Operation reads an operation through the original management origin.
func (c *Client) Operation(ctx context.Context, cluster, id string) (*Response, error) {
	return c.Do(ctx, http.MethodGet, Path("remote-cluster", cluster, "operations", id), nil, nil, true)
}

// OperationID validates an asynchronous response without trusting its host or path as a request target.
func OperationID(r *Response) (string, error) {
	u, err := url.Parse(r.Operation)
	if err != nil {
		return "", fmt.Errorf("invalid operation URL")
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[len(parts)-2] != "operations" {
		return "", fmt.Errorf("invalid operation URL")
	}
	id := parts[len(parts)-1]
	if id == "" || id == "." || id == ".." || strings.Contains(id, "/") {
		return "", fmt.Errorf("invalid operation ID")
	}
	var op operation
	if err := json.Unmarshal(r.Metadata, &op); err != nil {
		return "", fmt.Errorf("decode operation: %w", err)
	}
	if op.ID != "" && op.ID != id {
		return "", fmt.Errorf("operation ID does not match response URL")
	}
	return id, nil
}

// WaitOperation polls once per second until success, failure, cancellation, or the caller's deadline.
func (c *Client) WaitOperation(ctx context.Context, cluster, id string) (*Response, error) {
	for {
		r, err := c.Operation(ctx, cluster, id)
		if err != nil {
			return nil, fmt.Errorf("operation %s on %s: %w", id, cluster, err)
		}
		var op operation
		if err := json.Unmarshal(r.Metadata, &op); err != nil {
			return nil, err
		}
		if op.StatusCode >= 200 {
			if op.StatusCode != 200 {
				return r, fmt.Errorf("operation %s on %s: %s: %s", id, cluster, op.Status, op.Err)
			}
			return r, nil
		}
		if op.StatusCode < 100 {
			return nil, fmt.Errorf("operation %s returned invalid status code", id)
		}
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, fmt.Errorf("waiting for operation %s on %s: %w; waiting does not cancel the remote operation", id, cluster, ctx.Err())
		case <-timer.C:
		}
	}
}
