package main

import (
	"context"
	"encoding/json"
	"os/exec"
)

type NamespaceData struct {
	Ns      string
	Pid     string
	Command string
}

// Parse output from 'lsns --json --output ns,pid,command'
func parseLsnsOutput(blob []byte) ([]NamespaceData, error) {
	var dummy map[string][]NamespaceData

	err := json.Unmarshal(blob, &dummy)
	if err != nil {
		return nil, err
	}

	// The map should always contain a single element, with key
	// "namespaces"
	return dummy["namespaces"], nil
}

// Run lsns and parse the output.
// NOTE: if not run as root, lsns will succeed, but not necessarily
// return all namespaces
func listNetNamespaces() ([]NamespaceData, error) {
	ctx, _ := context.WithTimeout(context.Background(), subprocessTimeout)

	output, err := exec.CommandContext(ctx, "lsns", "--json", "--type", "net", "--output", "ns,pid,command").Output()
	if err != nil {
		return nil, err
	}

	return parseLsnsOutput(output)

}
