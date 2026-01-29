package csvpputil_test

import (
	"fmt"
	"log"
	"os"

	"github.com/osamingo/go-csvpp"
	"github.com/osamingo/go-csvpp/csvpputil"
)

func ExampleRecordToMap() {
	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "tags", Kind: csvpp.ArrayField},
	}

	record := []*csvpp.Field{
		{Value: "Alice"},
		{Values: []string{"go", "rust"}},
	}

	m := csvpputil.RecordToMap(headers, record)
	fmt.Printf("name: %s\n", m["name"])
	fmt.Printf("tags: %v\n", m["tags"])

	// Output:
	// name: Alice
	// tags: [go rust]
}

func ExampleMarshalJSON() {
	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "age", Kind: csvpp.SimpleField},
	}

	records := [][]*csvpp.Field{
		{{Value: "Alice"}, {Value: "30"}},
		{{Value: "Bob"}, {Value: "25"}},
	}

	data, err := csvpputil.MarshalJSON(headers, records, csvpputil.WithDeterministic(true))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))

	// Output:
	// [{"age":"30","name":"Alice"},{"age":"25","name":"Bob"}]
}

func ExampleMarshalYAML() {
	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "age", Kind: csvpp.SimpleField},
	}

	records := [][]*csvpp.Field{
		{{Value: "Alice"}, {Value: "30"}},
		{{Value: "Bob"}, {Value: "25"}},
	}

	data, err := csvpputil.MarshalYAML(headers, records)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(data))

	// Output:
	// - age: "30"
	//   name: Alice
	// - age: "25"
	//   name: Bob
}

func ExampleJSONEncoder() {
	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "score", Kind: csvpp.SimpleField},
	}

	enc := csvpputil.NewJSONEncoder(os.Stdout, headers, csvpputil.WithDeterministic(true))

	_ = enc.Encode([]*csvpp.Field{{Value: "Alice"}, {Value: "100"}})
	_ = enc.Encode([]*csvpp.Field{{Value: "Bob"}, {Value: "85"}})
	_ = enc.Close()

	// Output:
	// [{"name":"Alice","score":"100"},{"name":"Bob","score":"85"}]
}

func ExampleYAMLEncoder() {
	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "score", Kind: csvpp.SimpleField},
	}

	enc := csvpputil.NewYAMLEncoder(os.Stdout, headers)

	_ = enc.Encode([]*csvpp.Field{{Value: "Alice"}, {Value: "100"}})
	_ = enc.Encode([]*csvpp.Field{{Value: "Bob"}, {Value: "85"}})
	_ = enc.Close()

	// Output:
	// - name: Alice
	//   score: "100"
	// - name: Bob
	//   score: "85"
}
