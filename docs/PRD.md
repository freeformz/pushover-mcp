# Pushover MCP Server — Product Requirements Document

## Overview

An MCP (Model Context Protocol) server written in Go that exposes the Pushover notification service as MCP tools. This allows AI agents (e.g., Claude) to send push notifications and manage emergency-priority receipts via the Pushover API.

**Repository:** `github.com/freeformz/pushover-mcp`

## Goals

- Provide MCP tools for sending Pushover notifications with full parameter support
- Support emergency-priority notification lifecycle (send, poll, cancel)
- Follow the same project structure and patterns as `github.com/freeformz/tasks-mcp`
- Ship as a standalone binary, Docker container, and MCPB bundles
- Installable via `go install github.com/freeformz/pushover-mcp@latest`

## Non-Goals (for now)

- No CLI subcommands beyond `mcp` (future possibility, unlikely)
- No image/file attachments (text-based messages only)
- No per-call user key override (use environment variable only)

## Architecture

### Framework & Dependencies

- **Go 1.26.1**
- **MCP SDK:** `github.com/mark3labs/mcp-go` (same as tasks-mcp)
- **Transport:** stdio (via `server.ServeStdio`)
- **HTTP Client:** `net/http` (stdlib) for Pushover API calls

### Project Structure

```
pushover-mcp/
├── main.go              # Entry point, Cobra CLI with `mcp` subcommand
├── server.go            # MCP server factory (NewServer)
├── tools.go             # MCP tool definitions and registration
├── handlers.go          # Tool handler implementations
├── pushover.go          # Pushover API client
├── models.go            # Request/response structs
├── Dockerfile           # Multi-stage distroless build
├── Makefile             # Build/test/lint targets
├── .goreleaser.yml      # Cross-platform release config
├── manifest.json        # MCP manifest (for distribution)
├── .mcp.json            # Claude Code dev config
├── docs/
│   └── PRD.md           # This document
└── *_test.go            # Tests
```

### Configuration

All configuration via environment variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `PUSHOVER_TOKEN` | Yes | Application API token (30-char alphanumeric) |
| `PUSHOVER_USER_KEY` | Yes | Default user/group key (30-char alphanumeric) |

The server should validate these at startup and fail with a clear error if missing.

## MCP Tools

### 1. `pushover_send_message`

Send a push notification via the Pushover Messages API.

**Endpoint:** `POST https://api.pushover.net/1/messages.json`

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `message` | string | Yes | Message body (max 1024 UTF-8 chars) |
| `title` | string | No | Message title (max 250 chars, defaults to app name) |
| `priority` | number | No | -2 (lowest), -1 (low), 0 (normal, default), 1 (high), 2 (emergency) |
| `sound` | string | No | Notification sound name |
| `device` | string | No | Target device name (max 25 chars) |
| `url` | string | No | Supplementary URL (max 512 chars) |
| `url_title` | string | No | Title for supplementary URL (max 100 chars) |
| `html` | boolean | No | Enable HTML formatting in message body |
| `monospace` | boolean | No | Enable monospace font (mutually exclusive with html) |
| `timestamp` | number | No | Unix timestamp to display as message time |
| `ttl` | number | No | Time to live in seconds (auto-delete after expiry) |
| `retry` | number | No | Retry interval in seconds (min 30, required when priority=2) |
| `expire` | number | No | Expiration in seconds (max 10800, required when priority=2) |
| `callback` | string | No | Callback URL for emergency acknowledgment |
| `tags` | string | No | Comma-separated tags for receipt management |

**Validation:**
- If `priority` is 2, `retry` (>= 30) and `expire` (<= 10800) are required
- `html` and `monospace` are mutually exclusive
- Character limits enforced before sending

**Returns:** Request ID. For priority=2, also returns the receipt ID.

### 2. `pushover_check_receipt`

Poll the status of an emergency-priority notification receipt.

**Endpoint:** `GET https://api.pushover.net/1/receipts/{receipt}.json?token={token}`

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `receipt` | string | Yes | Receipt ID from an emergency-priority message |

**Returns:** Receipt status including: acknowledged (bool), acknowledged_at, acknowledged_by, acknowledged_by_device, last_delivered_at, expired (bool), expires_at, called_back (bool), called_back_at.

### 3. `pushover_cancel_receipt`

Cancel retries for an emergency-priority notification.

**Endpoint:** `POST https://api.pushover.net/1/receipts/{receipt}/cancel.json`

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `receipt` | string | Yes | Receipt ID to cancel |

**Returns:** Confirmation of cancellation.

### 4. `pushover_cancel_receipt_by_tag`

Cancel all emergency-priority notifications matching a tag.

**Endpoint:** `POST https://api.pushover.net/1/receipts/cancel_by_tag/{tag}.json`

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `tag` | string | Yes | Tag to cancel receipts for |

**Returns:** Confirmation of cancellation.

## Server Instructions

The MCP server should include instructions (via `server.WithInstructions`) that guide agents on:

- What each tool does and when to use it
- Emergency priority workflow (send → poll receipt → cancel if needed)
- That `retry` and `expire` are required for priority=2
- Rate limit awareness (don't send more than 2 concurrent requests)

## Error Handling

- Pushover API errors (HTTP 4xx) should be returned as tool errors with `IsError: true`
- Include the Pushover error messages in the response
- HTTP 429 (rate limited) should include a clear message about quota exhaustion
- Network errors should be returned as tool errors with actionable context

## Build & Distribution

### Dockerfile

Multi-stage build following tasks-mcp pattern:
1. `golang:1.26` builder stage with `CGO_ENABLED=0`
2. `gcr.io/distroless/static-debian12` runtime
3. Entrypoint: `["/pushover-mcp"]`

### Makefile

Targets: `build`, `test`, `test-coverage`, `vet`, `lint`, `release-snapshot`, `release`

### GoReleaser

Cross-platform builds for darwin/linux/windows on amd64/arm64, matching tasks-mcp config.

### MCP Manifest & MCPB Bundles

A `manifest.json` in the repo root serves as a template with:
- `manifest_version`: `"0.3"`
- Server type `"binary"`, entry point, and `mcp_config` (command/args/env using `${__dirname}` placeholder)
- Tool names and descriptions
- Placeholder version `"0.0.0"` and all platforms listed

**MCPB bundles** (`.mcpb` files) are ZIP archives containing:
1. The patched `manifest.json` (real version, single target platform)
2. A `server/` directory with the compiled binary

Built in the GitHub Actions release workflow:
1. GoReleaser produces cross-compiled binaries for all OS/arch combinations
2. A post-GoReleaser step parses `dist/artifacts.json` to find each binary
3. For each binary, creates a temp directory with `server/<binary>` and a patched `manifest.json` (version and platform injected via `jq`)
4. Zips into `pushover-mcp_{VERSION}_{OS}_{ARCH}.mcpb`
5. Uploads all `.mcpb` files to the GitHub release

This produces 6 bundles per release: darwin/linux/windows × amd64/arm64.

## Future Possibilities

These are explicitly out of scope for the initial version but noted for potential future work:

- **User/Group Validation API** — validate user/group keys
- **Sounds API** — list available notification sounds
- **App Limits API** — check remaining monthly message quota
- **Groups API** — create and manage delivery groups
- **Subscriptions API** — manage subscription-based user keys
- **Licensing API** — assign license credits to users
- **Glances API** — push data to smartwatch/lock screen widgets
- **Image attachments** — send images with notifications
- **Per-call user key override** — allow tools to target different users
- **CLI subcommands** — send notifications from the terminal (unlikely)
