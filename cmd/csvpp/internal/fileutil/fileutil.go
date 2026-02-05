// Package fileutil provides file I/O utilities for the csvpp CLI.
package fileutil

import (
	"fmt"
	"io"
	"os"
)

// OpenInput opens a file for reading or returns stdin if filename is empty.
// The caller must call Close() on the returned ReadCloser.
func OpenInput(filename string) (io.ReadCloser, error) {
	if filename == "" {
		return io.NopCloser(os.Stdin), nil
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return f, nil
}

// OpenInputFromArgs opens input based on command arguments.
// If args is empty, returns stdin. Otherwise opens args[0].
func OpenInputFromArgs(args []string) (io.ReadCloser, error) {
	if len(args) == 0 {
		return io.NopCloser(os.Stdin), nil
	}
	return OpenInput(args[0])
}

// OpenOutput opens a file for writing or returns a WriteCloser wrapping w if filename is empty.
// The caller must call Close() on the returned WriteCloser.
func OpenOutput(filename string, fallback io.Writer) (io.WriteCloser, error) {
	if filename == "" {
		return &nopWriteCloser{fallback}, nil
	}

	f, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	return f, nil
}

// nopWriteCloser wraps an io.Writer with a no-op Close method.
type nopWriteCloser struct {
	io.Writer
}

func (*nopWriteCloser) Close() error { return nil }
