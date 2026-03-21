# Hermes TUI

A beautiful standalone terminal UI client for the Hermes AI agent system. Built with Go, Bubble Tea, and lipgloss.

![Hermes TUI](https://img.shields.io/badge/Hermes-TUI-5EBED6?style=flat-square)

## Features

- **Full terminal interface** — powered by Bubble Tea for a responsive, modern TUI experience
- **Markdown rendering** — assistant messages rendered with glamour
- **5 color themes** — ocean, amber, rose, forest, aquarium
- **Session management** — connect to existing sessions or start new ones
- **Bearer token auth** — reads credentials from `~/.openclaw/openclaw.json`
- **Streaming responses** — watch tokens arrive in real-time

## Requirements

- Go 1.21+
- Hermes Gateway running at `localhost:18789` (or configured URL)
- `~/.openclaw/openclaw.json` with gateway credentials

## Installation

```bash
git clone https://github.com/DevvGwardo/hermes-tui.git
cd hermes-tui
go build -o hermes-tui ./cmd/hermes-tui/
```

## Usage

```bash
# Run with defaults
./hermes-tui

# Specify theme
./hermes-tui --theme rose

# Connect to a specific session
./hermes-tui --session <session-key>

# Point to a different gateway
./hermes-tui --gateway http://localhost:18789
```

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Ctrl+C` | Quit |
| Arrow keys | Scroll chat history |
| `Up/Down` | Navigate input history |

## Configuration

Config is stored at `~/.config/hermes-tui/config.json`:

```json
{
  "theme": "ocean",
  "session_id": "",
  "thinking": true,
  "history_size": 100
}
```

## Architecture

```
cmd/hermes-tui/
  main.go          — entry point, flag parsing, TUI bootstrap

internal/
  config/          — config load/save
  gateway/         — HTTP client for Hermes Gateway API
  tui/
    model.go       — main Bubble Tea model
    chat.go        — scrollable message viewport
    messages.go    — message rendering
    header.go      — top status bar
    statusbar.go   — bottom status bar
    input.go       — text input
    theme.go       — 5 color themes
```

## Themes

Preview the available themes:

| Theme | Description |
|-------|-------------|
| `ocean` | Deep blue, professional (default) |
| `amber` | Warm gold tones |
| `rose` | Soft pink and purple |
| `forest` | Earthy greens |
| `aquarium` | Deep teal, dark |

## License

MIT
