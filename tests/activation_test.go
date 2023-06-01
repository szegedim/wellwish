package tests

import (
	"fmt"
	"gitlab.com/eper.io/engine/burst"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
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
	MainTestLocalPorts.Lock()
	defer MainTestLocalPorts.Unlock()
	defer func() {
		time.Sleep(2 * burst.MaxBurstRuntime)
		burst.FinishCleanup()
		time.Sleep(2 * burst.MaxBurstRuntime)
	}()

	primary := "http://127.0.0.1:7778"
	metadata.NodePattern = "http://127.0.0.1:777*"
	wait := make(chan int)
	nowait := make(chan int)
	// Uncomment this to debug
	// go func(ready chan int) { time.Sleep(2 * time.Second); Main([]string{"go", ":7776"}) }(nowait)
	go func(ready chan int) { time.Sleep(2 * time.Second); runTestServer(t, ready, ":7776", 60*time.Second) }(nowait)
	go func(ready chan int) { time.Sleep(2 * time.Second); runTestServer(t, ready, ":7778", 60*time.Second) }(wait)
	go func(ready chan int) { time.Sleep(2 * time.Second); runTestServer(t, ready, ":7779", 60*time.Second) }(nowait)

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
			time.Sleep(3 * time.Second)
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

	select {
	case <-wait:
		return
	case <-time.After(60 * time.Second):
		t.Error("timeout")
		return
	}
	// nowait never exits in case of local cluster
	burst.FinishCleanup()
}
