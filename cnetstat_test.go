package main

import (
	"testing"
)

func TestSummarizeEmpty(t *testing.T) {
	empty := make([]KubeConnection, 0)
	val := summarizeKubeConnections(empty)

	if len(val) != 0 {
		t.Errorf("Expected no stats, got %v", val)
	}
}

func TestSummarize(t *testing.T) {
	// Two connections each to two different remote endpoints
	conns := []Connection{
		Connection{
			protocol: "tcp",
			localHost: "127.0.0.1",
			localPort: "https",
			remoteHost: "10.0.5.9",
			remotePort: "5086",
			connectionState: "ESTABLISHED",
			pid: 42,
		},
		Connection{
			protocol: "tcp",
			localHost: "127.0.0.1",
			localPort: "5069",
			remoteHost: "10.0.5.9",
			remotePort: "5086",
			connectionState: "TIME_WAIT",
			pid: 85,
		},
		Connection{
			protocol: "tcp6",
			localHost: "127.0.0.1",
			localPort: "1234",
			remoteHost: "10.0.3.4",
			remotePort: "6230",
			connectionState: "ESTABLISHED",
			pid: 42,
		},
		Connection{
			protocol: "tcp6",
			localHost: "127.0.0.1",
			localPort: "5982",
			remoteHost: "10.0.3.4",
			remotePort: "6230",
			connectionState: "ESTABLISHED",
			pid: 85,
		},
	}

	// The first two connections are from the same container. The
	// second two are from different containers.
	kubeConns := []KubeConnection{
		KubeConnection{
			conn: conns[0],
			container: ContainerPath{
				PodNamespace: "myapp",
				PodName: "frontend",
				ContainerName: "fe-server",
			},
		},
		KubeConnection{
			conn: conns[1],
			container: ContainerPath{
				PodNamespace: "myapp",
				PodName: "frontend",
				ContainerName: "fe-server",
			},
		},
		KubeConnection{
			conn: conns[2],
			container: ContainerPath{
				PodNamespace: "myapp",
				PodName: "frontend",
				ContainerName: "fe-server",
			},
		},
		KubeConnection{
			conn: conns[3],
			container: ContainerPath{
				PodNamespace: "myapp",
				PodName: "frontend",
				ContainerName: "log-shipper",
			},
		},
	}

	expectedStats := []ConnectionCount{
			ConnectionCount{
				connId: KubeConnectionId{
					container: ContainerPath{
						PodNamespace: "myapp",
						PodName: "frontend",
						ContainerName: "fe-server",
					},
					remoteHost: "10.0.5.9",
					remotePort: "5086",
				},
				count: 2,
			},
			ConnectionCount{
				connId: KubeConnectionId{
					container: ContainerPath{
						PodNamespace: "myapp",
						PodName: "frontend",
						ContainerName: "fe-server",
					},
					remoteHost: "10.0.3.4",
					remotePort: "6230",
				},
				count: 1,
			},
			ConnectionCount{
				connId: KubeConnectionId{
					container: ContainerPath{
						PodNamespace: "myapp",
						PodName: "frontend",
						ContainerName: "log-shipper",
					},
					remoteHost: "10.0.3.4",
					remotePort: "6230",
				},
				count: 1,
			},
		}

	stats := summarizeKubeConnections(kubeConns)

	// The order of stats is unspecified (it depends on Go's map
	// traversal order), so we just walk the array and keep a
	// count of how many of our expected statistics we've actually
	// seen.

	// Entries in stats that we expected to see
	statWasExpected := make([]bool, len(stats))
	// Entries in expectedOutput that we saw in stats
	expectedWasSeen := make([]bool, 3)

	for i, stat := range stats {
		for j, expected := range expectedStats {
			if stat == expected {
				statWasExpected[i] = true
				expectedWasSeen[j] = true
			}
		}
	}

	for i, expected := range statWasExpected {
		if !expected {
			t.Errorf("Unexpected stat: %v\n", stats[i])
		}
	}

	for i, seen := range expectedWasSeen {
		if !seen {
			t.Errorf("Didn't see expected stat: %v\n", expectedStats[i])
		}
	}
}
