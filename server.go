package main

import (
	"github.com/mark3labs/mcp-go/server"
)

func NewServer(client *PushoverClient) *server.MCPServer {
	srv := server.NewMCPServer(
		"pushover-mcp",
		version,
		server.WithToolCapabilities(true),
		server.WithInstructions(`Pushover notification MCP server. Use these tools to send push notifications and manage emergency-priority messages.

Available tools:
- pushover_send_message: Send a push notification. Supports priority levels from -2 (lowest) to 2 (emergency).
- pushover_check_receipt: Check the status of an emergency-priority notification (was it acknowledged, has it expired, etc.).
- pushover_cancel_receipt: Cancel retries for a specific emergency-priority notification.
- pushover_cancel_receipt_by_tag: Cancel all emergency-priority notifications matching a tag.

Emergency priority workflow:
1. Send a message with priority=2, retry (min 30 seconds), and expire (max 10800 seconds). You will receive a receipt ID.
2. Optionally poll the receipt with pushover_check_receipt to check if the user acknowledged it.
3. Cancel with pushover_cancel_receipt if no longer needed.

Important notes:
- priority=2 (emergency) REQUIRES retry and expire parameters.
- html and monospace formatting are mutually exclusive.
- Do not send more than 2 concurrent requests to avoid rate limiting.`),
	)

	registerTools(srv, client)
	return srv
}
