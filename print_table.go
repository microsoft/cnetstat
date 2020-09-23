package main

import (
	"bufio"
	"fmt"
	"io"
)

type Fielder interface {
	// Return the fields of this object, as strings
	Fields() []string
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

// Print a table. The fields will print in the order returned by
// Fielder.Fields(). header is the first row of fields to print, which
// can be used for column titles.
func prettyPrintTable(rows []Fielder, header []string, f io.Writer) {
	w := bufio.NewWriter(f)

	fieldWidths := make([]int, len(header))
	for i, field := range header {
		fieldWidths[i] = len(field)
	}

	// Get the widths of all fields
	for _, row := range rows {
		for i, field := range row.Fields() {
			fieldWidths[i] = max(fieldWidths[i], len(field))
		}
	}

	// Add 2 spaces in between columns
	for i := 0; i < len(fieldWidths)-1; i++ {
		fieldWidths[i] += 2
	}

	// Print the table, with appropriate spacing
	for i, field := range header {
		// Write field, left-padded to width fieldWidths[i]
		fmt.Fprintf(w, "%-*s", fieldWidths[i], field)
	}
	w.WriteString("\n")
	for _, row := range rows {
		for i, field := range row.Fields() {
			fmt.Fprintf(w, "%-*s", fieldWidths[i], field)
		}
		w.WriteString("\n")
	}
	w.Flush()
}
