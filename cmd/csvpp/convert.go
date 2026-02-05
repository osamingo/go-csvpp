package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/osamingo/go-csvpp"
	"github.com/osamingo/go-csvpp/cmd/csvpp/internal/converter"
	"github.com/osamingo/go-csvpp/cmd/csvpp/internal/fileutil"
	"github.com/osamingo/go-csvpp/csvpputil"
)

// Format represents output format.
type Format string

const (
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatCSVPP Format = "csvpp"
)

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert between CSV++ and JSON/YAML",
	Long: `Convert CSV++ files to JSON/YAML or vice versa.

Examples:
  # Convert CSVPP to JSON
  csvpp convert -i input.csvpp -o output.json
  csvpp convert -i input.csvpp --to json

  # Convert CSVPP to YAML
  csvpp convert -i input.csvpp -o output.yaml

  # Convert JSON to CSVPP
  csvpp convert -i input.json -o output.csvpp

  # Using stdin/stdout
  cat input.csvpp | csvpp convert --to json
  cat input.json | csvpp convert --from json --to csvpp`,
	RunE: runConvert,
}

func init() {
	convertCmd.Flags().StringP("input", "i", "", "input file (reads from stdin if not specified)")
	convertCmd.Flags().StringP("output", "o", "", "output file (writes to stdout if not specified)")
	convertCmd.Flags().String("from", "", "input format when using stdin (json, yaml, csvpp)")
	convertCmd.Flags().String("to", "", "output format (json, yaml, csvpp)")

	rootCmd.AddCommand(convertCmd)
}

func runConvert(cmd *cobra.Command, _ []string) (retErr error) {
	// Get flag values
	inputFile, err := cmd.Flags().GetString("input")
	if err != nil {
		return err
	}
	outputFile, err := cmd.Flags().GetString("output")
	if err != nil {
		return err
	}
	fromFormat, err := cmd.Flags().GetString("from")
	if err != nil {
		return err
	}
	toFormat, err := cmd.Flags().GetString("to")
	if err != nil {
		return err
	}

	// Determine input format
	var inputFormat Format
	if fromFormat != "" {
		inputFormat = Format(strings.ToLower(fromFormat))
	} else {
		inputFormat = detectFormat(inputFile)
	}

	// Determine output format
	var outFormat Format
	if toFormat != "" {
		outFormat = Format(strings.ToLower(toFormat))
	} else if outputFile != "" {
		outFormat = detectFormat(outputFile)
	}
	if outFormat == "" {
		return fmt.Errorf("output format must be specified via --to or output file extension")
	}

	// Infer input format from output format for stdin
	if inputFormat == "" && inputFile == "" {
		// If output is CSVPP, input must be JSON or YAML (default to JSON)
		// Otherwise, input is CSVPP
		if outFormat == FormatCSVPP {
			inputFormat = FormatJSON // Default to JSON when converting to CSVPP from stdin
		} else {
			inputFormat = FormatCSVPP
		}
	}

	// Open input
	r, err := fileutil.OpenInput(inputFile)
	if err != nil {
		return err
	}
	defer r.Close()

	// Open output
	w, err := fileutil.OpenOutput(outputFile, cmd.OutOrStdout())
	if err != nil {
		return err
	}
	defer func() {
		if cerr := w.Close(); cerr != nil && retErr == nil {
			retErr = fmt.Errorf("failed to close output: %w", cerr)
		}
	}()

	// Route to appropriate converter
	switch {
	case inputFormat == FormatCSVPP && (outFormat == FormatJSON || outFormat == FormatYAML):
		return convertFromCSVPP(r, w, outFormat)
	case (inputFormat == FormatJSON || inputFormat == FormatYAML) && outFormat == FormatCSVPP:
		return convertToCSVPP(r, w, inputFormat)
	case inputFormat == outFormat:
		return fmt.Errorf("input and output formats are the same: %s", inputFormat)
	default:
		return fmt.Errorf("unsupported conversion: %s -> %s", inputFormat, outFormat)
	}
}

// detectFormat detects format from file extension.
func detectFormat(filename string) Format {
	if filename == "" {
		return ""
	}
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".json":
		return FormatJSON
	case ".yaml", ".yml":
		return FormatYAML
	case ".csvpp", ".csv":
		return FormatCSVPP
	default:
		return ""
	}
}

// convertFromCSVPP converts CSVPP to JSON or YAML.
func convertFromCSVPP(r io.Reader, w io.Writer, outFormat Format) error {
	reader := csvpp.NewReader(r)

	headers, err := reader.Headers()
	if err != nil {
		return fmt.Errorf("failed to read headers: %w", err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read records: %w", err)
	}

	switch outFormat {
	case FormatJSON:
		return csvpputil.WriteJSON(w, headers, records)
	case FormatYAML:
		return csvpputil.WriteYAML(w, headers, records)
	default:
		return fmt.Errorf("unsupported output format: %s", outFormat)
	}
}

// convertToCSVPP converts JSON or YAML to CSVPP.
func convertToCSVPP(r io.Reader, w io.Writer, inputFormat Format) error {
	var headers []*csvpp.ColumnHeader
	var records [][]*csvpp.Field
	var err error

	switch inputFormat {
	case FormatJSON:
		headers, records, err = converter.FromJSON(r)
	case FormatYAML:
		headers, records, err = converter.FromYAML(r)
	default:
		return fmt.Errorf("unsupported input format: %s", inputFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", inputFormat, err)
	}

	if len(headers) == 0 {
		return fmt.Errorf("no data found in input")
	}

	writer := csvpp.NewWriter(w)
	writer.SetHeaders(headers)

	return writer.WriteAll(records)
}
