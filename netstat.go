package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

// A connection as returned by netstat (also as seen by the kernel)
type Connection struct {
	protocol        string // Either "tcp" or "tcp6"
	localHost       string // Either an IP address or a hostname
	localPort       string // Either a number or a well-known protocol like "http"
	remoteHost      string // Like localHost
	remotePort      string // Like localPort
	connectionState string // "ESTABLISHED", "TIME_WAIT", etc.
	pid             int    // 0 if unknown. Connections in TIME_WAIT will have a zero pid
}

// Split a netstat address into a host and a port. An address can be
//   hostname:port
//   IPv4addr:port
//   [IPv6addr]:port
// and port can be a number or a string describing a well-known service
// (i.e. 'http' instead of 80)
func hostAndPort(address string) (string, string, error) {
	split := strings.LastIndexByte(address, byte(':'))
	if split == -1 {
		return "", "", fmt.Errorf("No : in address %v", address)
	}

	return address[:split], address[split+1:], nil
}

// Parse the output of 'sudo netstat --tcp --program'
var expectedHeaderFields = []string{
	"Proto", "Recv-Q", "Send-Q", "Local", "Address", "Foreign",
	"Address", "State", "PID/Program", "name"}

func parseNetstatOutput(output io.Reader) ([]Connection, error) {
	lines := bufio.NewScanner(output)

	lines.Scan()
	if lines.Text() != "Active Internet connections (w/o servers)" {
		return nil, fmt.Errorf("Unexpected line 1 of netstat output: %s", lines.Text())
	}

	lines.Scan()
	header := lines.Text()
	if !stringSlicesEqual(strings.Fields(header), expectedHeaderFields) {
		return nil, fmt.Errorf("Unexpected line 2 of netstat output: %s", header)
	}

	var result []Connection
	for lines.Scan() {
		var proto, recv_q, send_q, local_address, remote_address, state, pid_name string
		// There may be extra space-separated groups at the
		// end of the line. Deliberately ignore them.
		n, err := fmt.Sscan(lines.Text(), &proto, &recv_q, &send_q, &local_address,
			&remote_address, &state, &pid_name)
		if n < 7 {
			return nil, fmt.Errorf("Couldn't scan netstat output line: %s", lines.Text())
		}
		if err != nil {
			return nil, err
		}

		localHost, localPort, err := hostAndPort(local_address)
		if err != nil {
			return nil, err
		}
		remoteHost, remotePort, err := hostAndPort(remote_address)
		if err != nil {
			return nil, err
		}

		parts := strings.Split(pid_name, "/")
		// parts[0] will be the pid
		var pid int
		if parts[0] == "-" {
			pid = 0
		} else {
			pid, err = strconv.Atoi(parts[0])
			if err != nil {
				return nil, err
			}
		}

		result = append(result, Connection{
			protocol:        proto,
			localHost:       localHost,
			localPort:       localPort,
			remoteHost:      remoteHost,
			remotePort:      remotePort,
			connectionState: state,
			pid:             pid,
		})
	}

	return result, nil
}

// Get open TCP connections from the namespace of pid, in the format of parseNetstatOutput
func getConnectionsFromNamespace(pid string) ([]Connection, error) {
	ctx, _ := context.WithTimeout(context.Background(), subprocessTimeout)

	netstatOutput, err := exec.CommandContext(ctx, "nsenter", "-t", pid, "-n", "netstat", "--tcp", "--program").Output()
	if err != nil {
		return nil, err
	}

	return parseNetstatOutput(strings.NewReader(string(netstatOutput)))
}
