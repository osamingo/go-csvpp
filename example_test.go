package csvpp_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/osamingo/go-csvpp"
)

func Example() {
	input := `name,phone[],geo(lat^lon)
Alice,555-1234~555-5678,34.0522^-118.2437
Bob,555-9999,40.7128^-74.0060
`

	reader := csvpp.NewReader(strings.NewReader(input))

	// Get headers
	headers, err := reader.Headers()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Headers: %s, %s, %s\n", headers[0].Name, headers[1].Name, headers[2].Name)

	// Read all records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		name := record[0].Value
		phones := record[1].Values
		lat := record[2].Components[0].Value
		lon := record[2].Components[1].Value

		fmt.Printf("%s: phones=%v, location=(%s, %s)\n", name, phones, lat, lon)
	}

	// Output:
	// Headers: name, phone, geo
	// Alice: phones=[555-1234 555-5678], location=(34.0522, -118.2437)
	// Bob: phones=[555-9999], location=(40.7128, -74.0060)
}

func ExampleReader_Headers() {
	input := `id,name,tags[],address(street^city^zip)
1,Alice,go~rust,123 Main^LA^90210
`

	reader := csvpp.NewReader(strings.NewReader(input))
	headers, err := reader.Headers()
	if err != nil {
		log.Fatal(err)
	}

	for _, h := range headers {
		fmt.Printf("%s: %s\n", h.Name, h.Kind)
	}

	// Output:
	// id: SimpleField
	// name: SimpleField
	// tags: ArrayField
	// address: StructuredField
}

func ExampleReader_Read() {
	input := `name,scores[]
Alice,100~95~88
Bob,77~82
`

	reader := csvpp.NewReader(strings.NewReader(input))

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s: %v\n", record[0].Value, record[1].Values)
	}

	// Output:
	// Alice: [100 95 88]
	// Bob: [77 82]
}

func ExampleReader_ReadAll() {
	input := `name,age
Alice,30
Bob,25
Charlie,35
`

	reader := csvpp.NewReader(strings.NewReader(input))
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Read %d records\n", len(records))
	for _, record := range records {
		fmt.Printf("%s is %s years old\n", record[0].Value, record[1].Value)
	}

	// Output:
	// Read 3 records
	// Alice is 30 years old
	// Bob is 25 years old
	// Charlie is 35 years old
}

func ExampleWriter() {
	var buf bytes.Buffer
	writer := csvpp.NewWriter(&buf)

	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "tags", Kind: csvpp.ArrayField, ArrayDelimiter: '~'},
	}
	writer.SetHeaders(headers)

	if err := writer.WriteHeader(); err != nil {
		log.Fatal(err)
	}

	records := [][]*csvpp.Field{
		{{Value: "Alice"}, {Values: []string{"go", "rust"}}},
		{{Value: "Bob"}, {Values: []string{"python"}}},
	}

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			log.Fatal(err)
		}
	}
	writer.Flush()

	fmt.Print(buf.String())

	// Output:
	// name,tags[]
	// Alice,go~rust
	// Bob,python
}

func ExampleWriter_WriteAll() {
	var buf bytes.Buffer
	writer := csvpp.NewWriter(&buf)

	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "score", Kind: csvpp.SimpleField},
	}
	writer.SetHeaders(headers)

	records := [][]*csvpp.Field{
		{{Value: "Alice"}, {Value: "100"}},
		{{Value: "Bob"}, {Value: "95"}},
	}

	if err := writer.WriteAll(records); err != nil {
		log.Fatal(err)
	}

	fmt.Print(buf.String())

	// Output:
	// name,score
	// Alice,100
	// Bob,95
}

// Person represents a person with contact information.
type Person struct {
	Name   string   `csvpp:"name"`
	Phones []string `csvpp:"phone[]"`
}

func ExampleUnmarshal() {
	input := `name,phone[]
Alice,555-1234~555-5678
Bob,555-9999
`

	var people []Person
	if err := csvpp.Unmarshal(strings.NewReader(input), &people); err != nil {
		log.Fatal(err)
	}

	for _, p := range people {
		fmt.Printf("%s: %v\n", p.Name, p.Phones)
	}

	// Output:
	// Alice: [555-1234 555-5678]
	// Bob: [555-9999]
}

func ExampleMarshal() {
	people := []Person{
		{Name: "Alice", Phones: []string{"555-1234", "555-5678"}},
		{Name: "Bob", Phones: []string{"555-9999"}},
	}

	var buf bytes.Buffer
	if err := csvpp.Marshal(&buf, people); err != nil {
		log.Fatal(err)
	}

	fmt.Print(buf.String())

	// Output:
	// name,phone[]
	// Alice,555-1234~555-5678
	// Bob,555-9999
}

// Location represents a geographic location.
type Location struct {
	Name string `csvpp:"name"`
	Geo  struct {
		Lat string
		Lon string
	} `csvpp:"geo(lat^lon)"`
}

func ExampleUnmarshal_structured() {
	input := `name,geo(lat^lon)
Los Angeles,34.0522^-118.2437
New York,40.7128^-74.0060
`

	var locations []Location
	if err := csvpp.Unmarshal(strings.NewReader(input), &locations); err != nil {
		log.Fatal(err)
	}

	for _, loc := range locations {
		fmt.Printf("%s: (%s, %s)\n", loc.Name, loc.Geo.Lat, loc.Geo.Lon)
	}

	// Output:
	// Los Angeles: (34.0522, -118.2437)
	// New York: (40.7128, -74.0060)
}

func ExampleNewReader_customDelimiter() {
	// Using semicolon as field delimiter (common in European locales)
	input := `name;age
Alice;30
Bob;25
`

	reader := csvpp.NewReader(strings.NewReader(input))
	reader.Comma = ';'

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, record := range records {
		fmt.Printf("%s is %s\n", record[0].Value, record[1].Value)
	}

	// Output:
	// Alice is 30
	// Bob is 25
}
