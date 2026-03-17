# ccmonitor

A terminal dashboard for monitoring [Claude Code](https://docs.anthropic.com/en/docs/claude-code) usage, sessions, and rate limits.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Features

- **Dashboard** — today/lifetime message & token counts, active sessions, rate limit utilization
- **Activity** — daily message sparkline chart + recent prompt history
- **Processes** — full-width table of running Claude Code instances with CPU/memory/uptime
- **Rate limits** (macOS only) — 5-hour and 7-day window utilization with progress bars, reset countdowns, and burn-rate indicators
- Auto-refreshing (configurable interval)
- Responsive layout (adapts to narrow terminals)
- Graceful degradation when data sources are unavailable

## Data sources

| Data | Source |
|------|--------|
| Messages, sessions, tokens | `~/.claude/stats-cache.json` |
| Recent prompts | `~/.claude/history.jsonl` |
| Active processes | `ps` + `lsof` (macOS) or `/proc` (Linux) |
| Rate limits | Anthropic OAuth usage API via macOS Keychain (macOS only) |

## Requirements

- Go 1.25+
- macOS or Linux
- Claude Code installed (for data in `~/.claude/`)
- macOS Keychain credentials for rate limit data (macOS only, optional)

## Install

### Pre-built binary

Download the latest release for your platform:

```bash
# macOS (Apple Silicon)
curl -L https://github.com/mkrowiarz/ccmonitor/releases/latest/download/ccmonitor-darwin-arm64 -o /usr/local/bin/ccmonitor && chmod +x /usr/local/bin/ccmonitor

# macOS (Intel)
curl -L https://github.com/mkrowiarz/ccmonitor/releases/latest/download/ccmonitor-darwin-amd64 -o /usr/local/bin/ccmonitor && chmod +x /usr/local/bin/ccmonitor

# Linux (x86_64)
curl -L https://github.com/mkrowiarz/ccmonitor/releases/latest/download/ccmonitor-linux-amd64 -o /usr/local/bin/ccmonitor && chmod +x /usr/local/bin/ccmonitor

# Linux (arm64)
curl -L https://github.com/mkrowiarz/ccmonitor/releases/latest/download/ccmonitor-linux-arm64 -o /usr/local/bin/ccmonitor && chmod +x /usr/local/bin/ccmonitor
```

### Go install

```
go install github.com/mkrowiarz/ccmonitor@latest
```

### From source

```
git clone https://github.com/mkrowiarz/ccmonitor.git
cd ccmonitor
make install        # installs to ~/.local/bin
```

Make sure `~/.local/bin` is in your `PATH`.

### Uninstall

```
make uninstall
```

## Usage

```
ccmonitor                    # default: 10s refresh
ccmonitor -interval 5        # 5s refresh
ccmonitor -no-rate-limits    # hide the rate limits panel
ccmonitor -minimal           # dashboard only, no activity/analytics tabs
```

| Flag | Description |
|------|-------------|
| `-interval N` | Refresh interval in seconds (default: 10) |
| `-no-rate-limits` | Disable the rate limits panel |
| `-minimal` | Dashboard only — no Activity/Analytics tabs |
| `-version` | Print version and exit |

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `q` | Quit |
| `r` | Force refresh |
| `Tab` | Cycle tabs |
| `1` | Dashboard |
| `2` | Activity |
| `3` | Analytics |

## Platform notes

| Feature | macOS | Linux |
|---------|-------|-------|
| Process monitoring | `ps` + `lsof` | `ps` + `/proc` |
| Usage stats | `~/.claude/stats-cache.json` | `~/.claude/stats-cache.json` |
| Activity history | `~/.claude/history.jsonl` | `~/.claude/history.jsonl` |
| Rate limits | Keychain + OAuth API | Not available |

On Linux, the rate limits panel is hidden and no API calls are made.

## Credits

- [kvaps/claude-code-usage](https://gist.github.com/kvaps/84fa5963df1bff9cec65b57afd54e1e4) — inspiration for the usage API integration
- [MacDev](https://github.com/arvindjuneja/MacDev) — initial idea for an open-source Claude Code monitor
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — terminal UI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — terminal styling and layout
- [ntcharts](https://github.com/NimbleMarkets/ntcharts) — terminal bar charts
- [Catppuccin](https://catppuccin.com/) — color theme
- [Zellij](https://zellij.dev/) — tmux replacement that makes the Claude Monitor look beautiful
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) — AI pair programmer that co-built this project

## Author

Michał Krowiarz — [@mkrowiarz](https://github.com/mkrowiarz)

Built with [Claude Code](https://docs.anthropic.com/en/docs/claude-code) (Anthropic's AI coding agent).

## License

MIT
