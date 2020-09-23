package main

import (
	"strings"
	"testing"
)

// Call t.Errorf(message) if a != b
func expectEqual(t *testing.T, a, b interface{}, message string) {
	if a != b {
		t.Errorf(message)
	}
}

func TestHostAndPort(t *testing.T) {
	host, port, _ := hostAndPort("127.0.0.1:234")
	expectEqual(t, host, "127.0.0.1", "Unexpected host from 127.0.0.1:234")
	expectEqual(t, port, "234", "Unexpected port from 127.0.0.1:234")

	host, port, _ = hostAndPort("foo.com:https")
	expectEqual(t, host, "foo.com", "Unexpected host from foo.com:https")
	expectEqual(t, port, "https", "Unexpected port from foo.com:https")

	host, port, _ = hostAndPort("[::16:5]:578")
	expectEqual(t, host, "[::16:5]", "Unexpected host from [::16:5]:578")
	expectEqual(t, port, "578", "Unexpected port from [::16:5]:578")
}

// This should match the output format of 'sudo netstat --tcp --program'

const netstatOutput = `Active Internet connections (w/o servers)
Proto Recv-Q Send-Q Local Address           Foreign Address         State       PID/Program name
tcp        0      0 kube-node-1:2960        10.0.1.2:https          TIME_WAIT   -
tcp        0      0 kube-node-1:9502        10.0.3.4:https          ESTABLISHED 36/abcd
tcp        0      0 kube-node-1:4587        10.0.5.6:8685           TIME_WAIT   -
tcp        0      0 kube-node-1:0178        10.0.7.8:http-alt       TIME_WAIT   -
tcp        0      0 kube-node-1:ssh         10.0.9.10:3920          ESTABLISHED 9486/sshd: user
tcp        0      0 kube-node-1:5639        kube-node-12:http-alt   TIME_WAIT   -
tcp6       0      0 kube-node-1:1234        kube-node-15:9294       TIME_WAIT   -
tcp6       0      0 kube-node-1:9168        [::16:5:3]:298          TIME_WAIT   -`

var netstatExpectedParse = [8]Connection{
	Connection{protocol: "tcp",
		localHost:       "kube-node-1",
		localPort:       "2960",
		remoteHost:      "10.0.1.2",
		remotePort:      "https",
		connectionState: "TIME_WAIT",
		pid:             0},
	Connection{protocol: "tcp",
		localHost:       "kube-node-1",
		localPort:       "9502",
		remoteHost:      "10.0.3.4",
		remotePort:      "https",
		connectionState: "ESTABLISHED",
		pid:             36},
	Connection{protocol: "tcp",
		localHost:       "kube-node-1",
		localPort:       "4587",
		remoteHost:      "10.0.5.6",
		remotePort:      "8685",
		connectionState: "TIME_WAIT",
		pid:             0},
	Connection{protocol: "tcp",
		localHost:       "kube-node-1",
		localPort:       "0178",
		remoteHost:      "10.0.7.8",
		remotePort:      "http-alt",
		connectionState: "TIME_WAIT",
		pid:             0},
	Connection{protocol: "tcp",
		localHost:       "kube-node-1",
		localPort:       "ssh",
		remoteHost:      "10.0.9.10",
		remotePort:      "3920",
		connectionState: "ESTABLISHED",
		pid:             9486},
	Connection{protocol: "tcp",
		localHost:       "kube-node-1",
		localPort:       "5639",
		remoteHost:      "kube-node-12",
		remotePort:      "http-alt",
		connectionState: "TIME_WAIT",
		pid:             0},
	Connection{protocol: "tcp6",
		localHost:       "kube-node-1",
		localPort:       "1234",
		remoteHost:      "kube-node-15",
		remotePort:      "9294",
		connectionState: "TIME_WAIT",
		pid:             0},
	Connection{protocol: "tcp6",
		localHost:       "kube-node-1",
		localPort:       "9168",
		remoteHost:      "[::16:5:3]",
		remotePort:      "298",
		connectionState: "TIME_WAIT",
		pid:             0},
}

func TestParseNetstatOutput(t *testing.T) {
	connections, err := parseNetstatOutput(strings.NewReader(netstatOutput))

	if err != nil {
		t.Logf("Got error %v from parse_netstat_output", err)
		t.FailNow()
	}

	if len(connections) != len(netstatExpectedParse) {
		t.Logf("Got %v connections, expected %v", len(connections), len(netstatExpectedParse))
		t.FailNow()
	}

	for i, expected := range netstatExpectedParse {
		if expected != connections[i] {
			t.Errorf("Got connection %v, expected %v", connections[i], expected)
		}
	}
}
