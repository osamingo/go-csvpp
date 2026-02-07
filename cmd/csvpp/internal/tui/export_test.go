package tui

import (
	"github.com/charmbracelet/bubbles/table"

	"github.com/osamingo/go-csvpp"
)

// FilterQuery is an exported wrapper of filterQuery for testing.
type FilterQuery struct {
	Column string
	Value  string
}

// ParseFilterQuery exports parseFilterQuery for testing.
func ParseFilterQuery(s string) FilterQuery {
	q := parseFilterQuery(s)
	return FilterQuery{Column: q.column, Value: q.value}
}

// MatchesFilter exports matchesFilter for testing.
func MatchesFilter(query FilterQuery, headers []*csvpp.ColumnHeader, row table.Row) bool {
	return matchesFilter(filterQuery{column: query.Column, value: query.Value}, headers, row)
}
