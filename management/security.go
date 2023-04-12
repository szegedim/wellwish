package management

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func QuantumGradeAuthorization() {
	// Just wait a second.
	// TODO mutex lock
	time.Sleep(1 * time.Second)
	// This is what you do before buying that doggycattycoin for too much.
	// Wait a sec... Is it a suspicious call?
}

func AddAdminForUrl(url string) string {
	if !strings.Contains(url, "?") {
		return fmt.Sprintf("%s?apikey=%s", url, administrationKey)
	}
	return fmt.Sprintf("%s&apikey=%s", url, administrationKey)
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
