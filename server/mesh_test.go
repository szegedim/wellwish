package server

import (
	"fmt"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
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
	t.SkipNow()
	_ = os.Chdir("..")
	primary := "http://127.0.0.1:7788"
	metadata.NodePattern = "http://127.0.0.1:778*"
	wait := make(chan int)
	nowait := make(chan int)
	// Uncomment this to debug
	//go func(ready chan int) { time.Sleep(2 * time.Second); Main([]string{"go", ":7784"}) }(nowait)
	//mesh.SetIndex(drawing.GenerateUniqueKey(), drawing.GenerateUniqueKey())
	go func(ready chan int) { time.Sleep(2 * time.Second); runServer(t, ready, ":7786", 60*time.Second) }(wait)
	go func(ready chan int) { time.Sleep(2 * time.Second); runServer(t, ready, ":7788", 60*time.Second) }(nowait)
	go func(ready chan int) { time.Sleep(2 * time.Second); runServer(t, ready, ":7789", 60*time.Second) }(nowait)

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
	if string(ret) != "3" {
		t.Error("something went wrong", mesh.IndexLengthForTestingOnly())
	}
	fmt.Println(string(ret))
	time.Sleep(15 * time.Second)
	ret, _ = management.HttpProxyRequest(englang.Printf("%s/healthz", primary), "", nil)
	fmt.Println(string(ret))
	time.Sleep(15 * time.Second)
	ret, _ = management.HttpProxyRequest(englang.Printf("%s/healthz", primary), "", nil)
	fmt.Println(string(ret))

	<-wait
	time.Sleep(2 * time.Second)
	// nowait never exits in case of local cluster
}
