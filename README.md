# opencode-usage

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A CLI tool to display usage statistics from the OpenCode SQLite database. It reads `~/.local/share/opencode/opencode.db` and provides formatted tables showing session details, daily aggregates, breakdowns by model/agent/project, and all-time summary statistics.

## Installation

### Option 1: Go install (requires Go 1.23+)

```bash
go install github.com/abuelgheit/opencode-usage@latest
```

### Option 2: One-liner (no Go needed)

```bash
curl -sSL https://raw.githubusercontent.com/abuelgheit/opencode-usage/main/install.sh | bash
```

### Option 3: Manual download

Download the latest binary from [GitHub Releases](https://github.com/abuelgheit/opencode-usage/releases).

### Verify

```bash
opencode-usage version
```

## Usage

```
opencode-usage [command] [flags]
```

### Global Flags

| Flag   | Default                                         | Description                  |
|--------|-------------------------------------------------|------------------------------|
| `--db` | `~/.local/share/opencode/opencode.db`           | Path to opencode database    |

### Commands

#### `session`

Show recent sessions with token usage and cost.

```bash
# Show last 20 sessions
opencode-usage session

# Show last 50 sessions
opencode-usage session --limit 50

# Show sessions from the last 7 days
opencode-usage session --days 7

# Show last 10 sessions from the last 30 days
opencode-usage session -n 10 -d 30
```

#### `daily`

Show daily aggregates of sessions, token usage, and cost.

```bash
# Show last 30 days (default)
opencode-usage daily

# Show last 7 days
opencode-usage daily --days 7
```

#### `models`

Show usage statistics grouped by model.

```bash
# Show all-time model breakdown
opencode-usage models

# Show model breakdown for the last 14 days
opencode-usage models --days 14
```

#### `agents`

Show usage statistics grouped by agent type.

```bash
# Show all-time agent breakdown
opencode-usage agents

# Show agent breakdown for the last 7 days
opencode-usage agents --days 7
```

#### `projects`

Show usage statistics grouped by project.

```bash
# Show all-time project breakdown
opencode-usage projects

# Show project breakdown for the last 30 days
opencode-usage projects --days 30
```

#### `summary`

Show all-time totals for sessions, tokens, and cost.

```bash
opencode-usage summary
```

### Examples

```bash
# Quick overview
opencode-usage summary

# Recent activity
opencode-usage session -d 7

# Cost breakdown by model
opencode-usage models

# Daily usage for the last week
opencode-usage daily -d 7

# Use a custom database path
opencode-usage --db /custom/path/opencode.db summary
```

## Build from Source

```bash
git clone <repo-url>
cd opencode-usage
make build    # builds to bin/opencode-usage
make test     # runs all tests with coverage
```

## Development

Run tests:

```bash
go test ./... -v -cover
```

## License

MIT © [Abuelgheit](https://github.com/abuelgheit)
