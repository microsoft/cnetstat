package main

import (
	"testing"
)

// This matches the format of 'sudo lsns --type net --output ns,pid'

var lsnsOutput = []byte(`  NS   PID
2342     1
9694   893
8766  3929
`)

var lsnsCorrectParse = []NamespaceData{
	NamespaceData{Ns: 2342, Pid: 1},
	NamespaceData{Ns: 9694, Pid: 893},
	NamespaceData{Ns: 8766, Pid: 3929},
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
