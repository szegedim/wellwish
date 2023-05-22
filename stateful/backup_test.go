package stateful

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/metadata"
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
	containerIndexLimit = 2

	SetupStateful()

	go func() {
		_ = http.ListenAndServe(metadata.Http11Port, nil)
	}()

	time.Sleep(1 * time.Second)

	expected := map[string]string{}
	go func() {
		time.Sleep(20 * time.Second)
		k := drawing.GenerateUniqueKey()
		v := drawing.GenerateUniqueKey()
		expected[k] = v
		SetStatefulItem(k, v)
	}()
	go func() {
		time.Sleep(7 * time.Second)
		k := drawing.GenerateUniqueKey()
		v := drawing.GenerateUniqueKey()
		expected[k] = v
		SetStatefulItem(k, v)
	}()
	go func() {
		time.Sleep(25 * time.Second)
		k := drawing.GenerateUniqueKey()
		v := drawing.GenerateUniqueKey()
		expected[k] = v
		SetStatefulItem(k, v)
	}()

	started := time.Now()
	for time.Now().Sub(started).Seconds() < 120 {
		time.Sleep(3 * time.Second)

		fmt.Println("Next epoch")
		if len(csi) > containerIndexLimit {
			t.Error("Cannot delete", len(csi))
		}
		for kk, vv := range expected {
			vvv := GetStatefulItem(kk)
			if vv != vvv {
				t.Error(vv, vvv)
			}
			fmt.Println(kk, vvv)
		}
	}
}

func Clone(data map[string]string) map[string]string {
	ret := make(map[string]string)
	for k, v := range data {
		ret[k] = v
	}
	return ret
}
