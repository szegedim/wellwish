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
	"sort"
	"strings"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// TODO make sure only activation keys can spread before activation on index

func SetupRing() {
	//stateful.RegisterModuleForBackup(&index)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		write := bufio.NewWriter(w)
		drawing.NoErrorWrite(write.WriteString(IndexLengthForTestingOnly()))
		drawing.NoErrorVoid(write.Flush())
	})

	http.HandleFunc("/healthy", func(w http.ResponseWriter, r *http.Request) {
		write := bufio.NewWriter(w)
		drawing.NoErrorWrite(write.WriteString(metadata.Http11Port))
		drawing.NoErrorVoid(write.Flush())
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/ring", func(w http.ResponseWriter, r *http.Request) {
		var err error
		apiKey := r.URL.Query().Get("apikey")
		management.QuantumGradeAuthorization()
		if apiKey != metadata.ActivationKey {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		called := drawing.NoErrorString(io.ReadAll(r.Body))
		handleRingBody(called, &index)
	})

	http.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			_, _ = w.Write([]byte(HostId))
		}
		if r.Method == "GET" {
			_, _ = w.Write([]byte(WhoAmI))
		}
	})

	InitializeNodeList()
	go func() {
		time.Sleep(1 * time.Second)
		whoAmI := GetWhoAmI()
		if whoAmI == "" {
			fmt.Println("I do not know my own address. I will probably make errors.")
			fmt.Println("Fix this setting NODEPATTERN like 10.55.0.0/21, 127.0.0.1/32.")
		}
		index["host"] = whoAmI
		// For testing
		SetIndex(drawing.GenerateUniqueKey(), whoAmI)
		fmt.Printf("whoami:%s\n", whoAmI)

		for {
			body := prepareRingBody(&index)

			localHost := index["host"]
			nodes := getNodes(Nodes)
			sort.Strings(nodes)
			next := nextRingNode(localHost, nodes, pingItem)
			if next != "" {
				call := englang.Printf("Call server %s path /ring?apikey=%s with method PUT and content %s. The call expects success.", next, metadata.ActivationKey, string(body))
				ret := EnglangRequest1(call)
				if ret != "success" {
					fmt.Println(ret)
				}
			}

			time.Sleep(2 * time.Second)
		}
	}()
}

func getNodes(sample map[string]string) []string {
	nodes := make([]string, 0)
	for node, status := range sample {
		if status != "This node got an eviction notice." {
			_, err := management.HttpProxyRequest(fmt.Sprintf("%s/health", node), "GET", nil)
			if err == nil {
				nodes = append(nodes, node)
			}
		}
	}
	return nodes
}

func GetWhoAmI() string {
	if WhoAmI != "" {
		return WhoAmI
	}
	done := make(chan string)
	for node, status := range Nodes {
		if status != "This node got an eviction notice." {
			go func(current string, d chan string) {
				path := fmt.Sprintf("/whoami?apikey=%s", HostId)
				call := fmt.Sprintf("Call server %s path %s with method %s and content %s. The call expects %s.", current, path, "PUT", current, "englang")
				meshId := EnglangRequest(call)
				if meshId == HostId {
					WhoAmI = current
					d <- current
				}
			}(node, done)
		}
	}
	select {
	case <-done:
		break
	case <-time.After(5 * time.Second):
		break
	}
	if WhoAmI == "" {
		fmt.Printf("Could not identify mesh address from: %v\n", Nodes)
	}
	return WhoAmI
}

func EnglangRequest(e string) string {
	var server, path, method, content, expect string
	err := englang.Scanf(e, "Call server %s path %s with method %s and content %s. The call expects %s.", &server, &path, &method, &content, &expect)
	if err == nil {
		if expect != "success" && expect != "englang" {
			return ""
		}
		response, err := management.HttpProxyRequest(fmt.Sprintf("%s%s", server, path), method, strings.NewReader(content))
		if err != nil || expect == "englang" {
			return string(response)
		}
		return "success"
	}
	fmt.Println("I do not understand " + e)
	return "error"
}

func EnglangRequest1(e string) string {
	var server, path, method, content, expect string
	err := englang.Scanf(e, "Call server %s path %s with method %s and content %s. The call expects %s.", &server, &path, &method, &content, &expect)
	if err == nil {
		if expect != "success" && expect != "englang" {
			fmt.Println("invalid type")
			return ""
		}
		response, err := management.HttpProxyRequest(fmt.Sprintf("%s%s", server, path), method, strings.NewReader(content))
		if err != nil {
			return ""
		}
		if expect == "englang" {
			return string(response)
		}
		return "success"
	}
	fmt.Println("I do not understand " + e)
	return ""
}

func nextRingNode(current string, ring []string, ping func(host string) bool) string {
	if ping == nil {
		ping = pingDefault
	}
	found := false
	for i := 0; i < len(ring)*2; i++ {
		ix := i % len(ring)
		next := ring[ix]
		if found {
			if ping(next) {
				return next
			}
			if next == current {
				break
			}
			continue
		}
		if next == current {
			found = true
		}
	}
	return ""
}

func pingItem(host string) bool {
	call := englang.Printf("%s/health", host)
	_, err := management.HttpProxyRequest(call, "GET", nil)
	return err == nil
}

func pingDefault(host string) bool {
	return true
}

func handleRingBody(body string, index *map[string]string) {
	scanner := bufio.NewScanner(strings.NewReader(body))

	for scanner.Scan() {
		line := scanner.Text()
		var k, v string
		drawing.NoErrorVoid(englang.Scanf1(line, "index %s is (%s)", &k, &v))
		if k != "" && v != "" {
			(*index)[k] = v
		}
	}
}

func prepareRingBody(index *map[string]string) []byte {
	forward := bytes.Buffer{}
	for k, v := range *index {
		if k != "" && v != "" && k != "host" {
			forward.WriteString(englang.Println("index %s is (%s)", k, v))
		}
	}

	return forward.Bytes()
}
