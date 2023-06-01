package burst

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/burst/php"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func TestBurst(t *testing.T) {
	go func() {
		err := http.ListenAndServe(metadata.Http11Port, nil)
		if err != nil {
			t.Error(err)
		}
	}()

	if BurstRunners == 0 {
		DummyBroker()
	}
	Setup()
	billing.Setup()

	time.Sleep(time.Second)
	payment, order := generateTestCoins(englang.Printf("http://127.0.0.1%s", metadata.Http11Port))
	finalStatus := bytes.NewBufferString("")
	billing.GetCoinFile(order, bufio.NewWriter(finalStatus))
	// There is one used item at the end
	t.Log(finalStatus.String())

	burstSession := mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /run.coin?apikey=%s with method PUT and content %s. The call expects englang.", metadata.Http11Port, "", payment))
	fmt.Println("Burst session", burstSession)

	result := mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /run.coin?apikey=%s with method GET and content %s. The call expects englang.", metadata.Http11Port, burstSession, ""))
	fmt.Println("Burst session", result)

	time.Sleep(1 * time.Second)

	result = mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /run?apikey=%s with method PUT and content %s. The call expects englang.", metadata.Http11Port, burstSession, "Run the following php code."+php.MockPhp))
	fmt.Println("Burst result", result)
	if result != "<html><body>Hello World!</body></html>" {
		t.Error("not expected")
	}

	time.Sleep(MaxBurstRuntime)
	if len(ContainerResults) > 0 {
		t.Error("no cleanup")
	}
}

func generateTestCoins(siteUrl string) (string, string) {
	me := fmt.Sprintf(metadata.OrderPattern, "\vExample Buyer Inc.\v", "\v111 S Ave\v, \vSan Fransisco\v, \vCA\v, \v55555\v, \vUnited States\v", "\vinfo\v@\vexample.com\v", "\v10\v", metadata.UnitPrice, "USD 10", "0")
	invoice := Curl(englang.Printf("curl -X PUT %s/checkout", siteUrl), me)
	if len(invoice) != len(drawing.GenerateUniqueKey()) {
		return "We could not order voucher", "We could not order voucher"
	}
	fmt.Println("Checked out invoice", invoice)
	// Get coin file
	coin := Curl(englang.Printf("curl -X GET %s/invoice.coin?apikey=%s", siteUrl, invoice), "")
	fmt.Println("Coin file", coin)
	return invoice, coin
}

// This is a module code that runs burst containers.
// The big difference between these and other modules is that it actually does not have
// an entry point.
// The locality is ensured by private keys distributed early called 'metal'.
// This ensures that we have a local runner.
// What does this mean?
// - /idle responds to local endpoints only like a co-located container in the same pod
// - idle returns a task and a key to complete the task
// - malicious tasks may go for idle again
// - we protect against this by letting bursts run for a term e.g. ten seconds
// - we protect against this also by not issuing a new key until the previous one finishes
// - Each runner connects to the main site as /idle using the activation key
// - The activation key is deleted from the container once used
// - The init task of the container is our burst runner. It should be set debuggable by workload.
// - The init task kills the container, if the workload tries to kill it.
// - The final column is time fencing allowing /idle calls only once every minute when workloads are already gone.
// - The runner restarts after each run, so that any local state and code is lost disabling double /idle calls.

func runInTest(code []byte, stdin io.ReadCloser, stdout io.Writer) {
	goPath := path.Join(os.Getenv("GOROOT"), "bin", "go")
	workDir := path.Join("/tmp")
	mainGo := path.Join(workDir, "main.go")
	_ = os.WriteFile(mainGo, code, 0700)
	cmd := &exec.Cmd{
		Path: path.Join(goPath),
		Args: []string{"", "run", mainGo},
		Dir:  workDir,
		Env: []string{fmt.Sprintf("HOSTNAME=%s", metadata.SiteUrl),
			fmt.Sprintf("HOME=%s", workDir),
			fmt.Sprintf("GOROOT=%s", os.Getenv("GOROOT"))},
		Stderr: stdout,
		Stdin:  stdin,
	}

	stdoutRead, err := cmd.StdoutPipe()
	if err != nil {
		_, _ = stdout.Write([]byte(err.Error()))
		return
	}

	err = cmd.Start()
	if err != nil {
		_, _ = stdout.Write([]byte(err.Error()))
		return
	}

	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	go func(r io.ReadCloser, w io.ReadCloser) {
		_, err = io.Copy(stdout, r)
		if err != nil {
			_, _ = stdout.Write([]byte(err.Error()))
			return
		}

		_ = r.Close()
		_ = w.Close()
	}(stdoutRead, stdin)

	err = cmd.Wait()
	if err != nil {
		_, _ = stdout.Write([]byte(err.Error()))
		return
	}
}

func TestRun(t *testing.T) {
	code, _ := io.ReadAll(drawing.NoErrorFile(os.Open("./helloworld/main.go")))
	stdout, in := io.Pipe()
	go func() {
		_, _ = in.Write([]byte("Hello Burst!"))
		_ = in.Close()
	}()
	out, stdin := io.Pipe()
	go func() {
		runInTest(code, stdout, stdin)
		_ = stdin.Close()
	}()

	x, _ := io.ReadAll(out)
	s := string(x)
	if s != "Hello World!\n" {
		t.Error(s)
	}
	t.Log(s)
}
