package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/osamingo/go-csvpp"
	"github.com/osamingo/go-csvpp/cmd/csvpp/internal/fileutil"
	"github.com/osamingo/go-csvpp/cmd/csvpp/internal/tui"
)

var viewCmd = &cobra.Command{
	Use:   "view [file]",
	Short: "View CSV++ file in a table",
	Long: `View CSV++ file contents in an interactive table.

Uses a TUI when running in a terminal, falls back to plain text output
when piped or not in a TTY.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runView,
}

func init() {
	rootCmd.AddCommand(viewCmd)
}

func runView(cmd *cobra.Command, args []string) error {
	r, err := fileutil.OpenInputFromArgs(args)
	if err != nil {
		return err
	}
	defer r.Close()

	reader := csvpp.NewReader(r)

	headers, err := reader.Headers()
	if err != nil {
		return fmt.Errorf("failed to read headers: %w", err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read records: %w", err)
	}

	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		// Plain text output for pipes
		fmt.Fprint(cmd.OutOrStdout(), tui.PlainView(headers, records))
		return nil
	}

	// Interactive TUI
	model := tui.NewModel(headers, records)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}

	return nil
}
