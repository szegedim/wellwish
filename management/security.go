package management

import (
	"bytes"
	"io"
	"net/http"
	url2 "net/url"
	"time"
)

func QuantumGradeAuthorization() {
	// Just wait a second.
	time.Sleep(1 * time.Second)
	// This is what you do before buying that doggycattycoin for too much.
	// Wait a sec... Is it a suspicious call?
}

func AddAdminForUrl(url string) string {
	url1, err := url2.Parse(url)
	if err != nil {
		return ""
	}

	url1.Query().Add("apikey", administrationKey)
	return url1.String()
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
		return nil, err
	}
	_ = resp.Body.Close()

	return body, nil
}
