package tests

import (
	"fmt"
	"gitlab.com/eper.io/engine/burst"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Add a few index entries and check whether they are propagated through the cluster.

func TestMesh(t *testing.T) {
	time.Sleep(100 * time.Millisecond)
	mainTestLocalPorts.Lock()
	defer mainTestLocalPorts.Unlock()
	defer func() {
		time.Sleep(2 * burst.MaxBurstRuntime)
		burst.FinishCleanup()
		time.Sleep(2 * burst.MaxBurstRuntime)
	}()

	primary := "http://127.0.0.1:7724"
	metadata.NodePattern = "http://127.0.0.1:772*"
	wait := make(chan int)
	nowait := make(chan int)
	// Uncomment this to debug
	//go func(ready chan int) { time.Sleep(2 * time.Second); Main([]string{"go", ":7724"}) }(nowait)
	//mesh.SetIndex(drawing.GenerateUniqueKey(), drawing.GenerateUniqueKey())
	go func(ready chan int) { time.Sleep(2 * time.Second); runTestServer(t, ready, ":7724", 60*time.Second) }(wait)
	go func(ready chan int) { time.Sleep(2 * time.Second); runTestServer(t, ready, ":7728", 60*time.Second) }(nowait)
	go func(ready chan int) { time.Sleep(2 * time.Second); runTestServer(t, ready, ":7729", 60*time.Second) }(nowait)

	// Wait for a stable state
	for {
		_, err := management.HttpProxyRequest(englang.Printf("%s/health", primary), "", nil)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	time.Sleep(15 * time.Second)
	fmt.Println("cluster is stable.")

	ret, _ := management.HttpProxyRequest(englang.Printf("%s/healthz", primary), "", nil)
	if englang.Decimal(string(ret)) < englang.Decimal("4") {
		//TODO Check why this is eight sometimes
		t.Error("something went wrong", string(ret))
	}
	time.Sleep(15 * time.Second)
	ret, _ = management.HttpProxyRequest(englang.Printf("%s/healthz", primary), "", nil)
	if englang.Decimal(string(ret)) < englang.Decimal("4") {
		t.Error("something went wrong", string(ret))
	}
	time.Sleep(15 * time.Second)
	ret, _ = management.HttpProxyRequest(englang.Printf("%s/healthz", primary), "", nil)
	if englang.Decimal(string(ret)) < englang.Decimal("4") {
		t.Error("something went wrong", string(ret))
	}

	burst.FinishCleanup()
	select {
	case <-wait:
		return
	case <-time.After(60 * time.Second):
		t.Error("timeout")
		return
	}
	// nowait never exits in case of local cluster
}

func runTestServer(t *testing.T, ready chan int, port string, timeout time.Duration) {
	goRoot := os.Getenv("GOROOT")
	p := exec.Cmd{
		Dir:  "../",
		Path: path.Join(goRoot, "bin", "go"),
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
