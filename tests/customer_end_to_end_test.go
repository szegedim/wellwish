package tests

import (
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/burst"
	"gitlab.com/eper.io/engine/burst/php"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func TestCustomerUseCase(t *testing.T) {
	// Install a cluster somewhere
	time.Sleep(900 * time.Millisecond)
	mainTestLocalPorts.Lock()
	defer mainTestLocalPorts.Unlock()
	defer func() {
		time.Sleep(2 * burst.MaxBurstRuntime)
		burst.FinishCleanup()
		time.Sleep(2 * burst.MaxBurstRuntime)
	}()

	metadata.NodePattern = "http://127.0.0.1:771*"
	var PrimaryPort = ":7716"
	var StreamPort = ":7717"
	siteUrl := "http://127.0.0.1:7716"
	var SiteSecondaryNodeUrl = "http://127.0.0.1" + StreamPort
	metadata.Http11Port = ":7716"
	done := make(chan int)
	// Uncomment this to debug
	//go func(ready chan int) {
	//	time.Sleep(2 * time.Second)
	//	_ = os.Chdir("..")
	//	server.Main([]string{"go", PrimaryPort})
	//}(done)
	go func(ready chan int) {
		time.Sleep(100 * time.Millisecond)
		runTestServer(t, ready, PrimaryPort, 60*time.Second)
	}(done)
	go func(ready chan int) {
		time.Sleep(100 * time.Millisecond)
		runTestServer(t, ready, StreamPort, 60*time.Second)
	}(done)

	for {
		x := burst.Curl("curl -X GET "+siteUrl+"/health", "")
		if x != "" {
			break
		}
		if x != "" {
			fmt.Println(x)
		}
		time.Sleep(100 * time.Millisecond)
	}
	time.Sleep(3 * time.Second)
	// Install a backup server
	// Activate
	curl(englang.Printf("curl -X GET %s/activate?apikey=%s", siteUrl, metadata.ActivationKey), "")
	time.Sleep(3 * time.Second)
	// Buy a voucher
	me := fmt.Sprintf(metadata.OrderPattern, "\vExample Buyer Inc.\v", "\v111 S Ave\v, \vSan Fransisco\v, \vCA\v, \v55555\v, \vUnited States\v", "\vinfo\v@\vexample.com\v", "\v10\v", metadata.UnitPrice, "USD 10", "0")
	invoice := curl(englang.Printf("curl -X PUT %s/checkout", siteUrl), me)
	if len(invoice) != len(drawing.GenerateUniqueKey()) {
		t.Error("We could not order voucher", invoice)
	}
	fmt.Println("Checked out invoice", invoice)
	// Get coin file
	coin := curl(englang.Printf("curl -X GET %s/invoice.coin?apikey=%s", siteUrl, invoice), "")
	fmt.Println("Coin file", coin)
	// Check vending logic
	bag0 := curl(englang.Printf("curl -X PUT %s/tmp.coin?apikey=%s", siteUrl, invoice), "")
	fmt.Println("Temporary bag", bag0)
	bag1 := curl(englang.Printf("curl -X PUT %s/tmp.coin?apikey=%s", siteUrl, invoice), coin)
	fmt.Println("Temporary bag From coin", bag1)
	bag2 := curl(englang.Printf("curl -X PUT %s/tmp.coin?apikey=%s", siteUrl, drawing.GenerateUniqueKey()), drawing.GenerateUniqueKey())
	fmt.Println("Temporary bag From malicious coin", bag2)
	if bag2 != "" {
		t.Error("security issue")
	}

	// Buy a temporary bag
	bag := curl(englang.Printf("curl -X PUT %s/tmp.coin?apikey=%s", siteUrl, invoice), "")
	fmt.Println("bag", bag)
	// Save a temporary bag
	upload := curl(englang.Printf("curl -X PUT %s/tmp?apikey=%s", siteUrl, bag), "abc")
	fmt.Println("bag upload", upload)
	// Read back the bag
	content := curl(englang.Printf("curl -X GET %s/tmp?apikey=%s", siteUrl, bag), "")
	fmt.Println("bag data", content)
	if content != "abc" {
		t.Error("content not stored")
	}
	info := curl(englang.Printf("curl -X TRACE %s/tmp?apikey=%s", siteUrl, bag), "")
	fmt.Println("bag info", info)
	if !strings.Contains(info, "This is a bag storage") {
		t.Error("bag info invalid")
	}
	time.Sleep(20 * time.Second)
	secondary := curl(englang.Printf("curl -X TRACE %s/tmp?apikey=%s", SiteSecondaryNodeUrl, bag), "")
	fmt.Println("bag info", secondary)
	if !strings.Contains(secondary, "This is a bag storage") {
		t.Error("indexing does not work", bag)
		secondary = curl(englang.Printf("curl -X TRACE %s/healthz", siteUrl), "")
		fmt.Println("bag info", secondary)
		secondary = curl(englang.Printf("curl -X TRACE %s/healthz", SiteSecondaryNodeUrl), "")
		fmt.Println("bag info", secondary)
	}

	deleted := curl(englang.Printf("curl -X DELETE %s/tmp?apikey=%s", siteUrl, bag), "")
	fmt.Println("bag deleted", deleted)
	if deleted != "success" {
		t.Error("404 page not found")
	}
	content = curl(englang.Printf("curl -X GET %s/tmp?apikey=%s", siteUrl, bag), "")
	fmt.Println("mined data", content)
	if deleted != "success" {
		t.Error("bag should be deleted")
	}

	// Mine a random number
	goldMine0 := curl(englang.Printf("curl -X PUT %s/cryptonugget.coin?apikey=%s", siteUrl, invoice), "")
	fmt.Println("Temporary mine", goldMine0)
	goldMine1 := curl(englang.Printf("curl -X PUT %s/cryptonugget.coin?apikey=%s", siteUrl, invoice), coin)
	fmt.Println("Temporary mine from coin", goldMine1)
	goldNugget := curl(englang.Printf("curl -X GET %s/cryptonugget?apikey=%s", siteUrl, goldMine0), "")
	fmt.Println("Crypto gold nugget data", goldNugget)
	fmt.Println()

	// Buy a burst session
	burstSession := curl(englang.Printf("curl -X PUT %s/run.coin?apikey=%s", siteUrl, ""), coin)
	fmt.Println("CPU Burst", burstSession)

	burst2 := curl(englang.Printf("curl -X GET %s/run.coin?apikey=%s", siteUrl, burstSession), "")
	fmt.Println("CPU Burst Session From coin", burst2)

	burst3 := curl(englang.Printf("curl -X PUT %s/run.coin?apikey=%s", siteUrl, drawing.GenerateUniqueKey()), drawing.GenerateUniqueKey())
	fmt.Println("CPU Burst From malicious coin", burst3)
	if burst3 != "" {
		t.Error("security issue")
	}

	go func() {
		// TODO attach directly?
		for {
			time.Sleep(100 * time.Millisecond)

			goRoot := os.Getenv("GOROOT")
			goroot := path.Join(goRoot, "bin", "go")
			p := path.Join("..", "burst", "box", "main.go")
			fmt.Println(p, metadata.Http11Port)
			err := exec.Command(goroot, "run", p, PrimaryPort).Run()
			if err != nil {
				fmt.Println("local result", err)
			}
		}
	}()
	fmt.Println("Running burst on", siteUrl)
	time.Sleep(2 * time.Second)
	run0 := curl(englang.Printf("curl -X PUT %s/run?apikey=%s", siteUrl, burstSession), "Run the following php code."+php.MockPhp)
	fmt.Println("CPU Burst result", run0)
	if run0 != "<html><body>Hello World!</body></html>" {
		t.Error("could not run sample php")
	}

	burst.FinishCleanup()
	<-done
	<-done
}

func curl(command string, data string) string {
	options := ""
	method := "GET"
	var url string
	_ = englang.Scanf1(command+"fdsgdfgfdvdds", "curl %s-X %s %s"+"fdsgdfgfdvdds", &options, &method, &url)
	redirect := false
	if strings.Contains(options, "-L") {
		redirect = true
	}
	upload := bytes.NewBufferString(data)
	request, _ := http.NewRequest(method, url, upload)
	var c http.Client
	resp, err := c.Do(request)
	download := make([]byte, 0)
	if err != nil {
		fmt.Println(err.Error())
	}
	if resp != nil && resp.StatusCode == http.StatusTemporaryRedirect && redirect {
		target := resp.Header.Get("Location")
		curl(strings.Replace(command, url, target, 1), data)
	}
	if resp != nil {
		download = drawing.NoErrorBytes(io.ReadAll(resp.Body))
	}
	if resp != nil && resp.StatusCode == http.StatusOK && len(download) == 0 {
		return "success"
	}
	return string(download)
}
