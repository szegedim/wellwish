package server

import (
	"fmt"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"os"
	"os/exec"
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
	x := make(chan int)
	y := make(chan int)
	z := make(chan int)
	go func(ready chan int) { time.Sleep(2 * time.Second); Main([]string{"go", ":7777"}) }(z)
	go func(ready chan int) { time.Sleep(2 * time.Second); runServer(t, ready, ":7778") }(y)
	go func(ready chan int) { time.Sleep(2 * time.Second); runServer(t, ready, ":7779") }(x)

	// Wait for a stable state
	for {
		_, err := management.HttpProxyRequest(englang.Printf("http://127.0.0.1:7777/healthz"), "", nil)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("cluster is stable")
	mesh.Index[metadata.ActivationKey] = metadata.ActivationKey
	fmt.Println("cluster is activated")

	t.Log(billing.IssueVouchers(
		drawing.GenerateUniqueKey(), "100",
		"Example Inc.", "1 First Ave, USA",
		"hq@opensource.eper.io", "USD 3"))

	time.Sleep(15 * time.Second)
	ret, err := management.HttpProxyRequest(englang.Printf("http://127.0.0.1:7777/management.html?apikey=%s", management.GetAdminKey()), "", nil)
	if err != nil {
		t.Error(err)
	}
	t.Log("management", string(ret))

	time.Sleep(15 * time.Second)
	if len(mesh.Index) != 5 {
		t.Error(mesh.Index)
	}

	<-x
	<-y
	// z Never exits
}

func runServer(t *testing.T, ready chan int, port string) {
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
		time.Sleep(60 * time.Second)
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
	if p.ProcessState.ExitCode() != 0 {
		t.Log(p.ProcessState.ExitCode())
	}
	ready <- 1
}
