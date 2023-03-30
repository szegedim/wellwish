package burst

import (
	"gitlab.com/eper.io/engine/drawing"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	code, _ := io.ReadAll(drawing.NoErrorFile(os.Open("./cgi/main.go")))
	stdout, in := io.Pipe()
	go func() {
		_, _ = in.Write([]byte("Hello Burst!"))
		_ = in.Close()
	}()
	out, stdin := io.Pipe()
	go func() {
		Run(code, stdout, stdin)
		_ = stdin.Close()
	}()

	x, _ := io.ReadAll(out)
	s := string(x)
	if s != "Hello World!\n" {
		t.Error(s)
	}
	t.Log(s)
}

func TestBurstClient(t *testing.T) {
	code := "http://127.0.0.1:8887/code"
	client := "127.0.0.1:8888"

	go func() {
		http.HandleFunc("/code", func(w http.ResponseWriter, r *http.Request) {
			dummyCode := []byte("package main\nimport \"fmt\"\nfunc main() {fmt.Println(\"Hello World!\")}\n")
			_, _ = w.Write(dummyCode)
		})

		err := http.ListenAndServe(":8887", nil)
		if err != nil {
			t.Error(err)
		}
	}()
	for true {
		_, err := DownloadCode(code)
		if err == nil {
			break
		}
	}

	l, _ := net.Listen("tcp", client)

	metalKey := drawing.GenerateUniqueKey()
	_ = os.WriteFile("/tmp/apikey", []byte(metalKey), 0700)
	go func(key string) {
		c, _ := l.Accept()
		w, r := TlsServer(c, client)
		s := drawing.GenerateUniqueKey()
		cmp := make([]byte, len(s))
		n, _ := r.Read(cmp)
		if n != len(s) {
			t.Error("no auth")
			return
		}
		auth := string(cmp)
		if key != auth {
			t.Error("bad auth")
		}
		_, _ = w.Write([]byte("Hello World!"))
		b := make([]byte, 1024)
		n, _ = r.Read(b)
		t.Log(string(b[0:n]))
	}(metalKey)
	// docker run ...
	Burst(client, code)
}
