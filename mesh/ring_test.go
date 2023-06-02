package mesh

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"strings"
	"testing"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func TestRing(t *testing.T) {
	// TODO
	t.SkipNow()
	index := []map[string]string{map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}, map[string]string{}}
	index[0]["host"] = "app0.example.com"
	index[1]["host"] = "app1.example.com"
	index[2]["host"] = "app2.example.com"
	index[3]["host"] = "app3.example.com"
	index[4]["host"] = "app4.example.com"
	index[0][drawing.GenerateUniqueKey()] = drawing.GenerateUniqueKey()
	index[1][drawing.GenerateUniqueKey()] = drawing.GenerateUniqueKey()
	index[2][drawing.GenerateUniqueKey()] = drawing.GenerateUniqueKey()
	index[3][drawing.GenerateUniqueKey()] = drawing.GenerateUniqueKey()
	index[4][drawing.GenerateUniqueKey()] = drawing.GenerateUniqueKey()
	ring := make([]string, len(index))
	for k, v := range index {
		ring[k] = v["host"]
	}

	for k, v := range index {
		next := k
		//body := ""
		WhoAmI = v["host"]
		for i := 0; i < 10; i++ {
			if next != -1 {
				//next, body = handleRing(body, ring, &index[next], nil)
			}
		}
	}
	if len(index[0]) != 6 || len(index[0]) != len(index[3]) {
		t.Error("not full propagation")
		t.Error(len(index[0]))
		t.Error(len(index[1]))
		t.Error(len(index[2]))
		t.Error(len(index[3]))
		t.Error(len(index[4]))
		t.Error(strings.ReplaceAll(fmt.Sprintf("%v", index), " map", "\nmap"))
	}
	fmt.Println(index)
}
