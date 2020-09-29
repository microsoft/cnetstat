package main

// A container-aware netstat.
//
// Dump a list of connections on a node, with their container and pod
// names. This currently assumes the containers run on docker with labels
// matching what my version of Kubelet does.

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// A connection with a Kubernetes pod identifier instead of a PID
type KubeConnection struct {
	conn      Connection
	container ContainerPath
}

const subprocessTimeout = 5 * time.Second

const ppidColon string = "PPid:"

// Either return the parent PID of its argument, or an error
func parentOfPid(pid int) (int, error) {
	fp, err := os.Open(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0, err
	}
	defer fp.Close()

	lines := bufio.NewScanner(fp)
	for lines.Scan() {
		line := lines.Text()
		if strings.HasPrefix(line, ppidColon) {
			pid, err := strconv.Atoi(line[len(ppidColon):])
			if err != nil {
				return 0, err
			}

			return pid, nil
		}
	}

	return 0, fmt.Errorf("Couldn't find parent PID of PID %d", pid)
}

// Find the container a particular PID runs in, or return an error
func pidToPod(pid int, pidMap map[int]ContainerPath) (ContainerPath, error) {
	// Remember the ancestors of this PID in case we have to
	// search a process hierarchy
	var ancestors []int
	for {
		kube_path, ok := pidMap[pid]
		if ok {
			// If we had to search for parents of the
			// original pid, update the map so we won't
			// have to do that again
			for process, _ := range ancestors {
				pidMap[process] = kube_path
			}

			return kube_path, nil
		}

		ancestors = append(ancestors, pid)
		var err error
		pid, err = parentOfPid(pid)
		if err != nil {
			return ContainerPath{}, err
		}
	}
}

// Map connections with PIDs into KubeConnections with container identifiers
func getKubeConnections(connections []Connection, pidMap map[int]ContainerPath) []KubeConnection {
	kubeConnections := make([]KubeConnection, len(connections))

	for i, conn := range connections {
		pid := conn.pid
		path, _ := pidToPod(pid, pidMap)
		// If pidToPod returns an error, then path will be
		// ContainerPath{}, which is what we want

		kubeConnections[i] = KubeConnection{
			conn:      conn,
			container: path,
		}
	}

	return kubeConnections
}

// Like the TCP 4-tuple, but with a ContainerPath for the local side
type KubeConnectionId struct {
	container  ContainerPath
	remoteHost string
	remotePort string
}

type ConnectionCount struct {
	connId KubeConnectionId
	count  int
}

func summarizeKubeConnections(connections []KubeConnection) []ConnectionCount {
	stats := make(map[KubeConnectionId]int)

	for _, conn := range connections {
		connId := KubeConnectionId{container: conn.container,
			remoteHost: conn.conn.remoteHost,
			remotePort: conn.conn.remotePort}
		count, ok := stats[connId]

		if ok {
			stats[connId] = count + 1
		} else {
			stats[connId] = 1
		}
	}

	result := make([]ConnectionCount, len(stats))
	index := 0
	for k, v := range stats {
		result[index] = ConnectionCount{k, v}
		index += 1
	}

	return result
}

var connectionStatFields = []string{
	"Namespace", "Pod", "Container", "Remote Host", "Remote Port", "Count",
}

func (cc ConnectionCount) Fields() []string {
	return []string{
		cc.connId.container.PodNamespace,
		cc.connId.container.PodName,
		cc.connId.container.ContainerName,
		cc.connId.remoteHost,
		cc.connId.remotePort,
		strconv.Itoa(cc.count),
	}
}

var kubeConnectionHeaders = []string{
	"Namespace", "Pod", "Container", "Protocol",
	"Local Host", "Local Port", "Remote Host", "Remote Port",
	"Connection State",
}

func (kc KubeConnection) Fields() []string {
	return []string{
		kc.container.PodNamespace,
		kc.container.PodName,
		kc.container.ContainerName,
		kc.conn.protocol,
		kc.conn.localHost,
		kc.conn.localPort,
		kc.conn.remoteHost,
		kc.conn.remotePort,
		kc.conn.connectionState,
	}
}

// CnetstatConfig holds our command-line arguments
type CnetstatConfig struct {
	outputFormat string // Either "table" or "json"
	summaryStats bool
}

// Parse our arguments
func parseArgs() (CnetstatConfig, error) {
	var config CnetstatConfig

	flag.StringVar(&config.outputFormat, "format", "table", "Output format. Either 'table' or 'json'")
	flag.BoolVar(&config.summaryStats, "summaryStatistics", true, "Print summary statistics rather than all connections")

	flag.Parse()

	// If we got any positional arguments, that's a user error
	if len(flag.Args()) > 0 {
		flag.Usage()
		return config, fmt.Errorf("got extra arguments %v", flag.Args())
	}

	if (config.outputFormat != "table") && (config.outputFormat != "json") {
		flag.Usage()
		return config, fmt.Errorf("unrecognized format %v", config.outputFormat)
	}

	return config, nil
}

// This is effectively main, but moving it to a separate function
// makes the error handling simpler
func cnetstat() error {
	config, err := parseArgs()
	if err != nil {
		return err
	}

	// It would be possible to run as non-root and return less
	// information, but that makes the netstat parsing more
	// complicated (since netstat will also print a warning
	// message), and for our use-case we really want all the data,
	// so just run it as root.
	if os.Geteuid() != 0 {
		return fmt.Errorf("cnetstat must run as root")
	}

	namespaces, err := listNetNamespaces()
	if err != nil {
		return err
	}

	pidMap, err := buildPidMap()
	if err != nil {
		return err
	}

	// connections has one slice of Connections for each namespace
	var connections = make([][]Connection, len(namespaces))
	for i, namespace := range namespaces {
		conns, err := getConnectionsFromNamespace(strconv.Itoa(namespace.Pid))
		if err != nil {
			return err
		}

		connections[i] = conns
	}

	// count the total number of connections, so we can ...
	var totalConnections int
	for _, conns := range connections {
		totalConnections += len(conns)
	}

	// ... flatten them into a single slice of all connections
	// with just one allocation
	allConnections := make([]Connection, totalConnections)
	offset := 0
	for _, conns := range connections {
		copy(allConnections[offset:], conns)
		offset += len(conns)
	}

	kubeConnections := getKubeConnections(allConnections, pidMap)
	println("Got", len(kubeConnections), "kubeConnections")

	var table []Fielder
	var columns []string
	if config.summaryStats {
		stats := summarizeKubeConnections(kubeConnections)
		table = make([]Fielder, len(stats))
		for i, _ := range stats {
			table[i] = &stats[i]
		}
		columns = connectionStatFields
	} else {
		table = make([]Fielder, len(kubeConnections))
		for i, _ := range kubeConnections {
			table[i] = &kubeConnections[i]
		}
		columns = kubeConnectionHeaders
	}

	switch config.outputFormat {
	case "json":
		printJsonTable(table, columns, os.Stdout)
	case "table":
		prettyPrintTable(table, columns, os.Stdout)
	}

	return nil
}

func main() {
	err := cnetstat()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
