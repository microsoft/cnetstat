package main

import (
	"testing"
)

// This matches the format of 'sudo lsns --json --type net --output ns,pid,command'

var lsnsOutput = []byte(`
{
   "namespaces": [
      {"ns": "23", "pid": "1", "command": "/sbin/init"},
      {"ns": "24", "pid": "96", "command": "/pause"},
      {"ns": "25", "pid": "284", "command": "/pause"}
   ]
}
`)

var lsnsCorrectParse = []NamespaceData{
	NamespaceData{Ns: "23", Pid: "1", Command: "/sbin/init"},
	NamespaceData{Ns: "24", Pid: "96", Command: "/pause"},
	NamespaceData{Ns: "25", Pid: "284", Command: "/pause"},
}

func TestParseLsnsOutput(t *testing.T) {
	namespaces, err := parseLsnsOutput(lsnsOutput)
	if err != nil {
		t.Logf("Got error '%v' from parseLsnsOutput", err)
		t.FailNow()
	}

	if len(namespaces) != len(lsnsCorrectParse) {
		t.Logf("Got %v namespaces from lsnsCorrectParse, expected %v",
			len(namespaces), len(lsnsCorrectParse))
		t.FailNow()
	}

	for i, expected := range lsnsCorrectParse {
		if namespaces[i] != expected {
			t.Errorf("Bad parse of namespace %v: expected %v, got %v",
				i, expected, namespaces[i])
		}
	}
}
