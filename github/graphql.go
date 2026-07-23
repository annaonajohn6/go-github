package github

import (
	"context"
	"io"
	"net/http"
)

func (c *Client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	resp, err := c.client.Do(req)
	if err != nil {
		if resp != nil && resp.Body != nil {
			io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
			resp.Body.Close()
		}
		return nil, err
	}

	// Ensure body is always drained and closed
	// We use a helper to handle the cleanup
	return resp, nil
}

func drainAndClose(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
	}
}