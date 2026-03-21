package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hermes-ai/hermes-tui/internal/config"
	"github.com/hermes-ai/hermes-tui/internal/gateway"
	"github.com/hermes-ai/hermes-tui/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	theme := flag.String("theme", "", "Theme: ocean, amber, rose, forest, aquarium")
	session := flag.String("session", "", "Session key to connect to")
	gatewayURL := flag.String("gateway", "http://localhost:18789", "Gateway URL")
	flag.Parse()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to load config: %v\n", err)
		cfg = config.Default()
	}

	// CLI flags override config
	if *theme != "" {
		cfg.Theme = *theme
	}
	if *session != "" {
		cfg.SessionID = *session
	}

	// Create gateway client
	gw, err := gateway.NewClient(*gatewayURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create gateway client: %v\n", err)
		os.Exit(1)
	}

	// Run TUI
	model := tui.NewModel(gw, cfg.SessionID, cfg.Theme, cfg)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
