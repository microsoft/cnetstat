package main

import (
	"strings"
	"testing"
)

// This should match the output format of 'docker ps --format "{{.ID}} {{.Labels}}"'
const dockerPsOutput = `56443455 component=foo,io.kubernetes.pod.name=frontend,io.kubernetes.pod.namespace=my-app,a=b,c=d,io.kubernetes.container.name=fe-server
fab8905c component=bar,io.kubernetes.pod.name=frontend,io.kubernetes.pod.namespace=my-app,x=y,z=w,io.kubernetes.container.name=log-shipper
a01098fd component=baz,io.kubernetes.pod.name=frontend,io.kubernetes.pod.namespace=my-app,g=x,h=j,io.kubernetes.container.name=POD
65323bda component=bot,io.kubernetes.pod.name=backend,io.kubernetes.pod.namespace=my-app,h=k,j=r,io.kubernetes.container.name=be-server`

var parsedOutput = [4]DockerContainer{
	DockerContainer{dockerId: "56443455",
		kubePath: ContainerPath{
			PodName:       "frontend",
			PodNamespace:  "my-app",
			ContainerName: "fe-server"}},
	DockerContainer{dockerId: "fab8905c",
		kubePath: ContainerPath{
			PodName:       "frontend",
			PodNamespace:  "my-app",
			ContainerName: "log-shipper"}},
	DockerContainer{dockerId: "a01098fd",
		kubePath: ContainerPath{
			PodName:       "frontend",
			PodNamespace:  "my-app",
			ContainerName: "POD"}},
	DockerContainer{dockerId: "65323bda",
		kubePath: ContainerPath{
			PodName:       "backend",
			PodNamespace:  "my-app",
			ContainerName: "be-server"}},
}

func TestParseDockerContainerList(t *testing.T) {
	got, err := parseDockerContainerList(strings.NewReader(dockerPsOutput))

	if err != nil {
		t.Logf("Got error %v from GetDockerContainers", err)
		t.FailNow()
	}

	if len(got) != len(parsedOutput) {
		t.Logf("Got %v containers, expected %v", len(got), len(parsedOutput))
		t.FailNow()
	}

	for i, expected := range parsedOutput {
		if expected != got[i] {
			t.Errorf("Mismatched docker output parse")
		}
	}
}
