package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client is a thin HTTP wrapper around the Flagr Management API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

func (c *Client) ListProjects(ctx context.Context) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, "/api/v1/projects", nil)
}

func (c *Client) ListEnvironments(ctx context.Context, projectID string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, "/api/v1/projects/"+projectID+"/environments", nil)
}

func (c *Client) ListFlags(ctx context.Context, projectID string) (json.RawMessage, error) {
	return c.do(ctx, http.MethodGet, "/api/v1/projects/"+projectID+"/flags", nil)
}

func (c *Client) GetFlagState(ctx context.Context, projectID, flagID, envID string) (json.RawMessage, error) {
	path := fmt.Sprintf("/api/v1/projects/%s/flags/%s/environments/%s", projectID, flagID, envID)
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) SetFlagState(ctx context.Context, projectID, flagID, envID, state string) (json.RawMessage, error) {
	path := fmt.Sprintf("/api/v1/projects/%s/flags/%s/environments/%s", projectID, flagID, envID)
	body := map[string]any{
		"state":        state,
		"enabled_list": []string{},
	}
	return c.do(ctx, http.MethodPatch, path, body)
}

func (c *Client) GetFlagHistory(ctx context.Context, projectID, flagID string) (json.RawMessage, error) {
	path := fmt.Sprintf("/api/v1/projects/%s/flags/%s/history", projectID, flagID)
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *Client) AddTenant(ctx context.Context, projectID, flagID, envID, tenantID string) (json.RawMessage, error) {
	path := fmt.Sprintf("/api/v1/projects/%s/flags/%s/environments/%s/tenants", projectID, flagID, envID)
	return c.do(ctx, http.MethodPost, path, map[string]string{"tenant_id": tenantID})
}

func (c *Client) RemoveTenant(ctx context.Context, projectID, flagID, envID, tenantID string) (json.RawMessage, error) {
	path := fmt.Sprintf("/api/v1/projects/%s/flags/%s/environments/%s/tenants/%s",
		projectID, flagID, envID, url.PathEscape(tenantID))
	return c.do(ctx, http.MethodDelete, path, nil)
}
