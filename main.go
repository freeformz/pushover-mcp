package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "pushover-mcp",
		Short: "Pushover MCP server for AI agents",
		Long:  "pushover-mcp is an MCP server that exposes the Pushover notification service as tools for AI agents.",
	}

	rootCmd.AddCommand(mcpCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func mcpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server (stdio)",
		RunE: func(cmd *cobra.Command, args []string) error {
			token := os.Getenv("PUSHOVER_TOKEN")
			if token == "" {
				return fmt.Errorf("PUSHOVER_TOKEN environment variable is required")
			}
			userKey := os.Getenv("PUSHOVER_USER_KEY")
			if userKey == "" {
				return fmt.Errorf("PUSHOVER_USER_KEY environment variable is required")
			}

			client := NewPushoverClient(token, userKey, http.DefaultClient)
			srv := NewServer(client)

			return server.ServeStdio(srv)
		},
	}
}
