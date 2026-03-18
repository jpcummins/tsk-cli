# tsk-cli

A terminal user interface for querying and browsing [tsk](https://github.com/jpcummins/tsk) repositories.

https://github.com/user-attachments/assets/8e944ca2-25e3-42dc-a494-13bdb2e258de

## Features

- **Dual search modes**: DSL query language and fuzzy search
- **Live search**: Fuzzy mode updates results as you type
- **Keyboard-driven navigation**: vim-style bindings (j/k), tab between panes
- **Result highlighting**: Matched terms highlighted in fuzzy search results
- **Task detail view**: Browse through results with j/k, highlighted matches in task bodies
- **Built-in help**: Ctrl+H for complete DSL reference

## Installation

```bash
go install github.com/jpcummins/tsk-cli@latest
```

Or build from source:

```bash
git clone https://github.com/jpcummins/tsk-cli
cd tsk-cli
go build -o tsk-cli .
```

## Usage

```bash
# Run in a tsk repository (looks for tasks/ directory)
tsk-cli

# Point to a specific repository
tsk-cli --repo /path/to/project

# Set user identity for me() queries
tsk-cli --me alice@example.com
```

## Key Bindings

### Global
- **Ctrl+C**: Quit
- **Ctrl+T**: Toggle between query (DSL) and fuzzy search modes
- **Ctrl+H**: Show query language help (query mode only)

### Search
- **Enter**: Execute search (query mode) or open selected result
- **Esc**: Clear results / return to search

### Results
- **Tab**: Switch focus between search box and results list
- **j/k** or **↑/↓**: Navigate results
- **Enter**: Open selected task

### Detail View
- **j/k** or **↑/↓**: Browse to next/previous result
- **Esc**: Return to results list
- Viewport scrolls with arrow keys

## Search Modes

### Query Mode (DSL)
Structured queries using the tsk query language. Press **Enter** to execute.

Examples:
```
status.category = in_progress
assignee = me() AND status.category != done
has(labels, "bug") AND due < date("tomorrow")
path ~ "backend" AND assignee = team("platform")
```

Press **Ctrl+H** for full reference.

### Fuzzy Mode
Real-time substring search across all task content (path, summary, labels, body). Results update as you type.

Examples:
- `payment timeout`
- `cart bug`
- `alice backend`

## Requirements

- Go 1.25+
- A tsk repository (see [tsk-spec](https://github.com/jpcummins/tsk-spec))

## Related Projects

- [tsk-spec](https://github.com/jpcummins/tsk-spec) - Formal specification
- [tsk-lib](https://github.com/jpcummins/tsk-lib) - Go library for parsing and querying tsk repositories

## License

MIT License - see [LICENSE](LICENSE)

## Author

J.P. Cummins <jcummins@hey.com>
