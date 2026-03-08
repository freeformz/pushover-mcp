package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestSendMessage_Required(t *testing.T) {
	handler := handleSendMessage(&PushoverClient{token: "t", userKey: "u"})
	ctx := t.Context()

	// Missing message
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}
	result, err := handler(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing message")
	}
}

func TestSendMessage_Validation(t *testing.T) {
	handler := handleSendMessage(&PushoverClient{token: "t", userKey: "u"})
	ctx := t.Context()

	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{
			name: "message too long",
			args: map[string]any{"message": string(make([]byte, 1025))},
			want: "message exceeds 1024 character limit",
		},
		{
			name: "title too long",
			args: map[string]any{"message": "test", "title": string(make([]byte, 251))},
			want: "title exceeds 250 character limit",
		},
		{
			name: "invalid priority",
			args: map[string]any{"message": "test", "priority": float64(3)},
			want: "priority must be between -2 and 2",
		},
		{
			name: "html and monospace",
			args: map[string]any{"message": "test", "html": true, "monospace": true},
			want: "html and monospace are mutually exclusive",
		},
		{
			name: "emergency without retry",
			args: map[string]any{"message": "test", "priority": float64(2), "retry": float64(0), "expire": float64(3600)},
			want: "retry must be at least 30 seconds for emergency priority",
		},
		{
			name: "emergency without expire",
			args: map[string]any{"message": "test", "priority": float64(2), "retry": float64(30)},
			want: "expire must be between 1 and 10800 seconds for emergency priority",
		},
		{
			name: "emergency expire too large",
			args: map[string]any{"message": "test", "priority": float64(2), "retry": float64(30), "expire": float64(20000)},
			want: "expire must be between 1 and 10800 seconds for emergency priority",
		},
		{
			name: "device name too long",
			args: map[string]any{"message": "test", "device": string(make([]byte, 26))},
			want: "device name exceeds 25 character limit",
		},
		{
			name: "url too long",
			args: map[string]any{"message": "test", "url": string(make([]byte, 513))},
			want: "url exceeds 512 character limit",
		},
		{
			name: "url_title too long",
			args: map[string]any{"message": "test", "url_title": string(make([]byte, 101))},
			want: "url_title exceeds 100 character limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcp.CallToolRequest{}
			req.Params.Arguments = tt.args
			result, err := handler(ctx, req)
			if err != nil {
				t.Fatal(err)
			}
			if !result.IsError {
				t.Error("expected error")
			}
			text := result.Content[0].(mcp.TextContent).Text
			if text != tt.want {
				t.Errorf("got %q, want %q", text, tt.want)
			}
		})
	}
}

func TestSendMessage_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("token") != "testtoken" {
			t.Errorf("unexpected token: %s", r.FormValue("token"))
		}
		if r.FormValue("user") != "testuserkey" {
			t.Errorf("unexpected user: %s", r.FormValue("user"))
		}
		if r.FormValue("message") != "hello" {
			t.Errorf("unexpected message: %s", r.FormValue("message"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MessageResponse{Status: 1, Request: "abc123"})
	}))
	defer srv.Close()

	client := NewPushoverClient("testtoken", "testuserkey", srv.Client())
	client.baseURL = srv.URL

	handler := handleSendMessage(client)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"message": "hello"}

	result, err := handler(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("unexpected error: %v", result.Content)
	}
}

func TestSendMessage_EmergencySuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.FormValue("priority") != "2" {
			t.Errorf("expected priority 2, got %s", r.FormValue("priority"))
		}
		if r.FormValue("retry") != "30" {
			t.Errorf("expected retry 30, got %s", r.FormValue("retry"))
		}
		if r.FormValue("expire") != "3600" {
			t.Errorf("expected expire 3600, got %s", r.FormValue("expire"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MessageResponse{Status: 1, Request: "abc123", Receipt: "receipt123"})
	}))
	defer srv.Close()

	client := NewPushoverClient("testtoken", "testuserkey", srv.Client())
	client.baseURL = srv.URL

	handler := handleSendMessage(client)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"message":  "emergency!",
		"priority": float64(2),
		"retry":    float64(30),
		"expire":   float64(3600),
	}

	result, err := handler(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("unexpected error: %v", result.Content)
	}
}

