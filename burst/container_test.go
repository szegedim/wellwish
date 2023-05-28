package burst

import (
	"fmt"
	"gitlab.com/eper.io/engine/burst/php"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"gitlab.com/eper.io/engine/tests"
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

func TestContainerRunner(t *testing.T) {
	code, _ := io.ReadAll(drawing.NoErrorFile(os.Open("./helloworld/main.go")))
	stdout, in := io.Pipe()
	go func() {
		_, _ = in.Write([]byte("Hello Burst!"))
		_ = in.Close()
	}()
	out, stdin := io.Pipe()
	go func() {
		RunInTest(code, stdout, stdin)
		_ = stdin.Close()
	}()

	x, _ := io.ReadAll(out)
	s := string(x)
	if s != "Hello World!\n" {
		t.Error(s)
	}
	t.Log(s)
}

func TestContainerEndToEnd(t *testing.T) {
	tests.MainTestLock.Lock()
	defer tests.MainTestLock.Unlock()
	// Tests that share the same port udp:2121 must run in a row
	testContainer(t)
	testBurstRunner(t)
	time.Sleep(5 * time.Second)
	testBurstRunner(t)
	time.Sleep(5 * time.Second)
	testBurstEndToEndApi(t, "")
	time.Sleep(5 * time.Second)
	coinToUse, _ := GenerateTestCoins()
	t.Log(coinToUse)
	// TODO use coin
	testBurstEndToEndApi(t, "")
	time.Sleep(5 * time.Second)
}

func testContainer(t *testing.T) {
	done := make(chan interface{})
	// Server
	go func() { SetupBurstIdleProcess() }()

	go func() {
		time.Sleep(3 * time.Second)
		deferredKey, result := RunBurst("Run the following php code." + php.MockPhp)
		if deferredKey != "" && result == "" {
			time.Sleep(3 * time.Second)
			result = GetBurst(deferredKey)
		}
		if result != "<html><body>Hello World!</body></html>" {
			t.Error(result)
		}
		t.Log("SUCCESS", result)
		done <- true
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)
		// Box
		err := RunBox()
		if err != nil {
			t.Error(err)
		}
	}()

	select {
	case <-time.After(15 * time.Second):
		t.Error("timeout")
	case <-done:
	}
	FinishCleanup()
}

func testBurstRunner(t *testing.T) {
	done := make(chan interface{})

	// Server
	go func() {
		SetupBurstIdleProcess()
	}()

	// Client
	go func() {
		time.Sleep(3 * time.Second)
		deferredKey, result := RunBurst("Run the following php code." + php.MockPhp)
		if deferredKey != "" && result == "" {
			time.Sleep(3 * time.Second)
			result = GetBurst(deferredKey)
		}
		if result != "<html><body>Hello World!</body></html>" {
			t.Error(result)
		}

		t.Log("SUCCESS", result)
		done <- true
	}()

	// Container
	go func() {
		time.Sleep(100 * time.Millisecond)

		goRoot := os.Getenv("GOROOT")
		goroot := path.Join(goRoot, "bin", "go")
		cmd := exec.Command(goroot, "run", path.Join("..", "burst", "box", "main.go"))
		output, err := cmd.Output()
		if err != nil {
			output = []byte(err.Error())
		}
		fmt.Println(string(output))
	}()

	select {
	case <-time.After(15 * time.Second):
		t.Error("timeout")
	case <-done:
	}
	FinishCleanup()
}

func testBurstEndToEndApi(t *testing.T, paidSession string) {
	const NumberOfContainers = 5
	const NumberOfLambdaCalls = 2
	testPath := "/" + drawing.RedactPublicKey(drawing.GenerateUniqueKey())

	done := make(chan interface{})

	// Server
	go func() {
		SetupBurstLambdaEndpoint(testPath, paidSession != "")
		SetupBurstIdleProcess()
	}()

	go func() {
		_ = http.ListenAndServe(metadata.Http11Port, nil)
	}()

	if paidSession != "" {
		result := mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /api?apikey=%s with method PUT and content %s. The call expects englang.", metadata.Http11Port, paidSession, paidSession))
		fmt.Println("Burst session", result)

	}

	// Client
	for i := 0; i < NumberOfLambdaCalls; i++ {
		go func(delay time.Duration, done chan interface{}) {
			time.Sleep(1 * time.Second)
			time.Sleep(delay)
			task := "Run the following php code." + php.MockPhp
			result := mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path %s?apikey=%s with method GET and content %s. The call expects englang.", metadata.Http11Port, testPath, paidSession, task))
			if result != "<html><body>Hello World!</body></html>" {
				t.Error("<html><body>Hello World!</body></html>")
			}
			t.Log("SUCCESS", result)
			if done != nil {
				done <- true
			}
		}(time.Duration(i)*time.Second, done)
	}

	for i := 0; i < NumberOfContainers; i++ {
		// Container
		go func() {
			time.Sleep(100 * time.Millisecond)

			for {
				goRoot := os.Getenv("GOROOT")
				goroot := path.Join(goRoot, "bin", "go")
				cmd := exec.Command(goroot, "run", path.Join("..", "burst", "box", "main.go"))
				output, err := cmd.Output()
				if err != nil {
					output = []byte(err.Error())
				}
				fmt.Println(string(output))
			}
		}()
	}

	for i := 0; i < NumberOfLambdaCalls; i++ {
		select {
		case <-time.After(15 * time.Second):
			t.Error("timeout")
		case <-done:
		}
	}

	FinishCleanup()
}
