package management

import (
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

var lock sync.Mutex

func QuantumGradeAuthorization() {
	// Just wait a second.
	// This is what you do before buying that doggycattycoin for too much.
	// Wait a sec... Is it a suspicious call?
	// We basically cap the bandwidth of attackers
	lock.Lock()
	time.Sleep(15 * time.Millisecond)
	lock.Unlock()
}

func AddAdminForUrl(url string) string {
	if !strings.Contains(url, "?") {
		return fmt.Sprintf("%s?apikey=%s", url, metadata.ManagementKey)
	}
	return fmt.Sprintf("%s&apikey=%s", url, metadata.ManagementKey)
}

func HttpProxyRequest(url string, method string, bodyIn io.Reader) ([]byte, error) {
	// Poke around within the mesh network
	if method == "" {
		method = "GET"
	}
	if bodyIn == nil {
		bodyIn = &bytes.Buffer{}
	}
	req, err := http.NewRequest(method, url, bodyIn)
	if err != nil {
		return nil, err
	}
	// Use a client not associated with the Server.
	var c http.Client
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		return body, fmt.Errorf(resp.Status)
	}
	return body, nil
}
