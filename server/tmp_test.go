package server

import (
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	_ = os.Remove("/tmp/a")
	_ = os.Remove("/tmp/b")
	_ = os.Remove("/tmp/checkpoint")
	_ = os.WriteFile("/tmp/a", []byte("abc"), 0700)
	_ = os.Link("/tmp/a", "/tmp/checkpoint")

	_ = os.WriteFile("/tmp/b", []byte("def"), 0700)
	_ = os.Remove("/tmp/checkpoint")
	_ = os.Link("/tmp/b", "/tmp/checkpoint")

	x, _ := io.ReadAll(drawing.NoErrorFile(os.Open("/tmp/checkpoint")))
	if string(x) != "def" {
		t.Error(string(x))
	}
}

func TestCluster(t *testing.T) {
	_ = os.Chdir("..")
	x := make(chan int)
	y := make(chan int)
	z := make(chan int)
	go func(ready chan int) { time.Sleep(2 * time.Second); runServer(t, ready, ":7777") }(z)
	go func(ready chan int) { time.Sleep(4 * time.Second); Main([]string{"go", ":7778"}) }(y)
	go func(ready chan int) { time.Sleep(2 * time.Second); runServer(t, ready, ":7779") }(x)
	for {
		_, err := management.HttpProxyRequest(englang.Printf("http://127.0.0.1:7777/healthz"), "", nil)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	ret, err := management.HttpProxyRequest(englang.Printf("http://127.0.0.1:7778/activate?activationkey=%s", metadata.ActivationKey), "", nil)
	t.Log("activated", string(ret), err.Error())
	t.Log(billing.IssueVouchers(
		drawing.GenerateUniqueKey(), "100",
		"Example Inc.", "1 First Ave, USA",
		"hq@opensource.eper.io", "USD 3"))
	<-x
	<-y
	<-z
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
		time.Sleep(9 * 60 * time.Second)
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
