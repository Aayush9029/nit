package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Aayush9029/nit/internal/config"
)

type Client struct {
	http        *http.Client
	bearerToken string
	baseURL     string
}

func NewClient(token string) *Client {
	return &Client{
		http:        &http.Client{Timeout: 15 * time.Second},
		bearerToken: token,
		baseURL:     config.APIBase,
	}
}

func (c *Client) get(ctx context.Context, endpoint string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+endpoint, nil)
	if err != nil {
		return nil, &RequestError{Message: fmt.Sprintf("failed to create request: %v", err)}
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &RequestError{Message: "network error, check your connection"}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &RequestError{Message: fmt.Sprintf("failed to read response: %v", err)}
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusTooManyRequests:
		resetStr := resp.Header.Get("x-rate-limit-reset")
		if resetStr != "" {
			if ts, err := strconv.ParseInt(resetStr, 10, 64); err == nil {
				return nil, &RateLimitError{ResetAt: time.Unix(ts, 0)}
			}
		}
		return nil, &RateLimitError{ResetAt: time.Now().Add(60 * time.Second)}
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, &AuthError{StatusCode: resp.StatusCode}
	default:
		msg := extractErrorMessage(body)
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return nil, &RequestError{Message: msg}
	}
}

func extractErrorMessage(body []byte) string {
	var resp struct {
		Errors []struct {
			Detail  string `json:"detail"`
			Message string `json:"message"`
		} `json:"errors"`
		Detail string `json:"detail"`
	}
	if json.Unmarshal(body, &resp) != nil {
		return ""
	}
	if len(resp.Errors) > 0 {
		if resp.Errors[0].Detail != "" {
			return resp.Errors[0].Detail
		}
		return resp.Errors[0].Message
	}
	return resp.Detail
}

func (c *Client) LookupUsers(ctx context.Context, usernames []string) (*UsersResponse, error) {
	params := url.Values{}
	params.Set("usernames", strings.Join(usernames, ","))
	params.Set("user.fields", config.UserFields)

	body, err := c.get(ctx, "/users/by?"+params.Encode())
	if err != nil {
		return nil, err
	}

	var resp UsersResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, &RequestError{Message: "failed to parse user response"}
	}
	return &resp, nil
}

func (c *Client) LookupUser(ctx context.Context, username string) (*UsersResponse, error) {
	// Use batch endpoint — single-user endpoint returns {data: object} not {data: [array]}
	return c.LookupUsers(ctx, []string{username})
}

func (c *Client) UserTimeline(ctx context.Context, userID string, opts TimelineOpts) (*TweetsResponse, error) {
	params := url.Values{}
	params.Set("tweet.fields", config.TweetFields)
	if opts.MaxResults > 0 {
		params.Set("max_results", strconv.Itoa(opts.MaxResults))
	}
	if opts.SinceID != "" {
		params.Set("since_id", opts.SinceID)
	}
	if opts.Exclude != "" {
		params.Set("exclude", opts.Exclude)
	}

	body, err := c.get(ctx, "/users/"+userID+"/tweets?"+params.Encode())
	if err != nil {
		return nil, err
	}

	var resp TweetsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, &RequestError{Message: "failed to parse tweets response"}
	}
	return &resp, nil
}
