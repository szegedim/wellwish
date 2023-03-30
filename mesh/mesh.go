package mesh

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Mesh containers do some heavy lifting for the entire cluster.
// Individual sack and burst containers are not aware of the cluster details.
// They have only a pointer to the cluster entry point, a https site address.

// Mesh containers listen to 7778 and communicate through in Englang.
// It would not require https within the VPC, but we use TLS closure for now.
// - Mesh reads sack checkpoint backups.
// - Mesh knows where to find a sack and forwards requests to other nodes
// - Mesh can restore an entire cluster
// - Mesh sets up a node metal file with key for burst nodes
// - Burst nodes log in with the key in the metal file to mesh to get tasks to run.
// - Mesh can be on the same container as sacks or others running static code
// - Burst is running dynamic code, it exits every time after a run.

func Setup() {
	http.HandleFunc("/node", func(w http.ResponseWriter, r *http.Request) {
		adminKey, err := management.EnsureAdministrator(w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Load and Propagate server names from api
		address := r.Header.Get("address")
		if address == "" {
			w.WriteHeader(http.StatusNoContent)
		}
		if Nodes[address] != "" {
			// No reflection, avoid hangs
			return
		}
		Nodes[address] = address
		for node := range Nodes {
			_, _ = HttpRequest(fmt.Sprintf("http://%s:7777/node?apikey=%s&address=%s", node, adminKey, address), "PUT", nil)
		}
	})

	http.HandleFunc("/checkpoint", func(w http.ResponseWriter, r *http.Request) {
		// Read state from stateful node containers (sacks)
		apiKey, err := management.EnsureAdministrator(w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		buf := bytes.Buffer{}
		for node := range Nodes {
			b, _ := HttpRequest(fmt.Sprintf("http://%s:7777/checkpoint?apikey=%s", node, apiKey), "GET", nil)
			buf.Write(b)
		}

		checkpoint := path.Join("/tmp", drawing.GenerateUniqueKey())
		_ = os.WriteFile(checkpoint, buf.Bytes(), 0700)
		_ = os.Remove("/tmp/checkpoint")
		_ = os.Link(checkpoint, "/tmp/checkpoint")
	})

	http.HandleFunc("/restore", func(w http.ResponseWriter, r *http.Request) {
		// Restore state to stateful node containers (sacks)
		_, err := management.EnsureAdministrator(w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		apiKey := r.Header.Get("apikey")
		if apiKey == "" {
			w.WriteHeader(http.StatusUnauthorized)
		}
		buf := bytes.Buffer{}
		for node := range Nodes {
			b, _ := HttpRequest(fmt.Sprintf("http://%s:7777/checkpoint?apikey=%s", node, apiKey), "GET", nil)
			buf.Write(b)
		}
		checkpoint := drawing.GenerateUniqueKey()
		_ = os.WriteFile(path.Join("/tmp", checkpoint), buf.Bytes(), 0700)
		_ = os.Remove("/tmp/checkpoint")
		_ = os.Rename("/tmp/backup", "/tmp/garbage")
		_ = os.Rename("/tmp/checkpoint", "/tmp/backup")
		_ = os.Remove("/tmp/garbage")

	})
}

func FindApikeyServer(apiKey string) string {
	return Index[apiKey]
}

func MeshProxy(w http.ResponseWriter, r *http.Request) {
	apiKey := r.Header.Get("apikey")
	if apiKey == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	original := r.URL.String()
	server := FindApikeyServer(apiKey)
	if server == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	modified := strings.Replace(original, metadata.SiteUrl, server, 1)
	if modified == original {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	b, _ := HttpRequest(fmt.Sprintf("http://%s:7777/checkpoint?apikey=%s", modified, apiKey), r.Method, r.Body)
	_, _ = w.Write(b)
}

func IndexState() {
	index := map[string]string{}
	scanner := bufio.NewScanner(drawing.NoErrorFile(os.Open("/tmp/checkpoint")))
	for scanner.Scan() {
		apikey := ""
		server := ""
		err := englang.Scanf(scanner.Text(), "Stateful item %s stored at %s server.", &apikey, &server)
		if err != nil {
			continue
		}
		index[apikey] = server
	}
	Index = index
}

func HttpRequest(url string, method string, bodyIn io.Reader) ([]byte, error) {
	if method == "" {
		method = "GET"
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
