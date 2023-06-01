package tests

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/burst"
	"gitlab.com/eper.io/engine/burst/php"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"net/http"
	"testing"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func TestContainerStandAlone(t *testing.T) {
	// Tests that share the same port udp:2121 must run in a row

	// Server
	burst.Setup()
	go func() {
		_ = http.ListenAndServe(metadata.Http11Port, nil)
	}()

	testContainer(t)

	testBurstEndToEndApi(t)

	time.Sleep(2 * burst.MaxBurstRuntime)
	burst.FinishCleanup()
}

func testContainer(t *testing.T) {
	MainTestLocalPorts.Lock()
	defer MainTestLocalPorts.Unlock()
	defer func() {
		time.Sleep(2 * burst.MaxBurstRuntime)
		burst.FinishCleanup()
		time.Sleep(2 * burst.MaxBurstRuntime)
	}()

	done := make(chan interface{})

	time.Sleep(1000 * time.Millisecond)

	// Generate payment
	burstSession, finalStatus := generateBurstSession()
	// There is one used item at the end
	t.Log(finalStatus.String())

	go func() {
		time.Sleep(5 * time.Second)

		result := runBurst("Run the following php code."+php.MockPhp, burstSession)

		t.Log("LOG", result)
		done <- true
	}()

	go func() {
		time.Sleep(100 * time.Millisecond)

		for {
			// Box
			err := burst.RunBox()
			if err != nil {
				t.Error(err)
			}
		}
	}()

	select {
	case <-time.After(30 * time.Second):
		t.Error("timeout")
	case <-done:
	}
	burst.FinishCleanup()
}

func generateBurstSession() (string, *bytes.Buffer) {
	payment, order := generateTestCoins()

	burstSession := mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /run.coin?apikey=%s with method PUT and content %s. The call expects englang.", metadata.Http11Port, "", payment))
	fmt.Println("Burst session", burstSession)

	result := mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /run.coin?apikey=%s with method GET and content %s. The call expects englang.", metadata.Http11Port, burstSession, ""))
	fmt.Println("Burst session", result)

	finalStatus := bytes.NewBufferString("")
	billing.GetCoinFile(order, bufio.NewWriter(finalStatus))
	return burstSession, finalStatus
}

func runBurst(request string, burstSession string) string {
	result := mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /run?apikey=%s with method PUT and content %s. The call expects englang.", metadata.Http11Port, burstSession, request))
	//fmt.Println("Burst result", result)
	return result
}

func testBurstEndToEndApi(t *testing.T) {
	MainTestLocalPorts.Lock()
	defer MainTestLocalPorts.Unlock()
	defer func() {
		time.Sleep(2 * burst.MaxBurstRuntime)
		burst.FinishCleanup()
		time.Sleep(2 * burst.MaxBurstRuntime)
	}()

	done := make(chan interface{})

	const NumberOfContainers = 5
	const NumberOfLambdaCalls = 2

	time.Sleep(2000 * time.Millisecond)

	// Generate payment
	burstSession, finalStatus := generateBurstSession()
	// There is one used item at the end
	t.Log(finalStatus.String())

	for i := 0; i < NumberOfLambdaCalls; i++ {
		go func(delay int) {
			time.Sleep(time.Duration(delay) * time.Second)

			result := runBurst("Run the following php code."+php.MockPhp, burstSession)

			t.Log("LOG", result)
			done <- true
		}(i)
	}

	go func() {
		for i := 0; i < NumberOfContainers; i++ {
			// Container
			go func() {
				time.Sleep(100 * time.Millisecond)

				for {
					// Box
					err := burst.RunBox()
					if err != nil {
						t.Error(err)
					}
				}
			}()
		}
	}()

	for i := 0; i < NumberOfLambdaCalls; i++ {
		select {
		case <-time.After(15 * time.Second):
			t.Error("timeout")
		case <-done:
		}
	}

	burst.FinishCleanup()
}

func generateTestCoins() (string, string) {
	voucher := drawing.GenerateUniqueKey()
	billing.IssueOrder(voucher, "100",
		"Example Inc.", "1 First Ave, United States",
		"hq@example.com", "USD 3")

	payment := bytes.NewBufferString("")
	billing.GetCoinFile(voucher, bufio.NewWriter(payment))
	return payment.String(), voucher
}
