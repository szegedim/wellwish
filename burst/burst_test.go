package burst

import (
	"bufio"
	"bytes"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/mesh"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

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
	BurstRunner(client, code)
}

func TestBurst(t *testing.T) {
	// api returns in few hundred milliseconds (polling unless asked)
	// idle returns in few hundred milliseconds (polling unless asked)
	// All containers are host (no access to each other)
	// Assumes cloud traffic is protected

	go func() { _ = http.ListenAndServe(":7777", nil) }()

	Setup()
	voucher := drawing.GenerateUniqueKey()
	billing.IssueOrder(voucher, "100",
		"Example Inc.", "1 First Ave, USA",
		"hq@opensource.eper.io", "USD 3")

	payment := bytes.NewBufferString("")
	billing.GetCoinFile(voucher, bufio.NewWriter(payment))
	x, _ := billing.RedeemCoin(payment.String())
	t.Log(x)

	done := make(chan bool)
	containerKey := drawing.GenerateUniqueKey()

	Container[containerKey] = "This container is idle"

	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(100 * time.Millisecond)
			ret := mesh.Englang(englang.Printf("Call server http://127.0.0.1:7777 path /idle?apikey=%s with method GET and content %s. The call expects englang.", containerKey, "Wait for 10 seconds for a new task."))
			if ret == "Hello World!" || ret == "Hello Moon!" {
				//t.Log(ret)
				time.Sleep(100 * time.Millisecond)
				ret = mesh.Englang(englang.Printf("Call server http://127.0.0.1:7777 path /idle?apikey=%s with method PUT and content %s. The call expects success.", containerKey, ret))
				if ret != "success" {
					t.Error(ret)
				}
			}
			if ret != "too early" && ret != "success" {
				t.Error(ret)
			}
		}
	}()

	go func() {
		var burstSession, burst string
		ret := mesh.Englang(englang.Printf("Call server http://127.0.0.1:7777 path /api with method PUT and content %s. The call expects englang.", payment.String()))
		if ret != "too early" {
			t.Log(ret)
			burstSession = ret
		}
		time.Sleep(100 * time.Millisecond)
		ret = mesh.Englang(englang.Printf("Call server http://127.0.0.1:7777 path /api?apikey=%s with method GET and content %s. The call expects englang.", burstSession, "Hello World!"))
		if ret != "too early" {
			t.Log(ret)
			burst = ret
		}

		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			ret = mesh.Englang(englang.Printf("Call server http://127.0.0.1:7777 path /api?apikey=%s with method GET and content %s. The call expects englang.", burst, ""))
			if ret == "Hello World!" {
				t.Log(ret)
				done <- true
				break
			}
			if ret != "too early" {
				t.Log(ret)
			}
		}
	}()

	go func() {
		var burstSession, burst string
		ret := mesh.Englang(englang.Printf("Call server http://127.0.0.1:7777 path /api with method PUT and content %s. The call expects englang.", payment.String()))
		if ret != "too early" {
			t.Log(ret)
			burstSession = ret
		}
		time.Sleep(100 * time.Millisecond)
		ret = mesh.Englang(englang.Printf("Call server http://127.0.0.1:7777 path /api?apikey=%s with method GET and content %s. The call expects englang.", burstSession, "Hello Moon!"))
		if ret != "too early" {
			t.Log(ret)
			burst = ret
		}

		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			ret := mesh.Englang(englang.Printf("Call server http://127.0.0.1:7777 path /api?apikey=%s with method GET and content %s. The call expects englang.", burst, ""))
			if ret == "Hello Moon!" {
				t.Log(ret)
				done <- true
				break
			}
			if ret != "too early" {
				t.Log(ret)
			}
		}
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-time.After(10 * time.Second):
			t.Error("timeout")
		case <-done:
		}
	}
}
