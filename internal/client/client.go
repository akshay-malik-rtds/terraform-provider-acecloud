package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// Pre-compiled regex used to collapse runs of whitespace produced by the
// JSON-error flattener.
var spaceRe = regexp.MustCompile(`\s{2,}`)

// sanitizeErrorMessage prepares an API error message for display. It flattens
// JSON validation objects (e.g. {"field":["msg1","msg2"]}) into a readable
// "field: msg1; field: msg2" string and collapses excess whitespace. The
// npc-api backend already returns user-safe error text, so no internal-tech
// filtering is applied here.
func sanitizeErrorMessage(msg string) string {
	if msg == "" {
		return msg
	}

	// If the message looks like a raw JSON validation object, flatten it.
	if strings.HasPrefix(strings.TrimSpace(msg), "{") {
		if flattened := flattenJSONError(msg); flattened != "" {
			msg = flattened
		}
	}

	return strings.TrimSpace(spaceRe.ReplaceAllString(msg, " "))
}

// flattenJSONError attempts to parse a JSON validation error object
// (e.g. {"tags":["tags must contain one of: ALB or NLB"]}) into a
// human-readable "field: message" string. Returns "" if not valid JSON.
func flattenJSONError(msg string) string {
	trimmed := strings.TrimSpace(msg)
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &obj); err != nil {
		return ""
	}

	var parts []string
	for field, val := range obj {
		// Try as string array first (most common validation format)
		var arr []string
		if err := json.Unmarshal(val, &arr); err == nil {
			for _, a := range arr {
				parts = append(parts, fmt.Sprintf("%s: %s", field, a))
			}
			continue
		}
		// Try as plain string
		var s string
		if err := json.Unmarshal(val, &s); err == nil {
			parts = append(parts, fmt.Sprintf("%s: %s", field, s))
			continue
		}
		// Fallback: raw value
		parts = append(parts, fmt.Sprintf("%s: %s", field, string(val)))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "; ")
}

// Client is the HTTP client for the Ace Cloud (npc-api) backend.
//
// Authentication: when APIKeyID and APIKeySecret are both set, the client
// sends an X-Ace-Api-Key header (long-lived API key auth, preferred for
// automation). Otherwise it falls back to APIToken via the standard
// Authorization: Bearer header.
type Client struct {
	BaseURL           string
	APIToken          string
	APIKeyID          string
	APIKeySecret      string
	APIKeyServiceName string
	Region            string
	ProjectID         string
	HTTPClient        *http.Client
}

// sensitiveFieldRe matches JSON fields that should be redacted in debug logs.
var sensitiveFieldRe = regexp.MustCompile(`(?i)"(password|admin_password|accessToken|access_token|api_token|secret|credentials?)":\s*"[^"]*"`)

// redactSensitiveFields masks sensitive values in JSON strings for safe logging.
func redactSensitiveFields(s string) string {
	return sensitiveFieldRe.ReplaceAllStringFunc(s, func(match string) string {
		idx := strings.Index(match, ":")
		if idx < 0 {
			return match
		}
		return match[:idx+1] + ` "**REDACTED**"`
	})
}

// APIResponse is the standard npc-api response envelope.
// Note: Message can be a string or an object (for validation errors), so we
// use json.RawMessage and extract a string afterwards.
// Some endpoints return "messages" (plural) instead of "message".
// The `error` field is heterogeneous across endpoints — it can be a boolean
// (custom envelope), a string error class name like "Bad Request" (NestJS
// default), or absent — so we keep it raw and interpret it in IsError.
type APIResponse struct {
	RawError    json.RawMessage `json:"error,omitempty"`
	IsErrorBool bool            `json:"-"`
	ErrorString string          `json:"-"`
	Success     *bool           `json:"success,omitempty"` // Some endpoints use {success: bool} instead of {error: bool}
	RawMessage  json.RawMessage `json:"message"`
	RawMessages json.RawMessage `json:"messages"`
	Message     string          `json:"-"`
	RawStatus   json.RawMessage `json:"status,omitempty"`
	Status      int             `json:"-"`
	Data        json.RawMessage `json:"data"`
	RawBody     json.RawMessage `json:"-"` // Full response body for non-standard envelopes (bare arrays, etc.)
}

