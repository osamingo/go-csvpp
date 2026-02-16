package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/osamingo/go-csvpp"
	"github.com/osamingo/go-csvpp/cmd/csvpp/internal/fileutil"
)

var validateCmd = &cobra.Command{
	Use:   "validate [file]",
	Short: "Validate CSV++ syntax",
	Long:  `Validate CSV++ file syntax. Reads from file or stdin if no file is specified.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) (retErr error) {
	r, err := fileutil.OpenInputFromArgs(args)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := r.Close(); cerr != nil && retErr == nil {
			retErr = fmt.Errorf("failed to close input: %w", cerr)
		}
	}()

	reader := csvpp.NewReader(r)

	// Read and validate headers
	if _, err := reader.Headers(); err != nil {
		return fmt.Errorf("header validation failed: %w", err)
	}

	// Read and validate all records
	recordCount := 0
	for {
		_, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("record validation failed: %w", err)
		}
		recordCount++
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Valid CSV++ file with %d record(s)\n", recordCount) //nolint:errcheck // stdout write error is not actionable
	return nil
}
