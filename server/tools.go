package server

import (
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

func runTestServer(t *testing.T, ready chan int, port string, timeout time.Duration) {
	goRoot := os.Getenv("GOROOT")
	p := exec.Cmd{
		Dir:  ".",
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
