# csvpp CLI

A command-line tool for working with CSV++ files.

## Installation

```bash
GOEXPERIMENT=jsonv2 go install github.com/osamingo/go-csvpp/cmd/csvpp@latest
```

Or build from source:

```bash
GOEXPERIMENT=jsonv2 go build -o csvpp ./cmd/csvpp
```

> **Note:** `GOEXPERIMENT=jsonv2` is required because this tool depends on `encoding/json/jsontext` (Go 1.26+).

## Commands

### validate

Validate CSV++ file syntax.

```bash
# Validate a file
csvpp validate input.csvpp

# Validate from stdin
cat input.csvpp | csvpp validate
```

### convert

Convert between CSV++ and other formats (JSON, YAML).

```bash
# CSV++ to JSON
csvpp convert -i input.csvpp -o output.json
csvpp convert -i input.csvpp --to json

# CSV++ to YAML
csvpp convert -i input.csvpp -o output.yaml
csvpp convert -i input.csvpp --to yaml

# JSON to CSV++
csvpp convert -i input.json -o output.csvpp
csvpp convert -i input.json --from json --to csvpp

# YAML to CSV++
csvpp convert -i input.yaml -o output.csvpp
csvpp convert -i input.yaml --from yaml --to csvpp

# Using stdin/stdout
cat input.csvpp | csvpp convert --to json
cat input.json | csvpp convert --from json --to csvpp
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--input` | `-i` | Input file path |
| `--output` | `-o` | Output file path |
| `--from` | | Input format (csvpp, json, yaml) - auto-detected from extension |
| `--to` | | Output format (csvpp, json, yaml) - auto-detected from extension |

### view

View CSV++ file in an interactive TUI table.

```bash
# View a file
csvpp view input.csvpp

# View from stdin
cat input.csvpp | csvpp view
```

**Key Bindings:**

| Key | Action |
|-----|--------|
| `↑` / `↓` | Navigate rows |
| `Space` | Toggle row selection |
| `y` / `c` | Copy header + selected rows to clipboard (CSV++ format) |
| `/` | Open filter input |
| `Enter` | Apply filter (in filter mode) |
| `Esc` | Cancel filter / Clear active filter / Clear selection |
| `q` / `Ctrl+C` | Quit |

**Filter syntax:**
- Type text to search all columns (e.g., `Alice`)
- Use `column:value` to search a specific column (e.g., `name:Alice`)

**Note:** When stdin is not a TTY (e.g., in a pipe), a plain text table is displayed instead of the interactive TUI.

## Examples

```bash
# Validate and convert if valid
csvpp validate data.csvpp && csvpp convert -i data.csvpp -o data.json

# Round-trip conversion
csvpp convert -i data.csvpp -o data.json
csvpp convert -i data.json -o data_restored.csvpp

# Quick preview
csvpp view data.csvpp
```

## License

See the [LICENSE](../../LICENSE) file in the repository root.
