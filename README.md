# opencode-stats

A CLI tool to display usage statistics from the OpenCode SQLite database. It reads `~/.local/share/opencode/opencode.db` and provides formatted tables showing session details, daily aggregates, breakdowns by model/agent/project, and all-time summary statistics.

## Installation

### Option 1: Go install (requires Go 1.23+)

```bash
go install github.com/abuelgheit/opencode-stats@latest
```

### Option 2: One-liner (no Go needed)

```bash
curl -sSL https://raw.githubusercontent.com/abuelgheit/opencode-stats/main/install.sh | bash
```

### Option 3: Manual download

Download the latest binary from [GitHub Releases](https://github.com/abuelgheit/opencode-stats/releases).

### Verify

```bash
opencode-stats version
```

## Usage

```
opencode-stats [command] [flags]
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
opencode-stats session

# Show last 50 sessions
opencode-stats session --limit 50

# Show sessions from the last 7 days
opencode-stats session --days 7

# Show last 10 sessions from the last 30 days
opencode-stats session -n 10 -d 30
```

#### `daily`

Show daily aggregates of sessions, token usage, and cost.

```bash
# Show last 30 days (default)
opencode-stats daily

# Show last 7 days
opencode-stats daily --days 7
```

#### `models`

Show usage statistics grouped by model.

```bash
# Show all-time model breakdown
opencode-stats models

# Show model breakdown for the last 14 days
opencode-stats models --days 14
```

#### `agents`

Show usage statistics grouped by agent type.

```bash
# Show all-time agent breakdown
opencode-stats agents

# Show agent breakdown for the last 7 days
opencode-stats agents --days 7
```

#### `projects`

Show usage statistics grouped by project.

```bash
# Show all-time project breakdown
opencode-stats projects

# Show project breakdown for the last 30 days
opencode-stats projects --days 30
```

#### `summary`

Show all-time totals for sessions, tokens, and cost.

```bash
opencode-stats summary
```

### Examples

```bash
# Quick overview
opencode-stats summary

# Recent activity
opencode-stats session -d 7

# Cost breakdown by model
opencode-stats models

# Daily usage for the last week
opencode-stats daily -d 7

# Use a custom database path
opencode-stats --db /custom/path/opencode.db summary
```

## Build from Source

```bash
git clone <repo-url>
cd opencode-stats
make build    # builds to bin/opencode-stats
make test     # runs all tests with coverage
```

## Development

Run tests:

```bash
go test ./... -v -cover
```
