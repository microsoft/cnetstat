package main

import (
	"bytes"
	"testing"
)

type TestTable struct {
	a string
	b string
	c string
}

func (t TestTable) Fields() []string {
	return []string{t.a, t.b, t.c}
}

const expectedTable = `AAA  B  C 
a    b  cc
aaa  b  c 
`

func TestPrettyPrintTable(t *testing.T) {
	var buf bytes.Buffer

	table := []Fielder{
		&TestTable{a: "a", b: "b", c: "cc"},
		&TestTable{a: "aaa", b: "b", c: "c"},
	}

	prettyPrintTable(table, []string{"AAA", "B", "C"}, &buf)
	written := buf.String()
	if written != expectedTable {
		t.Errorf("prettyPrintTable wrote %#v, expected %#v", written, expectedTable)
	}
}
