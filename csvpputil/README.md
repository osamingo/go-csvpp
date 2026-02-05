# csvpputil

Utility package for converting CSV++ data to JSON and YAML formats.

## Requirements

This package uses `encoding/json/jsontext` from Go's experimental JSON v2 implementation.

```bash
GOEXPERIMENT=jsonv2 go build ./...
GOEXPERIMENT=jsonv2 go test ./...
```

## Features

- **Streaming JSON output** - Memory-efficient for large files
- **YAML output** - With preserved key order
- **Full CSV++ field type support** - SimpleField, ArrayField, StructuredField, ArrayStructuredField

## API

### Streaming Writers

For large datasets, use streaming writers to minimize memory usage.

#### JSONArrayWriter

```go
w := csvpputil.NewJSONArrayWriter(os.Stdout, headers)

for _, record := range records {
    if err := w.Write(record); err != nil {
        return err
    }
}

if err := w.Close(); err != nil {
    return err
}
```

#### YAMLArrayWriter

```go
w := csvpputil.NewYAMLArrayWriter(os.Stdout, headers)

for _, record := range records {
    if err := w.Write(record); err != nil {
        return err
    }
}

if err := w.Close(); err != nil {
    return err
}
```

**Note:** YAML output is buffered until `Close()` due to go-yaml library constraints.

### Convenience Functions

For small to medium datasets, use these one-shot functions.

#### Marshal Functions

```go
// CSV++ to JSON bytes
jsonBytes, err := csvpputil.MarshalJSON(headers, records)

// CSV++ to YAML bytes
yamlBytes, err := csvpputil.MarshalYAML(headers, records)
```

#### Write Functions

```go
// Write JSON to io.Writer
err := csvpputil.WriteJSON(w, headers, records)

// Write YAML to io.Writer
err := csvpputil.WriteYAML(w, headers, records)
```

## Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/osamingo/go-csvpp"
    "github.com/osamingo/go-csvpp/csvpputil"
)

func main() {
    headers := []*csvpp.ColumnHeader{
        {Name: "name", Kind: csvpp.SimpleField},
        {Name: "age", Kind: csvpp.SimpleField},
    }

    records := [][]*csvpp.Field{
        {{Value: "Alice"}, {Value: "30"}},
        {{Value: "Bob"}, {Value: "25"}},
    }

    // Convert to JSON
    data, err := csvpputil.MarshalJSON(headers, records)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(data))
    // Output: [{"name":"Alice","age":"30"},{"name":"Bob","age":"25"}]

    // Convert to YAML
    yamlData, err := csvpputil.MarshalYAML(headers, records)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Print(string(yamlData))
    // Output:
    // - name: Alice
    //   age: "30"
    // - name: Bob
    //   age: "25"
}
```

## Field Type Mapping

| CSV++ Field Type | JSON Output | YAML Output |
|------------------|-------------|-------------|
| SimpleField | `"value"` | `value` |
| ArrayField | `["a", "b"]` | `- a`<br>`- b` |
| StructuredField | `{"k1": "v1", "k2": "v2"}` | `k1: v1`<br>`k2: v2` |
| ArrayStructuredField | `[{"k": "v"}, ...]` | `- k: v`<br>`- ...` |

## License

See the [LICENSE](../LICENSE) file in the repository root.
