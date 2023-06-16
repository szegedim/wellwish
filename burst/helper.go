package burst

import (
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

var MockPhp = "<html><body><?php echo \"Hello World!\" ?></body></html>"
var MockPhpResult = "<html><body>Hello World!</body></html>"

// Curl is a function that translates Englang to code.
// Englang is easy to transmit over networks and backup files.
// Even your accountant can read the raw data.
// This makes them safer and cheaper to use than JSON, COM/RPC, CORBA, or XML.
func Curl(command string, data string) string {
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
	resp, _ := c.Do(request)
	download := make([]byte, 0)
	if resp != nil && resp.StatusCode == http.StatusTemporaryRedirect && redirect {
		target := resp.Header.Get("Location")
		return Curl(strings.Replace(command, url, target, 1), data)
	}
	if resp != nil {
		download = drawing.NoErrorBytes(io.ReadAll(resp.Body))
	}
	if resp != nil && resp.StatusCode == http.StatusOK && len(download) == 0 {
		return "success"
	}
	return string(download)
}

func FinishCleanup() {
	ContainerRunning = map[string]string{}
	BurstSession = map[string]string{}
}

func BoxCoreForTests() {
	participationKey := drawing.NoErrorString(exec.Command("curl", "-X", "GET", fmt.Sprintf("http://127.0.0.1%s/idle?apikey=%s", metadata.Http11Port, metadata.ActivationKey)).Output())
	//participationKey := Curl(englang.Printf("curl -X GET http://127.0.0.1%s/idle?apikey=%s", metadata.Http11Port, metadata.ActivationKey), "")

	started := time.Now()
	for time.Now().Before(started.Add(MaxBurstRuntime * 4)) {
		instructions := drawing.NoErrorString(exec.Command("curl", "-X", "GET", fmt.Sprintf("http://127.0.0.1%s/idle?apikey=%s", metadata.Http11Port, participationKey)).Output())
		if instructions != "" {
			ret := drawing.NoErrorString(exec.Command("curl", "-d", "<html><body>Hello World!</body></html>", "-X", "PUT", fmt.Sprintf("http://127.0.0.1%s/idle?apikey=%s", metadata.Http11Port, participationKey)).Output())
			fmt.Println("got instructions", instructions, "result", ret)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
