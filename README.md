# ccmonitor

A terminal dashboard for monitoring [Claude Code](https://docs.anthropic.com/en/docs/claude-code) usage, sessions, and rate limits.

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Features

- **Dashboard** — today/lifetime message & token counts, active sessions, rate limit utilization
- **Activity** — daily message sparkline chart + recent prompt history
- **Processes** — full-width table of running Claude Code instances with CPU/memory/uptime
- **Rate limits** — 5-hour and 7-day window utilization with progress bars, reset countdowns, and burn-rate indicators
- Auto-refreshing (configurable interval)
- Graceful degradation when data sources are unavailable

## Data sources

| Data | Source |
|------|--------|
| Messages, sessions, tokens | `~/.claude/stats-cache.json` |
| Recent prompts | `~/.claude/history.jsonl` |
| Active processes | `ps` + `lsof` (macOS) or `/proc` (Linux) |
| Rate limits | Anthropic OAuth usage API (token from macOS Keychain) |

## Requirements

- Go 1.25+
- macOS or Linux
- Claude Code installed (for data in `~/.claude/`)
- macOS Keychain credentials for rate limit data (optional)

## Install

```
go install github.com/michal/ccmonitor@latest
```

Or build from source:

```
git clone https://github.com/mkrowiarz/ccmonitor.git
cd ccmonitor
make build
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

## License

MIT
