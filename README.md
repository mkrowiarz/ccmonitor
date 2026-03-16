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
make install        # installs to /usr/local/bin
```

Or to a custom location:

```
make install PREFIX=~/.local
```

### Uninstall

```
make uninstall
```

## Usage

```
ccmonitor                    # default: 5s refresh
ccmonitor -interval 10       # 10s refresh
```

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `1` `2` `3` | Switch tab (Dashboard, Activity, Processes) |
| `Tab` | Cycle tabs |
| `r` | Force refresh |
| `q` | Quit |

## Platform notes

| Feature | macOS | Linux |
|---------|-------|-------|
| Process monitoring | `ps` + `lsof` | `ps` + `/proc` |
| Usage stats | `~/.claude/stats-cache.json` | `~/.claude/stats-cache.json` |
| Activity history | `~/.claude/history.jsonl` | `~/.claude/history.jsonl` |
| Rate limits | Keychain + OAuth API | Not available |

On Linux, the rate limits panel is hidden and no API calls are made.

## Author

Michał Krowiarz — [@mkrowiarz](https://github.com/mkrowiarz)

## License

MIT