// parseError interprets the heterogeneous `error` field. Boolean true or any
// non-empty string class name signals an error.
func (r *APIResponse) parseError() {
	if len(r.RawError) == 0 {
		return
	}
	var b bool
	if err := json.Unmarshal(r.RawError, &b); err == nil {
		r.IsErrorBool = b
		return
	}
	var s string
	if err := json.Unmarshal(r.RawError, &s); err == nil {
		r.ErrorString = s
		r.IsErrorBool = s != ""
	}
}

// IsError returns true if the API response indicates an error.
// Handles both {error: true|"<class>"} and {success: false} envelopes.
func (r *APIResponse) IsError() bool {
	if r.IsErrorBool {
		return true
	}
	if r.Success != nil && !*r.Success {
		return true
	}
	return false
}

// parseStatus extracts an integer status code from the raw status field.
// Some endpoints return status as an int (e.g., 400), others as a string
// (e.g., "Cluster creation in process"). Non-integer values are ignored.
func (r *APIResponse) parseStatus() {
	if r.RawStatus == nil {
		return
	}
	var n int
	if err := json.Unmarshal(r.RawStatus, &n); err == nil {
		r.Status = n
	}
	// If it's a string, Status stays 0 — that's fine.
}

// parseMessage extracts a string message from the raw message field.
// It handles both "message" (string or object) and "messages" (validation object).
func (r *APIResponse) parseMessage() {
	// Try "message" field first
	if r.RawMessage != nil {
		var s string
		if err := json.Unmarshal(r.RawMessage, &s); err == nil {
			r.Message = s
			return
		}
		// Fallback: use the raw JSON as the message string
		r.Message = string(r.RawMessage)
		return
	}

	// Try "messages" field (used by some validation endpoints)
	if r.RawMessages != nil {
		r.Message = flattenValidationMessages(r.RawMessages)
	}
}

// flattenValidationMessages converts a nested validation messages object
// like {"command":["must be an array"],"networking":{"xForwardedFor":["must be boolean"]}}
// into a flat, human-readable string.
func flattenValidationMessages(raw json.RawMessage) string {
	var msgs map[string]json.RawMessage
	if err := json.Unmarshal(raw, &msgs); err != nil {
		return string(raw)
	}

	var parts []string
	for field, val := range msgs {
		// Try as string array first
		var arr []string
		if err := json.Unmarshal(val, &arr); err == nil {
			for _, a := range arr {
				parts = append(parts, fmt.Sprintf("%s: %s", field, a))
			}
			continue
		}
		// Try as nested object (one level deep)
		var nested map[string]json.RawMessage
		if err := json.Unmarshal(val, &nested); err == nil {
			for subField, subVal := range nested {
				var subArr []string
				if err := json.Unmarshal(subVal, &subArr); err == nil {
					for _, a := range subArr {
						parts = append(parts, fmt.Sprintf("%s.%s: %s", field, subField, a))
					}
				} else {
					parts = append(parts, fmt.Sprintf("%s.%s: %s", field, subField, string(subVal)))
				}
			}
			continue
		}
		// Fallback
		parts = append(parts, fmt.Sprintf("%s: %s", field, string(val)))
	}
	if len(parts) == 0 {
		return string(raw)
	}
	return strings.Join(parts, "; ")
}

