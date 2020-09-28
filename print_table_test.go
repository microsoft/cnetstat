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

var testTable = []Fielder{
	&TestTable{a: "a", b: "b", c: "cc"},
	&TestTable{a: "aaa", b: "b", c: "c"},
	&TestTable{a: "A", b: "", c: "c"},
}

var testFields = []string{"AAA", "B", "C"}

const expectedTable = `AAA  B  C 
a    b  cc
aaa  b  c 
A    -  c 
`

func TestPrettyPrintTable(t *testing.T) {
	var buf bytes.Buffer

	prettyPrintTable(testTable, testFields, &buf)
	written := buf.String()
	if written != expectedTable {
		t.Errorf("prettyPrintTable wrote %#v, expected %#v", written, expectedTable)
	}
}

const expectedJson = `{"AAA": "a", "B": "b", "C": "cc"}
{"AAA": "aaa", "B": "b", "C": "c"}
{"AAA": "A", "B": "", "C": "c"}
`

func testPrintJsonTable(t *testing.T) {
	var buf bytes.Buffer

	printJsonTable(testTable, testFields, &buf)
	written := buf.String()
	if written != expectedJson {
		t.Errorf("printJsonTable wrote %#v, expected %#v", written, expectedJson)
	}
}
