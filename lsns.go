package main

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type NamespaceData struct {
	Ns  int
	Pid int
}

var expectedHeaders = []string{
	"NS", "PID",
}

// Parse output from 'lsns --output ns,pid'
func parseLsnsOutput(blob []byte) ([]NamespaceData, error) {
	// Note: lsns has a --json option, so you would think that we
	// would use that, pass the result to Go's builtin JSON
	// decoder, and not have to write a custom parser. Except
	// that, somewhere between util-linux 2.31 and 2.34, the JSON
	// output format changed. The old version prints ns and pid as
	// JSON strings containing ints, and the new one prints them
	// as JSON numbers. The non-JSON output format is the same,
	// which makes it easier to support both versions this way.
	lines := bufio.NewScanner(strings.NewReader(string(blob)))
	var result []NamespaceData

	lines.Scan()
	header := lines.Text()
	if !stringSlicesEqual(strings.Fields(header), expectedHeaders) {
		return nil, fmt.Errorf("Unexpected header of lsns output: %s", header)
	}

	for lines.Scan() {
		var ns, pid int
		_, err := fmt.Sscanf(lines.Text(), "%d %d", &ns, &pid)
		if err != nil {
			return nil, err
		}

		result = append(result, NamespaceData{
			Ns:  ns,
			Pid: pid,
		})
	}

	return result, nil
}

// Run lsns and parse the output.
// NOTE: if not run as root, lsns will succeed, but not necessarily
// return all namespaces
func listNetNamespaces() ([]NamespaceData, error) {
	ctx, _ := context.WithTimeout(context.Background(), subprocessTimeout)

	output, err := exec.CommandContext(ctx, "lsns", "--type", "net", "--output", "ns,pid").Output()
	if err != nil {
		return nil, err
	}

	return parseLsnsOutput(output)

}
