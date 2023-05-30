package tests

import (
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
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

func TestCustomerScenario(t *testing.T) {
	// Install a cluster somewhere
	MainTestLock.Lock()
	defer MainTestLock.Unlock()
	metadata.NodePattern = "http://127.0.0.1:771*"
	metadata.SiteUrl = "http://127.0.0.1:7716"
	metadata.Http11Port = ":7716"
	done := make(chan int)
	// Uncomment this to debug
	//go func(ready chan int) {
	//	time.Sleep(2 * time.Second)
	//	_ = os.Chdir("..")
	//	server.Main([]string{"go", metadata.Http11Port})
	//}(done)
	go func(ready chan int) { time.Sleep(2 * time.Second); runTestServer(t, ready, ":7716", 60*time.Second) }(done)

	time.Sleep(3 * time.Second)
	// Install a backup server
	// Activate
	curl(englang.Printf("curl -X GET %s/activate?apikey=%s", metadata.SiteUrl, metadata.ActivationKey), "")
	time.Sleep(3 * time.Second)
	// Buy a voucher
	me := fmt.Sprintf(metadata.OrderPattern, "\vExample Buyer Inc.\v", "\v111 S Ave\v, \vSan Fransisco\v, \vCA\v, \v55555\v, \vUnited States\v", "\vinfo\v@\vexample.com\v", "\v10\v", metadata.UnitPrice, "USD 10", "0")
	invoice := curl(englang.Printf("curl -X PUT %s/checkout", metadata.SiteUrl), me)
	if len(invoice) != len(drawing.GenerateUniqueKey()) {
		t.Error("We could not order voucher")
	}
	fmt.Println("Checked out invoice", invoice)
	// Get coin file
	coin := curl(englang.Printf("curl -X GET %s/invoice.coin?apikey=%s", metadata.SiteUrl, invoice), "")
	fmt.Println("Coin file", coin)
	// Buy a temporary sack
	sack := curl(englang.Printf("curl -X PUT %s/tmp.coin?apikey=%s", metadata.SiteUrl, invoice), "")
	fmt.Println("Sack", sack)
	// Save a temporary sack
	upload := curl(englang.Printf("curl -X PUT %s/tmp?apikey=%s", metadata.SiteUrl, sack), "abc")
	fmt.Println("Sack upload", upload)
	// Read back the sack
	content := curl(englang.Printf("curl -X GET %s/tmp?apikey=%s", metadata.SiteUrl, sack), "")
	fmt.Println("Sack data", content)
	if content != "abc" {
		t.Error("content not stored")
	}
	info := curl(englang.Printf("curl -X TRACE %s/tmp?apikey=%s", metadata.SiteUrl, sack), "")
	fmt.Println("Sack info", info)
	if !strings.Contains(info, "Validated until") {
		t.Error("sack info invalid")
	}
	deleted := curl(englang.Printf("curl -X DELETE %s/tmp?apikey=%s", metadata.SiteUrl, sack), "")
	fmt.Println("Sack deleted", deleted)
	if deleted != "success" {
		t.Error("sack should be deleted")
	}
	content = curl(englang.Printf("curl -X GET %s/tmp?apikey=%s", metadata.SiteUrl, sack), "")
	fmt.Println("Sack data", content)
	// Run a burst
	// Run a burst with a sack
	// Mine a random number
	// Stamp a contract with a burst
	//curl(englang.Printf("curl -X GET %s", metadata.SiteUrl), "")
	<-done
}

func curl(command string, data string) string {
	options := ""
	method := "GET"
	var url string
	englang.Scanf1(command+"fdsgdfgfdvdds", "curl %s-X %s %s"+"fdsgdfgfdvdds", &options, &method, &url)
	redirect := false
	if strings.Contains(options, "-L") {
		redirect = true
	}
	upload := bytes.NewBufferString(data)
	request, _ := http.NewRequest(method, url, upload)
	var c http.Client
	resp, _ := c.Do(request)
	download := make([]byte, 0)
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
