package tests

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"os/exec"
	"testing"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func TestLongTermCosts(t *testing.T) {
	ret := drawing.NoErrorString(exec.Command("go", "run", "../burst/wc/main.go", "..").Output())
	var begin, lines, end string
	_ = englang.ScanfContains(ret, "clc:%s\n", &begin, &lines, &end)
	fmt.Println(lines)
	if englang.Decimal(lines) > 5100 {
		t.Error("project became larger than what can be handled by 0.6 developer")
		t.Log("\n" + ret)
	}
}
