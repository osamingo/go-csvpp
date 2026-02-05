# go-csvpp

[![Go Reference](https://pkg.go.dev/badge/github.com/osamingo/go-csvpp.svg)](https://pkg.go.dev/github.com/osamingo/go-csvpp)
[![Go Report Card](https://goreportcard.com/badge/github.com/osamingo/go-csvpp)](https://goreportcard.com/report/github.com/osamingo/go-csvpp)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go implementation of the [IETF CSV++ specification](https://datatracker.ietf.org/doc/draft-mscaldas-csvpp/) (draft-mscaldas-csvpp-01).

CSV++ extends traditional CSV to support **arrays** and **structured fields** within cells, enabling complex data representation while maintaining CSV's simplicity.

## Features

- Full IETF CSV++ specification compliance
- Wraps `encoding/csv` for RFC 4180 compatibility
- Four field types: Simple, Array, Structured, ArrayStructured
- Struct mapping with `csvpp` tags (Marshal/Unmarshal)
- Configurable delimiters
- Security-conscious design (nesting depth limits)

## Requirements

- Go 1.24 or later

## Installation

```bash
go get github.com/osamingo/go-csvpp
```

## Quick Start

### Reading CSV++ Data

```go
package main

import (
    "fmt"
    "io"
    "strings"

    "github.com/osamingo/go-csvpp"
)

func main() {
    input := `name,phone[],geo(lat^lon)
Alice,555-1234~555-5678,34.0522^-118.2437
Bob,555-9999,40.7128^-74.0060
`

    reader := csvpp.NewReader(strings.NewReader(input))

    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            panic(err)
        }

        name := record[0].Value
        phones := record[1].Values
        lat := record[2].Components[0].Value
        lon := record[2].Components[1].Value

        fmt.Printf("%s: phones=%v, location=(%s, %s)\n", name, phones, lat, lon)
    }
}
```

Output:
```
Alice: phones=[555-1234 555-5678], location=(34.0522, -118.2437)
Bob: phones=[555-9999], location=(40.7128, -74.0060)
```

### Writing CSV++ Data

```go
package main

import (
    "bytes"
    "fmt"

    "github.com/osamingo/go-csvpp"
)

func main() {
    var buf bytes.Buffer
    writer := csvpp.NewWriter(&buf)

    headers := []*csvpp.ColumnHeader{
        {Name: "name", Kind: csvpp.SimpleField},
        {Name: "tags", Kind: csvpp.ArrayField, ArrayDelimiter: '~'},
    }
    writer.SetHeaders(headers)

    if err := writer.WriteHeader(); err != nil {
        panic(err)
    }
    if err := writer.Write([]*csvpp.Field{
        {Value: "Alice"},
        {Values: []string{"go", "rust", "python"}},
    }); err != nil {
        panic(err)
    }
    writer.Flush()

    fmt.Print(buf.String())
}
```

Output:
```
name,tags[]
Alice,go~rust~python
```

### Struct Mapping

```go
package main

import (
    "fmt"
    "strings"

    "github.com/osamingo/go-csvpp"
)

type Person struct {
    Name   string   `csvpp:"name"`
    Phones []string `csvpp:"phone[]"`
    Geo    struct {
        Lat string
        Lon string
    } `csvpp:"geo(lat^lon)"`
}

func main() {
    input := `name,phone[],geo(lat^lon)
Alice,555-1234~555-5678,34.0522^-118.2437
`

    var people []Person
    if err := csvpp.Unmarshal(strings.NewReader(input), &people); err != nil {
        panic(err)
    }

    for _, p := range people {
        fmt.Printf("%s: phones=%v, geo=(%s, %s)\n",
            p.Name, p.Phones, p.Geo.Lat, p.Geo.Lon)
    }
}
```

Output:
```
Alice: phones=[555-1234 555-5678], geo=(34.0522, -118.2437)
```

## Field Types

CSV++ supports four field types in headers:

| Type | Header Syntax | Example Data | Description |
|------|---------------|--------------|-------------|
| Simple | `name` | `Alice` | Plain text value |
| Array | `tags[]` | `go~rust~python` | Multiple values with delimiter |
| Structured | `geo(lat^lon)` | `34.05^-118.24` | Named components |
| ArrayStructured | `addr[](city^zip)` | `LA^90210~NY^10001` | Array of structures |

### Default Delimiters

- Array delimiter: `~` (tilde)
- Component delimiter: `^` (caret)

Custom delimiters can be specified in the header:
- `phone[|]` - uses `|` as array delimiter
- `geo;(lat;lon)` - uses `;` as component delimiter

### Delimiter Progression

For nested structures, the IETF specification recommends:

| Level | Delimiter |
|-------|-----------|
| 1 (arrays) | `~` |
| 2 (components) | `^` |
| 3 | `;` |
| 4 | `:` |

## API Reference

### Reader

```go
reader := csvpp.NewReader(r) // r is io.Reader

// Configuration (same as encoding/csv)
reader.Comma = ','           // Field delimiter
reader.Comment = '#'         // Comment character
reader.LazyQuotes = false    // Relaxed quote handling
reader.TrimLeadingSpace = false
reader.MaxNestingDepth = 10  // Nesting limit (security)

// Methods
headers, err := reader.Headers()  // Get parsed headers
record, err := reader.Read()      // Read one record
records, err := reader.ReadAll()  // Read all records
```

### Writer

```go
writer := csvpp.NewWriter(w) // w is io.Writer

// Configuration
writer.Comma = ','      // Field delimiter
writer.UseCRLF = false  // Use \r\n line endings

// Methods
writer.SetHeaders(headers)  // Set column headers
writer.WriteHeader()        // Write header row
writer.Write(record)        // Write one record
writer.WriteAll(records)    // Write all records
writer.Flush()              // Flush buffer
```

### Marshal/Unmarshal

```go
// Unmarshal CSV++ data into structs
var people []Person
err := csvpp.Unmarshal(reader, &people)

// Marshal structs to CSV++ data
err := csvpp.Marshal(writer, people)
```

### Struct Tags

Use `csvpp` struct tags to map fields:

```go
type Record struct {
    Name     string   `csvpp:"name"`           // Simple field
    Tags     []string `csvpp:"tags[]"`         // Array field
    Location struct {                          // Structured field
        Lat string
        Lon string
    } `csvpp:"geo(lat^lon)"`
    Addresses []Address `csvpp:"addr[](street^city)"` // Array structured
}
```

## JSON/YAML Conversion (csvpputil)

The `csvpputil` package provides utilities for converting CSV++ data to JSON and YAML formats.

### Installation

```bash
go get github.com/osamingo/go-csvpp/csvpputil
```

### Quick Conversion

```go
import "github.com/osamingo/go-csvpp/csvpputil"

// Convert to JSON
jsonData, err := csvpputil.MarshalJSON(headers, records)

// Convert to YAML
yamlData, err := csvpputil.MarshalYAML(headers, records)

// Write directly to io.Writer
err = csvpputil.WriteJSON(w, headers, records)
err = csvpputil.WriteYAML(w, headers, records)
```

### Streaming Output

For large datasets, use streaming writers:

```go
// JSON streaming
w := csvpputil.NewJSONArrayWriter(out, headers)
for _, record := range records {
    if err := w.Write(record); err != nil {
        return err
    }
}
if err := w.Close(); err != nil {
    return err
}

// YAML streaming
w := csvpputil.NewYAMLArrayWriter(out, headers)
for _, record := range records {
    if err := w.Write(record); err != nil {
        return err
    }
}
if err := w.Close(); err != nil {
    return err
}
```

### Example Output

Given CSV++ data:
```
name,tags[],geo(lat^lon)
Alice,go~rust,35.6762^139.6503
```

JSON output:
```json
[{"name":"Alice","tags":["go","rust"],"geo":{"lat":"35.6762","lon":"139.6503"}}]
```

YAML output:
```yaml
- name: Alice
  tags:
  - go
  - rust
  geo:
    lat: "35.6762"
    lon: "139.6503"
```

## Compatibility

This package wraps `encoding/csv` and inherits:
- Full RFC 4180 compliance
- Quoted field handling
- Configurable field/line delimiters
- Comment support

## Security

- **MaxNestingDepth**: Limits nested structure depth (default: 10) to prevent stack overflow from malicious input
- Header names are restricted to ASCII characters per IETF specification

### CSV Injection Prevention

When CSV files are opened in spreadsheet applications, values starting with `=`, `+`, `-`, or `@` may be interpreted as formulas. Use `HasFormulaPrefix` to detect and escape dangerous values:

```go
if csvpp.HasFormulaPrefix(value) {
    value = "'" + value // Escape for spreadsheet safety
}
```

## Specification

This implementation follows the IETF CSV++ specification:
- [draft-mscaldas-csvpp-01](https://datatracker.ietf.org/doc/draft-mscaldas-csvpp/)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
