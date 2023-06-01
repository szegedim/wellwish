package main

import (
	"bytes"
	"gitlab.com/eper.io/engine/burst"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
	"strings"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Box is a container code that waits for a single burst and exits
// Box can run in a container launched as
// docker run -d --rm --restart=always --name box1 wellwish go run burst/box/main.go
// There are two ways to input and output data to and from boxes
// One is the burst `/run.coin` body request and return.
// Keep this small as it contains Englang that is duplicated in memory two times at least.
// The other way is to pass a bag url or cloud bucket url where the box streams any input or results.
// We do not log runtime or errors, the server takes care of that.
// TODO add timeout logic on paid vouchers

func main() {
	content := curl(englang.Printf("curl -X GET http://127.0.0.1%s/idle?apikey=%s", metadata.Http11Port, "bag"), "")

	var command, port, key string
	_ = englang.Scanf1(content, "Run this %s and return in http://127.0.0.1%s/idle?apikey=%s.", &command, &port, &key)
	ret := burst.RunExternalShell(command)

	x := englang.Printf("Run this %s and return in http://127.0.0.1%s/idle?apikey=%s.", ret, port, key)
	curl(englang.Printf("curl -X PUT http://127.0.0.1%s/idle?apikey=%s", port, key), x)
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
