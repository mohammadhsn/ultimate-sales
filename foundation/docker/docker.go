// Package docker provides support for starting and sopping docker containers
// for running tests.
package docker

import (
	"bytes"
	"encoding/json"
	"log"
	"net"
	"os/exec"
	"testing"
)

type Container struct {
	Id   string
	Host string // IP:Port
}

// StartContainer starts the specified container for running tests.
func StartContainer(t *testing.T, image string, port string, args ...string) *Container {
	defaultArgs := []string{"run", "-P", "-d"}
	defaultArgs = append(defaultArgs, args...)
	defaultArgs = append(defaultArgs, image)

	cmd := exec.Command("docker", defaultArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not start container %s: %v", image, err)
	}

	id := out.String()[:12]

	cmd = exec.Command("docker", "container", "inspect", id)
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not inspect container %s: %v", id, err)
	}

	var doc []map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("could not decode json: %v", err)
	}

	ip, randPort := extractIpPort(t, doc, port)

	c := Container{
		Id:   id,
		Host: net.JoinHostPort(ip, randPort),
	}

	t.Logf("Image:       %s", image)
	t.Logf("ContainerId: %s", c.Id)
	t.Logf("Host:        %s", c.Host)

	return &c
}

// StopContainer stops and removes the specified container.
func StopContainer(t *testing.T, id string) {
	if err := exec.Command("docker", "container", "stop", id).Run(); err != nil {
		log.Fatalf("could not stop container: %v", err)
	}
	t.Log("Stopped:", id)

	if err := exec.Command("docker", "container", "rm", id, "-v").Run(); err != nil {
		t.Fatalf("could not remove container: %v", err)
	}
	t.Log("Removed:", id)
}

// DumpContainerLogs outputs logs from the running docker container.
func DumpContainerLogs(t *testing.T, id string) {
	out, err := exec.Command("docker", "container", "logs", id).CombinedOutput()
	if err != nil {
		t.Fatalf("could not log container: %v", err)
	}
	t.Logf("Logs for %s\n%s", id, out)
}

func extractIpPort(t *testing.T, doc []map[string]interface{}, port string) (string, string) {
	nw, exists := doc[0]["NetworkSettings"]
	if !exists {
		t.Fatal("could not get network settings")
	}

	ports, exists := nw.(map[string]interface{})["Ports"]
	if !exists {
		t.Fatal("could not get network ports settings")
	}

	tcp, exists := ports.(map[string]interface{})[port+"/tcp"]
	if !exists {
		t.Fatal("could not get network ports/tcp settings")
	}

	list, exists := tcp.([]interface{})
	if !exists {
		t.Fatal("could not get network ports/tcp list settings")
	}

	var hostIp, hostPort string

	for _, l := range list {
		data, exists := l.(map[string]interface{})
		if !exists {
			t.Fatal("could not get network ports/tcp list data")
		}

		hostIp = data["HostIp"].(string)
		if hostIp != "::" {
			hostIp = data["HostIp"].(string)
		}

		hostPort = data["HostPort"].(string)
	}
	return hostIp, hostPort
}
