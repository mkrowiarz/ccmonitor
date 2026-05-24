# ccmonitor

<a href="https://github.com/mkrowiarz/ccmonitor/releases"><img src="https://img.shields.io/github/release/mkrowiarz/ccmonitor.svg" alt="Latest Release"></a>
<a href="https://github.com/mkrowiarz/ccmonitor/blob/main/LICENSE"><img src="https://img.shields.io/github/license/mkrowiarz/ccmonitor" alt="License"></a>

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
GOBIN=~/.local/bin go install github.com/mkrowiarz/ccmonitor@latest
```

Or without `GOBIN`, it installs to `~/go/bin` by default.

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
ccmonitor -waybar            # print one Waybar JSON line and exit
```

| Flag | Description |
|------|-------------|
| `-interval N` | Refresh interval in seconds (default: 10) |
| `-no-rate-limits` | Disable the rate limits panel |
| `-minimal` | Dashboard only — no Activity/Analytics tabs |
| `-waybar` | Print one [Waybar](#waybar) JSON line for rate limits and exit |
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
| Rate limits | Keychain + OAuth API | `~/.claude/.credentials.json` + OAuth API |

## Rate limits

The rate limits panel shows your Anthropic API usage across two rolling windows:

- **5-hour window** — short-term burst usage
- **7-day window** — longer-term sustained usage

Each window displays utilization percentage, a progress bar, time until reset, and a burn-rate indicator dot that compares your usage pace against the window's elapsed time:

| Dot | Condition | Meaning |
|-----|-----------|---------|
| Green | usage ≤ elapsed + 5% | On track — usage is proportional to time elapsed |
| Yellow | usage > elapsed + 5% | Elevated — burning faster than the window replenishes |
| Red | usage > elapsed + 15% | Hot — at risk of hitting the limit before the window resets |

For example, if 50% of the 5-hour window has elapsed but you've used 70% of your quota, the difference is +20% — the dot turns red.

### How it works

1. ccmonitor reads your OAuth access token — from the macOS Keychain (`Claude Code-credentials`) on macOS, or from `~/.claude/.credentials.json` on Linux
2. It calls the Anthropic usage API (`api.anthropic.com/api/oauth/usage`) to fetch current utilization
3. Results are cached locally (`~/.ccmonitor/usage-cache.json`) with a **10-minute TTL** — the API is not called more frequently than that
4. If the API returns 429 (rate limited), ccmonitor enters a ~10-minute cooldown before retrying

Use `-no-rate-limits` to disable this feature entirely.

## Waybar

`ccmonitor -waybar` prints a single line of [Waybar](https://github.com/Alexays/Waybar) JSON
(`text`, `tooltip`, `class`) describing your rate-limit windows, then exits. It reuses the same
10-minute on-disk cache as the dashboard, so polling every minute only hits the API every 10 minutes.

Add a custom module to `~/.config/waybar/config.jsonc`:

```jsonc
"custom/claude": {
    "exec": "ccmonitor -waybar",
    "return-type": "json",
    "interval": 60,
    "tooltip": true,
    "on-click": "ghostty -e ccmonitor"  // optional: open the full dashboard
}
```

…and reference `"custom/claude"` in one of your `modules-*` arrays. The module exposes a CSS
`class` you can style by utilization:

```css
#custom-claude.ok       { color: #a6d189; }            /* < 50%  */
#custom-claude.warning  { color: #e5c890; }            /* 50–80% */
#custom-claude.critical { color: #e78284; font-weight: bold; } /* ≥ 80% */
#custom-claude.error    { color: #838ba7; }            /* token missing / API down */
```

The `text` starts with a Nerd Font glyph (`󰚩`), so the module needs a Nerd Font in your bar.

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
