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

// Convert empty strings to "-". Why? Because that's what netstat does
func emptyToDash(val string) string {
	if len(val) > 0 {
		return val
	} else {
		return "-"
	}
}

// Print a table. The fields will print in the order returned by
// Fielder.Fields(). header is the first row of fields to print, which
// can be used for column titles. Empty fields will be printed as "-".
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
			fmt.Fprintf(w, "%-*s", fieldWidths[i], emptyToDash(field))
		}
		w.WriteString("\n")
	}
	w.Flush()
}

// Print a table as a series of JSON rows, one row per line of
// output. Each row will be a JSON object where property names come
// from header and values come from the row being printed.
func printJsonTable(rows []Fielder, fieldNames []string, f io.Writer) {
	w := bufio.NewWriter(f)

	for _, row := range rows {
		w.WriteString("{")
		for i, field := range row.Fields() {
			// Using fprintf with a buffered writer
			// results in two buffers, but this still
			// seems better than the alternatives.
			fmt.Fprintf(w, "\"%s\": \"%s\"", fieldNames[i], field)
			if i < (len(fieldNames) - 1) {
				w.WriteString(", ")
			}
		}
		w.WriteString("}\n")
	}

	w.Flush()
}
