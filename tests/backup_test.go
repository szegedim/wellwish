package tests

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/metadata"
	"gitlab.com/eper.io/engine/stateful"
	"net/http"
	"testing"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func TestBackup(t *testing.T) {
	mainTestLocalPorts.Lock()
	defer mainTestLocalPorts.Unlock()
	stateful.ContainerIndexLimit = 2
	// Normally this is /var/lib
	metadata.DataRoot = "/tmp"
	metadata.Http11Port = ":7599"
	metadata.SiteUrl = "http://127.0.0.1:7599"
	metadata.StatefulBackupUrl = "http://127.0.0.1:7599"

	module := map[string]string{}
	stateful.RegisterModuleForBackup(&module)

	stateful.SetupStateful()

	go func() {
		_ = http.ListenAndServe(metadata.Http11Port, nil)
	}()

	time.Sleep(1 * time.Second)

	metadata.ManagementKey = ""
	managementKey := drawing.GenerateUniqueKey()
	drawing.NoErrorResponse(http.Post(fmt.Sprintf("%s/setkey?apikey=%s", metadata.StatefulBackupUrl, managementKey), "", nil))
	if metadata.ManagementKey != managementKey {
		t.Error("cannot set key")
	}

	expectedTestResults := map[string]string{}

	go func() {
		time.Sleep(20 * time.Second)
		k := drawing.GenerateUniqueKey()
		v := drawing.GenerateUniqueKey()
		expectedTestResults[k] = v
		stateful.SetStatefulItem(&module, k, v)
	}()
	go func() {
		time.Sleep(7 * time.Second)
		k := drawing.GenerateUniqueKey()
		v := drawing.GenerateUniqueKey()
		expectedTestResults[k] = v
		stateful.SetStatefulItem(&module, k, v)
	}()
	go func() {
		time.Sleep(25 * time.Second)
		k := drawing.GenerateUniqueKey()
		v := drawing.GenerateUniqueKey()
		expectedTestResults[k] = v
		stateful.SetStatefulItem(&module, k, v)
	}()

	started := time.Now()
	for time.Now().Sub(started).Seconds() < 120 {
		time.Sleep(3 * time.Second)

		fmt.Println("Next epoch")
		stateful.RegularCleanup()
		if len(module) > stateful.ContainerIndexLimit {
			t.Error("Cannot delete", len(module))
		}
		for kk, vv := range expectedTestResults {
			vvv := stateful.GetStatefulItem(&module, kk)
			if vv != vvv {
				t.Error(vv, vvv)
			}
			fmt.Println(kk, vvv)
		}
	}
}
