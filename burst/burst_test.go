package burst

import (
	"bufio"
	"bytes"
	"gitlab.com/eper.io/engine/billing"
	box "gitlab.com/eper.io/engine/burst/box/englang"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"io"
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
	code, _ := io.ReadAll(drawing.NoErrorFile(os.Open("./helloworld/main.go")))
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

func TestBurst(t *testing.T) {
	// api returns in few hundred milliseconds (polling unless asked)
	// idle returns in few hundred milliseconds (polling unless asked)
	// All containers are host (no access to each other)
	// Assumes cloud traffic is protected

	go func() { _ = http.ListenAndServe(metadata.Http11Port, nil) }()

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

	SetupRunner()

	go func() {
		for i := 0; i < 100; i++ {
			time.Sleep(100 * time.Millisecond)
			box.Vars["APIKEY"] = "/tmp/container2.metal"
			box.Englang("Initialize container with key from file from environment variable APIKEY.")
			if box.Vars["accumulator"] == "Hello World!" || box.Vars["accumulator"] == "Hello Moon!" {
				time.Sleep(100 * time.Millisecond)
				box.Englang("Finish container with content from accumulator and key from file from environment variable APIKEY.")
			}
		}
	}()

	burst1 := func(message string) {
		var burstSession, burst string
		ret := mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /api with method PUT and content %s. The call expects englang.", metadata.Http11Port, payment.String()))
		if ret != "too early" {
			burstSession = ret
		}
		time.Sleep(100 * time.Millisecond)
		ret = mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /api?apikey=%s with method GET and content %s. The call expects englang.", metadata.Http11Port, burstSession, message))
		if ret != "too early" {
			burst = ret
		}

		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			ret = mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /api?apikey=%s with method GET and content %s. The call expects englang.", metadata.Http11Port, burst, ""))
			if ret == message {
				t.Log(ret)
				done <- true
				break
			}
			if ret != "too early" {
				t.Log(ret)
			}
		}
	}

	var messages = []string{"Hello World!", "Hello Moon!"}
	for _, v := range messages {
		go burst1(v)
	}

	for range messages {
		select {
		case <-time.After(10 * time.Second):
			// Timeout may mean a port conflict
			t.Error("timeout")
		case <-done:
		}
	}
}
