# Architecture

```
main.go                         Entry point
internal/
  backend/                      Backend interface + registry
  claude/
    backend.go                  Collector: processes, stats, history, rate limits
    credentials.go              OAuth token from macOS Keychain
    usage.go                    Rate limit API client with disk-backed cache
    process_darwin.go           macOS process discovery
    process_linux.go            Linux process discovery
    statscache.go               Stats cache parser
    history.go                  History parser
  domain/
    snapshot.go                 Shared types
  format/
    duration.go                 Compact duration formatting
    numbers.go                  Compact number formatting
  tui/
    app.go                      Bubble Tea model + layout
    panels.go                   Panel renderers
    header.go / footer.go       Chrome
    keys.go                     Key bindings
    styles.go                   Color scheme
```