// NewClient creates a new Ace Cloud API client using JWT bearer token auth.
// Warns if baseURL uses HTTP instead of HTTPS.
func NewClient(baseURL, apiToken, region, projectID string) *Client {
	if strings.HasPrefix(baseURL, "http://") {
		fmt.Fprintf(os.Stderr, "[WARN] AceCloud API URL uses HTTP (not HTTPS). Credentials will be sent unencrypted. Use https:// for production.\n")
	}
	return &Client{
		BaseURL:   baseURL,
		APIToken:  apiToken,
		Region:    region,
		ProjectID: projectID,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// NewClientWithAPIKey creates a new Ace Cloud API client using API key auth.
// API keys are long-lived credentials suitable for automation (CI, Terraform).
// serviceName is optional metadata for audit logs; pass "" to skip.
func NewClientWithAPIKey(baseURL, apiKeyID, apiKeySecret, serviceName, region, projectID string) *Client {
	if strings.HasPrefix(baseURL, "http://") {
		fmt.Fprintf(os.Stderr, "[WARN] AceCloud API URL uses HTTP (not HTTPS). Credentials will be sent unencrypted. Use https:// for production.\n")
	}
	return &Client{
		BaseURL:           baseURL,
		APIKeyID:          apiKeyID,
		APIKeySecret:      apiKeySecret,
		APIKeyServiceName: serviceName,
		Region:            region,
		ProjectID:         projectID,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// hasAPIKey returns true if the client has API key credentials set.
func (c *Client) hasAPIKey() bool {
	return c.APIKeyID != "" && c.APIKeySecret != ""
}

// setAuthHeader applies the appropriate auth header to req based on
// configured credentials. Prefers API key over JWT when both are set.
func (c *Client) setAuthHeader(req *http.Request) {
	if c.hasAPIKey() {
		req.Header.Set("X-Ace-Api-Key", c.APIKeyID+"."+c.APIKeySecret)
		if c.APIKeyServiceName != "" {
			req.Header.Set("X-Api-Key-Service-Name", c.APIKeyServiceName)
		}
		return
	}
	if c.APIToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.APIToken))
	}
}

// buildURL constructs the full URL with region and project_id query params.
func (c *Client) buildURL(path string, extraParams map[string]string) string {
	u, _ := url.Parse(fmt.Sprintf("%s%s", c.BaseURL, path))
	q := u.Query()
	q.Set("region", c.Region)
	q.Set("project_id", c.ProjectID)
	for k, v := range extraParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// maxRetries is the number of times to retry on transient errors (e.g. 400/500 from overloaded backend).
const maxRetries = 3

// retryableStatus returns true if the HTTP status code suggests a transient error worth retrying.
func retryableStatus(statusCode int) bool {
	return statusCode == http.StatusBadRequest || // 400 — backend often returns this under load
		statusCode == http.StatusBadGateway || // 502
		statusCode == http.StatusServiceUnavailable || // 503
		statusCode == http.StatusGatewayTimeout // 504
}

// DoRequest executes an HTTP request against the npc-api and returns the parsed response.
// It retries transient errors up to maxRetries times with exponential backoff.
func (c *Client) DoRequest(ctx context.Context, method, path string, body interface{}, extraParams map[string]string) (*APIResponse, error) {
	fullURL := c.buildURL(path, extraParams)

	var jsonBody []byte
	if body != nil {
		var err error
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s...
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		var reqBody io.Reader
		if jsonBody != nil {
			reqBody = bytes.NewReader(jsonBody)
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		if jsonBody != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Header.Set("Accept", "application/json")
		c.setAuthHeader(req)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		// On 401, don't retry — it's a permanent auth failure.
		if resp.StatusCode == http.StatusUnauthorized {
			var apiResp APIResponse
			if err := json.Unmarshal(respBody, &apiResp); err == nil {
				apiResp.parseMessage()
				apiResp.parseError()
				return nil, fmt.Errorf("authentication failed (401): %s", sanitizeErrorMessage(apiResp.Message))
			}
			return nil, fmt.Errorf("authentication failed (401)")
		}

		// DEBUG: log request and response with sensitive field redaction
		if os.Getenv("ACECLOUD_DEBUG") != "" {
			debugPath := os.Getenv("ACECLOUD_DEBUG_FILE")
			if debugPath == "" {
				debugPath = "/tmp/acecloud_debug.log"
			}
			f, _ := os.OpenFile(debugPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if f != nil {
				safeReqBody := redactSensitiveFields(string(jsonBody))
				safeRespBody := redactSensitiveFields(string(respBody))
				_, _ = fmt.Fprintf(f, "--- Attempt %d ---\nMethod: %s\nURL: %s\nReqBody: %s\nHTTP Status: %d\nRespBody: %s\n\n", attempt+1, method, fullURL, safeReqBody, resp.StatusCode, safeRespBody)
				_ = f.Close()
			}
		}

		var apiResp APIResponse
		apiResp.RawBody = respBody // Store full body for non-standard envelopes
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			// Some endpoints return bare arrays instead of {error, data} envelope.
			// If it looks like valid JSON, treat the raw body as the data.
			trimmed := bytes.TrimSpace(respBody)
			if len(trimmed) > 0 && (trimmed[0] == '[' || trimmed[0] == '{') && resp.StatusCode >= 200 && resp.StatusCode < 300 {
				apiResp.Data = trimmed
				// fall through to success path
			} else {
				// If we can't parse the response and status is retryable, retry.
				if retryableStatus(resp.StatusCode) {
					lastErr = fmt.Errorf("failed to parse API response (status %d)", resp.StatusCode)
					continue
				}
				return nil, fmt.Errorf("failed to parse API response (status %d)", resp.StatusCode)
			}
		}
		apiResp.parseStatus()
		apiResp.parseMessage()
		apiResp.parseError()

		// If the API returned an error with a retryable status code, retry.
		if apiResp.IsError() && retryableStatus(resp.StatusCode) {
			lastErr = fmt.Errorf("API error (status %d): %s", resp.StatusCode, sanitizeErrorMessage(apiResp.Message))
			continue
		}

		if apiResp.IsError() {
			return &apiResp, fmt.Errorf("API error: %s", sanitizeErrorMessage(apiResp.Message))
		}

		return &apiResp, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, extraParams map[string]string) (*APIResponse, error) {
	return c.DoRequest(ctx, http.MethodGet, path, nil, extraParams)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*APIResponse, error) {
	return c.DoRequest(ctx, http.MethodPost, path, body, nil)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, body interface{}) (*APIResponse, error) {
	return c.DoRequest(ctx, http.MethodPut, path, body, nil)
}

// Patch performs a PATCH request.
func (c *Client) Patch(ctx context.Context, path string, body interface{}) (*APIResponse, error) {
	return c.DoRequest(ctx, http.MethodPatch, path, body, nil)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string, body interface{}) (*APIResponse, error) {
	return c.DoRequest(ctx, http.MethodDelete, path, body, nil)
}

// DeleteWithParams performs a DELETE request with extra query parameters.
// Used for cascade deletes (e.g. load balancers) and action-based deletes.
func (c *Client) DeleteWithParams(ctx context.Context, path string, body interface{}, extraParams map[string]string) (*APIResponse, error) {
	return c.DoRequest(ctx, http.MethodDelete, path, body, extraParams)
}

// PutWithParams performs a PUT request with extra query parameters.
// Used for action-based updates (e.g. floating IP associate/disassociate).
func (c *Client) PutWithParams(ctx context.Context, path string, body interface{}, extraParams map[string]string) (*APIResponse, error) {
	return c.DoRequest(ctx, http.MethodPut, path, body, extraParams)
}

// GetData is a helper that performs a GET request and unmarshals data into target.
func (c *Client) GetData(ctx context.Context, path string, target interface{}) error {
	resp, err := c.Get(ctx, path, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp.Data, target)
}

// PostData performs a POST and unmarshals the data field into target.
func (c *Client) PostData(ctx context.Context, path string, body interface{}, target interface{}) error {
	resp, err := c.Post(ctx, path, body)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp.Data, target)
}

// ValidateToken verifies the current token by calling GET /auth/me.
// This does NOT use buildURL because /auth/me is an auth endpoint that
// does not require region/project_id query params.
// Returns nil if the token is valid, or an error describing the problem.
func (c *Client) ValidateToken(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/auth/me", c.BaseURL), nil)
	if err != nil {
		return fmt.Errorf("failed to create token validation request: %w", err)
	}
	c.setAuthHeader(req)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("token validation request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("token is invalid or expired (HTTP 401)")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("token validation failed (HTTP %d)", resp.StatusCode)
	}

	return nil
}
