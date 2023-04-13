package server

import (
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"os"
	"os/exec"
	"testing"
	"time"
)

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

	managementKey := drawing.GenerateUniqueKey()
	ret, err := management.HttpProxyRequest(englang.Printf("http://127.0.0.1:7777/activate?activationkey=%s&apikey=%s", metadata.ActivationKey, managementKey), "", nil)
	t.Log("activated", string(ret))
	if err != nil {
		t.Error(err)
	}
	t.Log(billing.IssueVouchers(
		drawing.GenerateUniqueKey(), "100",
		"Example Inc.", "1 First Ave, USA",
		"hq@opensource.eper.io", "USD 3"))
	<-x
	<-y
	// Never exits
	//<-z
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
	t.Log(string(b))
	t.Log(p.ProcessState.ExitCode())
	ready <- 1
}
