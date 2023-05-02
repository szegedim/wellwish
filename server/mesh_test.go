package server

import (
	"fmt"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/mesh"
	"os"
	"testing"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func TestMesh(t *testing.T) {
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
	//mesh.Index[metadata.ActivationKey] = metadata.ActivationKey
	//fmt.Println("cluster is activated")

	time.Sleep(15 * time.Second)
	if len(mesh.Index) != 4 {
		t.Error(mesh.Index)
	}

	<-x
	<-y
	// z Never exits
}
