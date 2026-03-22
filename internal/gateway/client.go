package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Client is an HTTP+WebSocket client for the Hermes/OpenClaw Gateway.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	wsURL      string
}

// Session represents an active session in the gateway.
type Session struct {
	Key         string `json:"key"`
	Kind        string `json:"kind"`
	Channel     string `json:"channel,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	UpdatedAt   int64  `json:"updatedAt,omitempty"`
	SessionID   string `json:"sessionId,omitempty"`
	Model       string `json:"model,omitempty"`
	TotalTokens int    `json:"totalTokens,omitempty"`
	LastChannel string `json:"lastChannel,omitempty"`
	ParentKey   string `json:"parentKey,omitempty"`
}

// Message represents a single message in a session.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NewClient creates a Gateway client. It reads the bearer token from ~/.openclaw/openclaw.json.
func NewClient(baseURL string) (*Client, error) {
	token, err := readToken()
	if err != nil {
		return nil, fmt.Errorf("read gateway token: %w", err)
	}
	wsURL := "ws" + baseURL[len("http"):] + "/ws"
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		wsURL: wsURL,
	}, nil
}

// NewClientWithToken creates a Gateway client with an explicit token.
func NewClientWithToken(baseURL, token string) *Client {
	wsURL := "ws" + baseURL[len("http"):] + "/ws"
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		wsURL: wsURL,
	}
}

func readToken() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(home, ".openclaw", "openclaw.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read openclaw config: %w", err)
	}
	var cfg struct {
		Gateway struct {
			Auth struct {
				Token string `json:"token"`
			} `json:"auth"`
		} `json:"gateway"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parse openclaw config: %w", err)
	}
	if cfg.Gateway.Auth.Token == "" {
		return "", fmt.Errorf("no gateway token found in openclaw config")
	}
	return cfg.Gateway.Auth.Token, nil
}

func (c *Client) do(tool string, args map[string]interface{}) (json.RawMessage, error) {
	body := map[string]interface{}{"tool": tool, "args": args}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/tools/invoke", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gateway returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		OK    bool            `json:"ok"`
		Result json.RawMessage `json:"result,omitempty"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !result.OK {
		msg := "unknown error"
		if result.Error != nil {
			msg = result.Error.Message
		}
		return nil, fmt.Errorf("gateway error: %s", msg)
	}

	return result.Result, nil
}

// ListSessions returns all active sessions.
func (c *Client) ListSessions() ([]Session, error) {
	result, err := c.do("sessions_list", map[string]interface{}{"limit": 50, "messageLimit": 0})
	if err != nil {
		return nil, err
	}

	// Unwrap the content: {content: [{type:"text", text:"{count:N,sessions:[...]}"}]}
	var wrapper struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Details *struct {
			Count    int       `json:"count"`
			Sessions []Session `json:"sessions"`
		} `json:"details"`
	}
	if err := json.Unmarshal(result, &wrapper); err != nil {
		// Try direct parse
		var sessions []Session
		if err2 := json.Unmarshal(result, &sessions); err2 == nil {
			return sessions, nil
		}
		return nil, fmt.Errorf("parse sessions response: %w", err)
	}
	if wrapper.Details != nil {
		return wrapper.Details.Sessions, nil
	}
	// Try parsing inner text
	if len(wrapper.Content) > 0 && wrapper.Content[0].Text != "" {
		var inner struct {
			Count    int       `json:"count"`
			Sessions []Session `json:"sessions"`
		}
		if err := json.Unmarshal([]byte(wrapper.Content[0].Text), &inner); err == nil {
			return inner.Sessions, nil
		}
	}
	return []Session{}, nil
}

// GetSessionHistory returns the message history for a session.
func (c *Client) GetSessionHistory(sessionKey string, limit int) ([]Message, error) {
	result, err := c.do("sessions_history", map[string]interface{}{
		"sessionKey":   sessionKey,
		"limit":        limit,
		"includeTools": true,
	})
	if err != nil {
		return nil, err
	}

	// Unwrap content: {content: [{type:"text", text:"{messages:[...]}"}]}
	var wrapper struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(result, &wrapper); err == nil && len(wrapper.Content) > 0 {
		var inner struct {
			Messages []Message `json:"messages"`
		}
		if err := json.Unmarshal([]byte(wrapper.Content[0].Text), &inner); err == nil {
			return inner.Messages, nil
		}
	}

	// Try direct parse
	var msgs []Message
	if err := json.Unmarshal(result, &msgs); err == nil {
		return msgs, nil
	}
	var direct struct {
		Messages []Message `json:"messages"`
	}
	if err := json.Unmarshal(result, &direct); err == nil {
		return direct.Messages, nil
	}
	return []Message{}, nil
}

// SendMessage sends a message to a session. It returns a channel that signals
// completion. Note: this does not stream individual tokens. The channel will
// either close cleanly (indicating the gateway processed the message) or
// emit a single "[error] ..." string. After the channel closes, callers
// should use GetSessionHistory to retrieve the assistant's response.
func (c *Client) SendMessage(sessionKey, content string) (<-chan string, error) {
	idempotencyKey := uuid.New().String()
	ch := make(chan string, 64)

	// We do a non-blocking async call — the response comes via events
	// For HTTP mode we just fire and return; streaming handled by caller reading events
	go func() {
		defer close(ch)
		_, err := c.do("chat.send", map[string]interface{}{
			"sessionKey":      sessionKey,
			"message":         content,
			"thinking":        "adaptive",
			"idempotencyKey":  idempotencyKey,
		})
		if err != nil {
			// Send error as a special chunk
			ch <- "[error] " + err.Error()
		}
	}()

	return ch, nil
}

// InvokeTool calls an arbitrary gateway tool and returns the raw result.
func (c *Client) InvokeTool(tool string, args map[string]interface{}) (json.RawMessage, error) {
	return c.do(tool, args)
}

// BaseURL returns the gateway's base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Health checks if the gateway is reachable.
func (c *Client) Health() error {
	req, err := http.NewRequest("GET", c.baseURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("gateway unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("gateway returned %d", resp.StatusCode)
	}
	return nil
}
