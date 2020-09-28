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
	"io"
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

// I'm not using the standard json module for JSON output because I
// want to flatten the KubeConnection before printing it
func writeKubeConnectionAsJSON(kc *KubeConnection, w io.Writer) error {
	_, err := fmt.Fprintf(w,
		"{\"protocol\": %v,"+
			"\"local_host\": %v, "+
			"\"local_port\": %v, "+
			"\"remote_host\": %v, "+
			"\"remote_port\": %v, "+
			"\"connection_state\": %v, "+
			"\"pod_namespace\": %v, "+
			"\"pod_name\": %v, "+
			"\"container_name\": %v}",
		kc.conn.protocol,
		kc.conn.localHost,
		kc.conn.localPort,
		kc.conn.remoteHost,
		kc.conn.remotePort,
		kc.conn.connectionState,
		kc.container.PodNamespace,
		kc.container.PodName,
		kc.container.ContainerName)

	return err
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

// Convert empty strings to "-". Why? Because that's what netstat does
func emptyToDash(val string) string {
	if len(val) > 0 {
		return val
	} else {
		return "-"
	}
}

var kubeConnectionHeaders = []string{
	"Namespace", "Pod", "Container", "Protocol",
	"Local Host", "Local Port", "Remote Host", "Remote Port",
	"Connection State",
}

func (kc KubeConnection) Fields() []string {
	return []string{
		emptyToDash(kc.container.PodNamespace),
		emptyToDash(kc.container.PodName),
		emptyToDash(kc.container.ContainerName),
		kc.conn.protocol,
		kc.conn.localHost,
		kc.conn.localPort,
		kc.conn.remoteHost,
		kc.conn.remotePort,
		kc.conn.connectionState,
	}
}

// Parse our arguments. Return the value of the format argument - either "table" or "json"
func parseArgs() (string, error) {
	var format = flag.String("format", "table", "Output format. Either 'table' or 'json'")

	flag.Parse()

	// If we got any positional arguments, that's a user error
	if len(flag.Args()) > 0 {
		flag.Usage()
		return "", fmt.Errorf("got extra arguments %v", flag.Args())
	}

	if (*format != "table") && (*format != "json") {
		flag.Usage()
		return "", fmt.Errorf("unrecognized format %v", format)
	}

	return *format, nil
}

// This is effectively main, but moving it to a separate function
// makes the error handling simpler
func cnetstat() error {
	format, err := parseArgs()
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

	switch format {
	case "json":
		for _, conn := range kubeConnections {
			err := writeKubeConnectionAsJSON(&conn, os.Stdout)
			if err != nil {
				return err
			}
			os.Stdout.WriteString("\n")
		}
	case "table":
		table := make([]Fielder, len(kubeConnections))
		for i, _ := range kubeConnections {
			table[i] = &kubeConnections[i]
		}
		prettyPrintTable(table, kubeConnectionHeaders, os.Stdout)
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
