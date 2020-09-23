package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// A ContainerPath identifies a container in Kubernetes
type ContainerPath struct {
	PodNamespace  string
	PodName       string
	ContainerName string
}

// A DockerContainer connects a container's docker ID and its
// Kubernetes ContainerPath
type DockerContainer struct {
	kubePath ContainerPath
	dockerId string
}

// parseDockerContainerList parses the output of
//    `docker ps --format "{{.ID}} {{.Labels}}"`
// and returns a list of DockerContainers
//
// There will be one Docker container per pod with the special
// container_name 'POD'. This container holds the cgroups for the pod,
// but doesn't correspond to any Kubernetes container.
func parseDockerContainerList(docker_out io.Reader) ([]DockerContainer, error) {
	var result []DockerContainer

	var scanner = bufio.NewScanner(docker_out)
	for scanner.Scan() {
		line := scanner.Text()
		columns := strings.SplitN(line, " ", 2)
		if len(columns) != 2 {
			return nil, fmt.Errorf("Couldn't parse Docker output line %v", line)
		}

		container_id := columns[0]
		labels := strings.Split(columns[1], ",")
		container_path := ContainerPath{}

		for _, label := range labels {
			if strings.Contains(label, "=") {
				parts := strings.SplitN(label, "=", 2)
				if len(parts) != 2 {
					return nil, fmt.Errorf("Couldn't parse container label %v", label)
				}

				key := parts[0]
				value := parts[1]

				switch key {
				case "io.kubernetes.pod.name":
					container_path.PodName = value
				case "io.kubernetes.pod.namespace":
					container_path.PodNamespace = value
				case "io.kubernetes.container.name":
					container_path.ContainerName = value
				}
			}
		}

		result = append(result, DockerContainer{kubePath: container_path,
			dockerId: container_id})
	}

	return result, nil
}

// Build a map from host PIDs to ContainerPaths.
func buildPidMap() (map[int]ContainerPath, error) {
	ctx, _ := context.WithTimeout(context.Background(), subprocessTimeout)
	dockerPsOut, err := exec.CommandContext(ctx, "docker", "ps", "--format", "{{.ID}} {{.Labels}}").Output()
	if err != nil {
		return nil, err
	}

	dockerContainers, err := parseDockerContainerList(strings.NewReader(string(dockerPsOut)))
	if err != nil {
		return nil, err
	}

	pidMap := make(map[int]ContainerPath)

	for _, container := range dockerContainers {
		ctx, _ := context.WithTimeout(context.Background(), subprocessTimeout)
		dockerInspectOut, err := exec.CommandContext(
			ctx, "docker", "inspect", "--format", "{{.State.Pid}}", container.dockerId).Output()
		if err != nil {
			// We expect errors here if a container was
			// deleted between `docker ps` and here.
			continue
		}

		var rootPid int
		_, err = fmt.Fscan(strings.NewReader(string(dockerInspectOut)), &rootPid)
		if err != nil {
			return nil, err
		}

		pidMap[rootPid] = container.kubePath
	}

	return pidMap, nil
}
