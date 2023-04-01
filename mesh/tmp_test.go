package mesh

import (
	"gitlab.com/eper.io/engine/drawing"
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
	x := make(chan int)
	y := make(chan int)
	go func(ready chan int) { runServer(t, ready, "127.0.0.1:7779") }(y)
	go func(ready chan int) { time.Sleep(2 * time.Second); runServer(t, ready, "127.0.0.1:7780") }(x)
	<-x
	<-y
}

func runServer(t *testing.T, ready chan int, port string) {
	p := exec.Cmd{
		Dir:  "..",
		Path: "/Users/miklos_szegedi/schmied.us/private/go-darwin-arm64-bootstrap/bin/go",
		Args: []string{"go", "run", "main.go", port},
	}
	err := p.Start()
	if err != nil {
		t.Error(err)
	}
	go func() {
		time.Sleep(30 * time.Second)
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
