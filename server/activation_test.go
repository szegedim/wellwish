package server

import (
	"fmt"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func TestClusterActivation(t *testing.T) {
	_ = os.Chdir("..")
	primary := "http://127.0.0.1:7778"
	metadata.NodePattern = "http://127.0.0.1:777*"
	wait := make(chan int)
	nowait := make(chan int)
	// Uncomment this to debug
	// go func(ready chan int) { time.Sleep(2 * time.Second); Main([]string{"go", ":7776"}) }(nowait)
	go func(ready chan int) { time.Sleep(2 * time.Second); runServer(t, ready, ":7776", 60*time.Second) }(nowait)
	go func(ready chan int) { time.Sleep(2 * time.Second); runServer(t, ready, ":7778", 60*time.Second) }(wait)
	go func(ready chan int) { time.Sleep(2 * time.Second); runServer(t, ready, ":7779", 60*time.Second) }(nowait)

	// Wait for a stable state
	for {
		_, err := management.HttpProxyRequest(englang.Printf("%s/health", primary), "", nil)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	time.Sleep(1 * time.Second)
	fmt.Println("Cluster stable.")
	managementKeyBytes, err := management.HttpProxyRequest(englang.Printf("%s/activate?apikey=%s", primary, metadata.ActivationKey), "", nil)
	managementKey := string(managementKeyBytes)
	fmt.Println("Management key", englang.Printf("%s/management.html?apikey=%s", primary, managementKey))
	fmt.Println("Cluster activation requested.")

	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		managementKeyBytes, err = management.HttpProxyRequest(englang.Printf("http://127.0.0.1:7776/", management.GetAdminKey()), "", nil)
		if err != nil {
			t.Error(err, string(managementKeyBytes))
		}
		if strings.Contains(string(managementKeyBytes), "activate.html") {
			continue
		} else {
			fmt.Println("Activated in ", i, "seconds.")
			break
		}
	}

	ret, err := management.HttpProxyRequest(englang.Printf("http://127.0.0.1:7776/", managementKey), "", nil)
	if err != nil {
		t.Error(err, string(ret))
	}
	if strings.Contains(string(ret), "activate.html") {
		t.Errorf("Server not activated")
	}
	ret, err = management.HttpProxyRequest(englang.Printf("http://127.0.0.1:7778/", managementKey), "", nil)
	if err != nil {
		t.Error(err, string(ret))
	}
	if strings.Contains(string(ret), "activate.html") {
		t.Errorf("Server not activated")
	}
	ret, err = management.HttpProxyRequest(englang.Printf("http://127.0.0.1:7779/", managementKey), "", nil)
	if err != nil {
		t.Error(err)
	}
	if strings.Contains(string(ret), "activate.html") {
		t.Errorf("Server not activated")
	}

	<-wait
	// nowait is not waited for
	time.Sleep(1 * time.Second)
}

func runServer(t *testing.T, ready chan int, port string, timeout time.Duration) {
	p := exec.Cmd{
		Dir:  ".",
		Path: "/Users/miklos_szegedi/schmied.us/private/go-darwin-arm64-bootstrap/bin/go",
		Args: []string{"go", "run", "main.go", port},
	}
	err := p.Start()
	if err != nil {
		t.Error(err)
	}
	go func() {
		time.Sleep(timeout)
		_ = p.Process.Kill()
	}()
	err = p.Wait()
	if err != nil && err.Error() != "signal: killed" {
		t.Error(err)
	}
	b, _ := p.CombinedOutput()
	if len(b) > 0 {
		t.Log(string(b))
	}
	if p.ProcessState.ExitCode() != 0 && p.ProcessState.ExitCode() != -1 {
		t.Log(p.ProcessState.ExitCode())
	}
	ready <- 1
}
