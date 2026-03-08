# pushover-mcp

A [Pushover](https://pushover.net/) notification [MCP](https://modelcontextprotocol.io/) server for AI agents, built in Go. Send push notifications and manage emergency-priority messages via the Pushover API.

## Features

- **Send notifications**: Push messages to any Pushover device with full parameter support (priority, sounds, HTML, URLs, TTL)
- **Emergency priority**: Send priority=2 notifications that repeat until acknowledged, with receipt tracking
- **Receipt management**: Poll receipt status, cancel individual receipts, or cancel by tag
- **Simple configuration**: Just two environment variables (`PUSHOVER_TOKEN`, `PUSHOVER_USER_KEY`)

## MCP Tools

| Tool | Description |
|------|-------------|
| `pushover_send_message` | Send a push notification with configurable priority, sound, formatting, and more |
| `pushover_check_receipt` | Check the status of an emergency-priority notification (acknowledged, expired, etc.) |
| `pushover_cancel_receipt` | Cancel retries for a specific emergency-priority notification |
| `pushover_cancel_receipt_by_tag` | Cancel all emergency-priority notifications matching a tag |

## Install

```sh
go install github.com/freeformz/pushover-mcp@latest
```

Or pull the Docker image:

```sh
docker pull ghcr.io/freeformz/pushover-mcp:latest
```

## Configure

You need a [Pushover application token](https://pushover.net/apps) and your [user key](https://pushover.net/).

Add to your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "pushover": {
      "type": "stdio",
      "command": "pushover-mcp",
      "args": ["mcp"],
      "env": {
        "PUSHOVER_TOKEN": "your-app-token",
        "PUSHOVER_USER_KEY": "your-user-key"
      }
    }
  }
}
```

### Docker

```json
{
  "mcpServers": {
    "pushover": {
      "type": "stdio",
      "command": "docker",
      "args": ["run", "--rm", "-i", "-e", "PUSHOVER_TOKEN", "-e", "PUSHOVER_USER_KEY", "ghcr.io/freeformz/pushover-mcp:latest", "mcp"],
      "env": {
        "PUSHOVER_TOKEN": "your-app-token",
        "PUSHOVER_USER_KEY": "your-user-key"
      }
    }
  }
}
```

## Priority Levels

| Priority | Name | Behavior |
|----------|------|----------|
| -2 | Lowest | No notification, badge count only |
| -1 | Low | No sound/vibration |
| 0 | Normal | Default sound and vibration |
| 1 | High | Bypasses quiet hours, highlighted in red |
| 2 | Emergency | Repeats until acknowledged (requires `retry` and `expire`) |

### Emergency Priority Workflow

1. Send a message with `priority=2`, `retry` (min 30s), and `expire` (max 10800s)
2. You'll receive a receipt ID in the response
3. Poll the receipt with `pushover_check_receipt` to check acknowledgment
4. Cancel with `pushover_cancel_receipt` if no longer needed

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `PUSHOVER_TOKEN` | Yes | Application API token from [pushover.net/apps](https://pushover.net/apps) |
| `PUSHOVER_USER_KEY` | Yes | Your user key from [pushover.net](https://pushover.net/) |

## License

MIT
