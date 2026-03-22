package gateway

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// wsFrame is the request frame sent to the gateway.
type wsFrame struct {
	Type   string      `json:"type"`
	ID     string      `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// wsResponse is a response frame from the gateway.
type wsResponse struct {
	OK      bool            `json:"ok"`
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Error   *struct {
		Code    string `json:"code,omitempty"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// wsEvent is an event frame from the gateway.
type wsEvent struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Seq     *int            `json:"seq,omitempty"`
}

// chatEvent is the payload of a "chat" event.
type chatEvent struct {
	SessionKey   string          `json:"sessionKey"`
	RunID        string          `json:"runId"`
	State        string          `json:"state"` // "delta", "final", "error", "aborted"
	Message      json.RawMessage `json:"message,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
}

// WSConn wraps a WebSocket connection to the OpenClaw gateway.
type WSConn struct {
	conn    *websocket.Conn
	pending map[string]chan wsResponse
	mu      sync.Mutex
	events  chan wsEvent
	done    chan struct{}
}

// connectWS opens a WebSocket to the gateway, performs the challenge-response
// handshake, and returns a ready WSConn.
func connectWS(wsURL, token string) (*WSConn, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ws dial: %w", err)
	}

	wsc := &WSConn{
		conn:    conn,
		pending: make(map[string]chan wsResponse),
		events:  make(chan wsEvent, 128),
		done:    make(chan struct{}),
	}

	// Start reader goroutine
	go wsc.readLoop()

	// Wait for connect.challenge event
	var nonce string
	select {
	case evt := <-wsc.events:
		if evt.Event != "connect.challenge" {
			conn.Close()
			return nil, fmt.Errorf("expected connect.challenge, got %s", evt.Event)
		}
		var challenge struct {
			Nonce string `json:"nonce"`
		}
		if err := json.Unmarshal(evt.Payload, &challenge); err != nil {
			conn.Close()
			return nil, fmt.Errorf("parse challenge: %w", err)
		}
		nonce = challenge.Nonce
	case <-time.After(5 * time.Second):
		conn.Close()
		return nil, fmt.Errorf("timeout waiting for connect challenge")
	}

	// Send connect request
	connectParams := map[string]interface{}{
		"minProtocol": 3,
		"maxProtocol": 3,
		"client": map[string]interface{}{
			"id":          "hermes-tui",
			"displayName": "Hermes TUI",
			"version":     "1.0.0",
			"platform":    "darwin",
			"mode":        "ui",
			"instanceId":  uuid.New().String(),
		},
		"caps": []string{"tool.events"},
		"auth": map[string]interface{}{
			"token": token,
		},
		"role":   "operator",
		"scopes": []string{"operator.admin"},
	}
	_ = nonce // nonce is part of the challenge flow; connect proceeds after receiving it

	resp, err := wsc.request("connect", connectParams, 10*time.Second)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ws connect: %w", err)
	}
	if !resp.OK {
		conn.Close()
		msg := "unknown"
		if resp.Error != nil {
			msg = resp.Error.Message
		}
		return nil, fmt.Errorf("ws connect rejected: %s", msg)
	}

	return wsc, nil
}

func (wsc *WSConn) readLoop() {
	defer close(wsc.done)
	for {
		_, data, err := wsc.conn.ReadMessage()
		if err != nil {
			return
		}

		// Try parsing as response frame (has "ok" field)
		var resp wsResponse
		if err := json.Unmarshal(data, &resp); err == nil && resp.ID != "" {
			wsc.mu.Lock()
			ch, ok := wsc.pending[resp.ID]
			if ok {
				delete(wsc.pending, resp.ID)
			}
			wsc.mu.Unlock()
			if ok {
				ch <- resp
			}
			continue
		}

		// Try parsing as event frame (has "event" field)
		var evt wsEvent
		if err := json.Unmarshal(data, &evt); err == nil && evt.Event != "" {
			select {
			case wsc.events <- evt:
			default:
				// drop if buffer full
			}
		}
	}
}

func (wsc *WSConn) request(method string, params interface{}, timeout time.Duration) (wsResponse, error) {
	id := uuid.New().String()
	frame := wsFrame{
		Type:   "req",
		ID:     id,
		Method: method,
		Params: params,
	}

	ch := make(chan wsResponse, 1)
	wsc.mu.Lock()
	wsc.pending[id] = ch
	wsc.mu.Unlock()

	data, err := json.Marshal(frame)
	if err != nil {
		return wsResponse{}, err
	}
	if err := wsc.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		wsc.mu.Lock()
		delete(wsc.pending, id)
		wsc.mu.Unlock()
		return wsResponse{}, fmt.Errorf("ws write: %w", err)
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(timeout):
		wsc.mu.Lock()
		delete(wsc.pending, id)
		wsc.mu.Unlock()
		return wsResponse{}, fmt.Errorf("request timeout for %s", method)
	case <-wsc.done:
		return wsResponse{}, fmt.Errorf("connection closed")
	}
}

// Close closes the WebSocket connection.
func (wsc *WSConn) Close() {
	wsc.conn.Close()
}

// ChatResult holds the final assistant response from a chat.send call.
type ChatResult struct {
	Content string
	RunID   string
}

// SendChatWS sends a message via WebSocket and blocks until the final response
// event arrives. It returns the assembled assistant text.
func (c *Client) SendChatWS(sessionKey, message string) (*ChatResult, error) {
	wsc, err := connectWS(c.wsURL, c.token)
	if err != nil {
		return nil, fmt.Errorf("ws connect: %w", err)
	}
	defer wsc.Close()

	runID := uuid.New().String()

	// Send chat.send request
	resp, err := wsc.request("chat.send", map[string]interface{}{
		"sessionKey":     sessionKey,
		"message":        message,
		"thinking":       "adaptive",
		"idempotencyKey": runID,
	}, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("chat.send: %w", err)
	}
	if !resp.OK {
		msg := "unknown"
		if resp.Error != nil {
			msg = resp.Error.Message
		}
		return nil, fmt.Errorf("chat.send rejected: %s", msg)
	}

	// Wait for chat events until we get a "final" or "error" state
	timeout := time.After(5 * time.Minute)
	var lastContent string

	for {
		select {
		case evt := <-wsc.events:
			if evt.Event == "tick" {
				continue
			}
			if evt.Event != "chat" {
				continue
			}

			var ce chatEvent
			if err := json.Unmarshal(evt.Payload, &ce); err != nil {
				continue
			}

			switch ce.State {
			case "delta":
				text := extractTextFromEventMessage(ce.Message)
				if text != "" {
					lastContent = text
				}
			case "final":
				text := extractTextFromEventMessage(ce.Message)
				if text != "" {
					lastContent = text
				}
				return &ChatResult{Content: lastContent, RunID: ce.RunID}, nil
			case "error":
				errMsg := ce.ErrorMessage
				if errMsg == "" {
					errMsg = "agent error"
				}
				return nil, fmt.Errorf("chat error: %s", errMsg)
			case "aborted":
				return nil, fmt.Errorf("chat aborted")
			}

		case <-timeout:
			if lastContent != "" {
				return &ChatResult{Content: lastContent, RunID: runID}, nil
			}
			return nil, fmt.Errorf("timeout waiting for response")
		case <-wsc.done:
			return nil, fmt.Errorf("connection closed while waiting for response")
		}
	}
}

// ListSessionsWS lists sessions via WebSocket.
func (c *Client) ListSessionsWS() ([]Session, error) {
	wsc, err := connectWS(c.wsURL, c.token)
	if err != nil {
		return nil, fmt.Errorf("ws connect: %w", err)
	}
	defer wsc.Close()

	resp, err := wsc.request("sessions.list", map[string]interface{}{
		"limit":                50,
		"includeDerivedTitles": true,
		"includeLastMessage":   true,
	}, 10*time.Second)
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		msg := "unknown"
		if resp.Error != nil {
			msg = resp.Error.Message
		}
		return nil, fmt.Errorf("sessions.list: %s", msg)
	}

	var result struct {
		Sessions []Session `json:"sessions"`
	}
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
		return nil, fmt.Errorf("parse sessions: %w", err)
	}
	return result.Sessions, nil
}

// GetSessionHistoryWS fetches session history via WebSocket.
func (c *Client) GetSessionHistoryWS(sessionKey string, limit int) ([]Message, error) {
	wsc, err := connectWS(c.wsURL, c.token)
	if err != nil {
		return nil, fmt.Errorf("ws connect: %w", err)
	}
	defer wsc.Close()

	resp, err := wsc.request("chat.history", map[string]interface{}{
		"sessionKey": sessionKey,
		"limit":      limit,
	}, 10*time.Second)
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		msg := "unknown"
		if resp.Error != nil {
			msg = resp.Error.Message
		}
		return nil, fmt.Errorf("chat.history: %s", msg)
	}

	var result struct {
		Messages []struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(resp.Payload, &result); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}

	var msgs []Message
	for _, m := range result.Messages {
		text := extractTextFromContent(m.Content)
		if text != "" {
			msgs = append(msgs, Message{Role: m.Role, Content: text})
		}
	}
	return msgs, nil
}

// extractTextFromEventMessage extracts text content from a chat event message.
// The message can be a string, an object with content blocks, or nested formats.
func extractTextFromEventMessage(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Try as a plain string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}

	// Try as message object with content field
	var msg struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(raw, &msg); err == nil && len(msg.Content) > 0 {
		return extractTextFromContent(msg.Content)
	}

	return ""
}

// extractTextFromContent extracts text from a content field that can be
// a string or an array of content blocks.
func extractTextFromContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Try as string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}

	// Try as array of content blocks
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var text string
		for _, b := range blocks {
			if b.Type == "text" {
				text += b.Text
			}
		}
		return text
	}

	return ""
}