func TestCheckReceipt_Required(t *testing.T) {
	handler := handleCheckReceipt(&PushoverClient{token: "t", userKey: "u"})
	ctx := t.Context()

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}
	result, err := handler(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing receipt")
	}
}

func TestCancelReceipt_Required(t *testing.T) {
	handler := handleCancelReceipt(&PushoverClient{token: "t", userKey: "u"})
	ctx := t.Context()

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}
	result, err := handler(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing receipt")
	}
}

func TestCancelReceiptByTag_Required(t *testing.T) {
	handler := handleCancelReceiptByTag(&PushoverClient{token: "t", userKey: "u"})
	ctx := t.Context()

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}
	result, err := handler(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for missing tag")
	}
}

func TestCheckReceipt_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ReceiptResponse{
			Status:       1,
			Acknowledged: 1,
			Expired:      0,
		})
	}))
	defer srv.Close()

	client := NewPushoverClient("testtoken", "testuserkey", srv.Client())
	client.baseURL = srv.URL

	handler := handleCheckReceipt(client)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"receipt": "r123"}

	result, err := handler(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("unexpected error: %v", result.Content)
	}
}

func TestCancelReceipt_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("token") != "testtoken" {
			t.Errorf("unexpected token: %s", r.FormValue("token"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CancelResponse{Status: 1, Request: "abc123"})
	}))
	defer srv.Close()

	client := NewPushoverClient("testtoken", "testuserkey", srv.Client())
	client.baseURL = srv.URL

	handler := handleCancelReceipt(client)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"receipt": "r123"}

	result, err := handler(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("unexpected error: %v", result.Content)
	}
}

func TestCancelReceiptByTag_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("token") != "testtoken" {
			t.Errorf("unexpected token: %s", r.FormValue("token"))
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CancelResponse{Status: 1, Request: "abc123"})
	}))
	defer srv.Close()

	client := NewPushoverClient("testtoken", "testuserkey", srv.Client())
	client.baseURL = srv.URL

	handler := handleCancelReceiptByTag(client)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"tag": "deploy"}

	result, err := handler(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Errorf("unexpected error: %v", result.Content)
	}
}

func TestSendMessage_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MessageResponse{Status: 0, Errors: []string{"invalid token"}})
	}))
	defer srv.Close()

	client := NewPushoverClient("testtoken", "testuserkey", srv.Client())
	client.baseURL = srv.URL

	handler := handleSendMessage(client)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"message": "hello"}

	result, err := handler(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for API error response")
	}
}

func TestCheckReceipt_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ReceiptResponse{Status: 0, Errors: []string{"receipt not found"}})
	}))
	defer srv.Close()

	client := NewPushoverClient("testtoken", "testuserkey", srv.Client())
	client.baseURL = srv.URL

	handler := handleCheckReceipt(client)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"receipt": "bad"}

	result, err := handler(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for API error response")
	}
}

func TestCancelReceipt_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CancelResponse{Status: 0, Errors: []string{"receipt not found"}})
	}))
	defer srv.Close()

	client := NewPushoverClient("testtoken", "testuserkey", srv.Client())
	client.baseURL = srv.URL

	handler := handleCancelReceipt(client)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"receipt": "bad"}

	result, err := handler(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for API error response")
	}
}

func TestCancelReceiptByTag_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CancelResponse{Status: 0, Errors: []string{"tag not found"}})
	}))
	defer srv.Close()

	client := NewPushoverClient("testtoken", "testuserkey", srv.Client())
	client.baseURL = srv.URL

	handler := handleCancelReceiptByTag(client)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"tag": "bad"}

	result, err := handler(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for API error response")
	}
}

func TestConfigured(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		userKey string
		wantErr bool
	}{
		{"both set", "t", "u", false},
		{"missing token", "", "u", true},
		{"missing user key", "t", "", true},
		{"both missing", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &PushoverClient{token: tt.token, userKey: tt.userKey}
			err := c.Configured()
			if (err != nil) != tt.wantErr {
				t.Errorf("Configured() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandleSendMessage_NotConfigured(t *testing.T) {
	handler := handleSendMessage(&PushoverClient{})
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"message": "test"}

	result, err := handler(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected error for unconfigured client")
	}
}
