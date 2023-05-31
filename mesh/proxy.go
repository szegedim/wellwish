package mesh

import (
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/englang"
	"io"
	"net/http"
)

func RedirectToPeerServer(w http.ResponseWriter, r *http.Request) error {
	apiKey := r.URL.Query().Get("apikey")
	if apiKey == "" {
		return fmt.Errorf("not found")
	}
	server := GetIndex(apiKey)
	if server == "" || server == WhoAmI {
		return fmt.Errorf("not found")
	}
	if englang.Synonym(Nodes[server], "This node got an eviction notice.") {
		return fmt.Errorf("not found")
	}
	modified := fmt.Sprintf("%s%s", server, r.URL.RequestURI())
	resp, status, _ := httpProxyRequest(modified, r.Method, r.Body)
	_, _ = io.Copy(w, resp)
	w.WriteHeader(status)
	return nil
}

func httpProxyRequest(url string, method string, bodyIn io.Reader) (io.ReadCloser, int, error) {
	// Poke around within the mesh network
	if method == "" {
		method = "GET"
	}
	if bodyIn == nil {
		bodyIn = &bytes.Buffer{}
	}
	req, err := http.NewRequest(method, url, bodyIn)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// Use a client not associated with the Server.
	var c http.Client
	resp, err := c.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return resp.Body, resp.StatusCode, nil
}
