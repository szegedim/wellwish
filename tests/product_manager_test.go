package tests

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"os/exec"
	"testing"
)

func TestLongTermCosts(t *testing.T) {
	ret := drawing.NoErrorString(exec.Command("go", "run", "../burst/wc/main.go", "..").Output())
	var begin, lines, end string
	_ = englang.ScanfContains(ret, "clc:%s\n", &begin, &lines, &end)
	fmt.Println(lines)
	if englang.Decimal(lines) > 6000 {
		t.Error("project became larger than what can be handled by one developer")
		t.Log("\n" + ret)
	}
}
