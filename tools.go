package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerTools(srv *server.MCPServer, client *PushoverClient) {
	srv.AddTool(
		mcp.NewTool("pushover_send_message",
			mcp.WithDescription("Send a push notification via the Pushover API."),
			mcp.WithString("message", mcp.Description("Message body (max 1024 UTF-8 characters)"), mcp.Required()),
			mcp.WithString("title", mcp.Description("Message title (max 250 characters, defaults to app name)")),
			mcp.WithNumber("priority", mcp.Description("Priority: -2 (lowest), -1 (low), 0 (normal), 1 (high), 2 (emergency). Default: 0")),
			mcp.WithString("sound", mcp.Description("Notification sound name")),
			mcp.WithString("device", mcp.Description("Target device name (max 25 characters)")),
			mcp.WithString("url", mcp.Description("Supplementary URL (max 512 characters)")),
			mcp.WithString("url_title", mcp.Description("Title for supplementary URL (max 100 characters)")),
			mcp.WithBoolean("html", mcp.Description("Enable HTML formatting (mutually exclusive with monospace)")),
			mcp.WithBoolean("monospace", mcp.Description("Enable monospace font (mutually exclusive with html)")),
			mcp.WithNumber("timestamp", mcp.Description("Unix timestamp to display as message time")),
			mcp.WithNumber("ttl", mcp.Description("Time to live in seconds (auto-delete after expiry)")),
			mcp.WithNumber("retry", mcp.Description("Retry interval in seconds (min 30, required for priority=2)")),
			mcp.WithNumber("expire", mcp.Description("Expiration in seconds (max 10800, required for priority=2)")),
			mcp.WithString("callback", mcp.Description("Callback URL for emergency acknowledgment")),
			mcp.WithString("tags", mcp.Description("Comma-separated tags for receipt management")),
		),
		handleSendMessage(client),
	)

	srv.AddTool(
		mcp.NewTool("pushover_check_receipt",
			mcp.WithDescription("Check the status of an emergency-priority notification receipt."),
			mcp.WithString("receipt", mcp.Description("Receipt ID from an emergency-priority message"), mcp.Required()),
		),
		handleCheckReceipt(client),
	)

	srv.AddTool(
		mcp.NewTool("pushover_cancel_receipt",
			mcp.WithDescription("Cancel retries for an emergency-priority notification."),
			mcp.WithString("receipt", mcp.Description("Receipt ID to cancel"), mcp.Required()),
		),
		handleCancelReceipt(client),
	)

	srv.AddTool(
		mcp.NewTool("pushover_cancel_receipt_by_tag",
			mcp.WithDescription("Cancel all emergency-priority notifications matching a tag."),
			mcp.WithString("tag", mcp.Description("Tag to cancel receipts for"), mcp.Required()),
		),
		handleCancelReceiptByTag(client),
	)
}

func handleSendMessage(client *PushoverClient) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.Configured(); err != nil {
			return errResult(err.Error()), nil
		}

		message := request.GetString("message", "")
		if message == "" {
			return errResult("message is required"), nil
		}
		if len(message) > 1024 {
			return errResult("message exceeds 1024 character limit"), nil
		}

		title := request.GetString("title", "")
		if len(title) > 250 {
			return errResult("title exceeds 250 character limit"), nil
		}

		priority := int(request.GetFloat("priority", 0))
		if priority < -2 || priority > 2 {
			return errResult("priority must be between -2 and 2"), nil
		}

		html := request.GetBool("html", false)
		monospace := request.GetBool("monospace", false)
		if html && monospace {
			return errResult("html and monospace are mutually exclusive"), nil
		}

		retry := int(request.GetFloat("retry", 0))
		expire := int(request.GetFloat("expire", 0))

		if priority == 2 {
			if retry < 30 {
				return errResult("retry must be at least 30 seconds for emergency priority"), nil
			}
			if expire <= 0 || expire > 10800 {
				return errResult("expire must be between 1 and 10800 seconds for emergency priority"), nil
			}
		}

		device := request.GetString("device", "")
		if len(device) > 25 {
			return errResult("device name exceeds 25 character limit"), nil
		}

		urlStr := request.GetString("url", "")
		if len(urlStr) > 512 {
			return errResult("url exceeds 512 character limit"), nil
		}

		urlTitle := request.GetString("url_title", "")
		if len(urlTitle) > 100 {
			return errResult("url_title exceeds 100 character limit"), nil
		}

		req := &MessageRequest{
			Message:   message,
			Title:     title,
			Priority:  priority,
			Sound:     request.GetString("sound", ""),
			Device:    device,
			URL:       urlStr,
			URLTitle:  urlTitle,
			Timestamp: int64(request.GetFloat("timestamp", 0)),
			TTL:       int(request.GetFloat("ttl", 0)),
			Retry:     retry,
			Expire:    expire,
			Callback:  request.GetString("callback", ""),
			Tags:      request.GetString("tags", ""),
		}

		if html {
			req.HTML = 1
		}
		if monospace {
			req.Monospace = 1
		}

		resp, err := client.SendMessage(ctx, req)
		if err != nil {
			return errResult(fmt.Sprintf("failed to send message: %s", err)), nil
		}

		if resp.Status != 1 {
			return errResult(fmt.Sprintf("pushover API error: %s", strings.Join(resp.Errors, "; "))), nil
		}

		return jsonResult(resp)
	}
}

func handleCheckReceipt(client *PushoverClient) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.Configured(); err != nil {
			return errResult(err.Error()), nil
		}

		receipt := request.GetString("receipt", "")
		if receipt == "" {
			return errResult("receipt is required"), nil
		}

		resp, err := client.CheckReceipt(ctx, receipt)
		if err != nil {
			return errResult(fmt.Sprintf("failed to check receipt: %s", err)), nil
		}

		if resp.Status != 1 {
			return errResult(fmt.Sprintf("pushover API error: %s", strings.Join(resp.Errors, "; "))), nil
		}

		return jsonResult(resp)
	}
}

func handleCancelReceipt(client *PushoverClient) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.Configured(); err != nil {
			return errResult(err.Error()), nil
		}

		receipt := request.GetString("receipt", "")
		if receipt == "" {
			return errResult("receipt is required"), nil
		}

		resp, err := client.CancelReceipt(ctx, receipt)
		if err != nil {
			return errResult(fmt.Sprintf("failed to cancel receipt: %s", err)), nil
		}

		if resp.Status != 1 {
			return errResult(fmt.Sprintf("pushover API error: %s", strings.Join(resp.Errors, "; "))), nil
		}

		return jsonResult(resp)
	}
}

func handleCancelReceiptByTag(client *PushoverClient) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if err := client.Configured(); err != nil {
			return errResult(err.Error()), nil
		}

		tag := request.GetString("tag", "")
		if tag == "" {
			return errResult("tag is required"), nil
		}

		resp, err := client.CancelReceiptByTag(ctx, tag)
		if err != nil {
			return errResult(fmt.Sprintf("failed to cancel receipts by tag: %s", err)), nil
		}

		if resp.Status != 1 {
			return errResult(fmt.Sprintf("pushover API error: %s", strings.Join(resp.Errors, "; "))), nil
		}

		return jsonResult(resp)
	}
}

func errResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{mcp.NewTextContent(msg)},
		IsError: true,
	}
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return mcp.NewToolResultText(string(b)), nil
}
