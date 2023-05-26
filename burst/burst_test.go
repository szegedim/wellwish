package burst

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
	"os"
	"testing"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func TestRun(t *testing.T) {
	t.SkipNow()
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
	SetupBurst()
	payment, order := GenerateTestCoins()

	go func() { _ = http.ListenAndServe(metadata.Http11Port, nil) }()

	burstSession := mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /api?apikey=%s with method PUT and content %s. The call expects englang.", metadata.Http11Port, "", payment))
	fmt.Println("Burst session", burstSession)

	result := mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /api?apikey=%s with method GET and content %s. The call expects englang.", metadata.Http11Port, burstSession, ""))
	fmt.Println("Burst session", result)

	finalStatus := bytes.NewBufferString("")
	billing.GetCoinFile(order, bufio.NewWriter(finalStatus))
	// There is one used item at the end
	t.Log(finalStatus.String())
}

func GenerateTestCoins() (string, string) {
	voucher := drawing.GenerateUniqueKey()
	billing.IssueOrder(voucher, "100",
		"Example Inc.", "1 First Ave, USA",
		"hq@example.com", "USD 3")

	payment := bytes.NewBufferString("")
	billing.GetCoinFile(voucher, bufio.NewWriter(payment))
	return payment.String(), voucher
}
