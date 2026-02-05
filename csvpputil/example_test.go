package csvpputil_test

import (
	"fmt"
	"log"
	"os"

	"github.com/osamingo/go-csvpp"
	"github.com/osamingo/go-csvpp/csvpputil"
)

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
	// [{"name":"Alice","age":"30"},{"name":"Bob","age":"25"}]
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
	// - name: Alice
	//   age: "30"
	// - name: Bob
	//   age: "25"
}

func ExampleJSONArrayWriter() {
	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "score", Kind: csvpp.SimpleField},
	}

	w := csvpputil.NewJSONArrayWriter(os.Stdout, headers, csvpputil.WithDeterministic(true))

	if err := w.Write([]*csvpp.Field{{Value: "Alice"}, {Value: "100"}}); err != nil {
		log.Fatal(err)
	}
	if err := w.Write([]*csvpp.Field{{Value: "Bob"}, {Value: "85"}}); err != nil {
		log.Fatal(err)
	}
	if err := w.Close(); err != nil {
		log.Fatal(err)
	}

	// Output:
	// [{"name":"Alice","score":"100"},{"name":"Bob","score":"85"}]
}

func ExampleYAMLArrayWriter() {
	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "score", Kind: csvpp.SimpleField},
	}

	w := csvpputil.NewYAMLArrayWriter(os.Stdout, headers)

	if err := w.Write([]*csvpp.Field{{Value: "Alice"}, {Value: "100"}}); err != nil {
		log.Fatal(err)
	}
	if err := w.Write([]*csvpp.Field{{Value: "Bob"}, {Value: "85"}}); err != nil {
		log.Fatal(err)
	}
	if err := w.Close(); err != nil {
		log.Fatal(err)
	}

	// Output:
	// - name: Alice
	//   score: "100"
	// - name: Bob
	//   score: "85"
}

func ExampleWriteJSON() {
	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "score", Kind: csvpp.SimpleField},
	}

	records := [][]*csvpp.Field{
		{{Value: "Alice"}, {Value: "100"}},
		{{Value: "Bob"}, {Value: "85"}},
	}

	if err := csvpputil.WriteJSON(os.Stdout, headers, records, csvpputil.WithDeterministic(true)); err != nil {
		log.Fatal(err)
	}

	// Output:
	// [{"name":"Alice","score":"100"},{"name":"Bob","score":"85"}]
}

func ExampleWriteYAML() {
	headers := []*csvpp.ColumnHeader{
		{Name: "name", Kind: csvpp.SimpleField},
		{Name: "score", Kind: csvpp.SimpleField},
	}

	records := [][]*csvpp.Field{
		{{Value: "Alice"}, {Value: "100"}},
		{{Value: "Bob"}, {Value: "85"}},
	}

	if err := csvpputil.WriteYAML(os.Stdout, headers, records); err != nil {
		log.Fatal(err)
	}

	// Output:
	// - name: Alice
	//   score: "100"
	// - name: Bob
	//   score: "85"
}
