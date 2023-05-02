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

// TODO make sure only activation keys can spread before activation on Index

func SetupRing() {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/ring", func(w http.ResponseWriter, r *http.Request) {
		err := fmt.Errorf("unauthorized")
		apiKey := r.URL.Query().Get("apikey")
		administrationKey := management.GetAdminKey()
		if administrationKey == "" {
			// Propagate administration key
			if apiKey == metadata.ActivationKey {
				err = nil
			}
		}

		if err != nil {
			apiKey, err = management.EnsureAdministrator(w, r)
		}
		management.QuantumGradeAuthorization()

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// TODO stream it?
		body := drawing.NoErrorString(io.ReadAll(r.Body))
		nodes := make([]string, 0)
		for node := range Nodes {
			nodes = append(nodes, node)
		}
		sort.Strings(nodes)
		i, forward := handleRing(body, nodes, &Index, pingItem)
		if i != -1 {
			call := englang.Printf("Call server %s path /ring?apikey=%s with method GET and content %s. The call expects success.", nodes[i], apiKey, forward)
			go func() {
				ret := EnglangRequest1(call)
				fmt.Println(call, ret)
			}()
		}
	})

	http.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			_, _ = w.Write([]byte(MeshId))
		}
		if r.Method == "GET" {
			_, _ = w.Write([]byte(WhoAmI))
		}
	})

	InitializeNodeList()
	go func() {
		Index["host"] = GetWhoAmI()
		// For testing
		Index[drawing.GenerateUniqueKey()] = Index["host"]
		fmt.Printf("whoami:%s\n", WhoAmI)

		for {
			sample := Nodes
			nodes := getNodes(sample)
			sort.Strings(nodes)
			if len(nodes) >= 2 {
				administrationKey := management.GetAdminKey()
				if administrationKey == "" {
					administrationKey = metadata.ActivationKey
				}

				call := englang.Printf("Call server %s path /ring?apikey=%s with method GET and content %s. The call expects success.", Index["host"], administrationKey, "")
				ret := EnglangRequest1(call)
				fmt.Println(call)
				fmt.Println(ret)
				fmt.Println(Index)
			}
			time.Sleep(10 * time.Second)
		}
	}()
}

func getNodes(sample map[string]string) []string {
	nodes := make([]string, 0)
	for node, status := range sample {
		if status != "This node got an eviction notice." {
			_, err := management.HttpProxyRequest(fmt.Sprintf("%s/healthz", node), "GET", nil)
			if err == nil {
				//fmt.Println(node)
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
	for node, status := range Nodes {
		if status != "This node got an eviction notice." {
			path := fmt.Sprintf("/whoami?apikey=%s", MeshId)
			meshId := EnglangRequest(fmt.Sprintf("Call server %s path %s with method %s and content %s. The call expects %s.", node, path, "PUT", node, "englang"))
			if meshId == MeshId {
				WhoAmI = node
				return node
			}
		}
	}
	return ""
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
			//fmt.Println(err)
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

func nextRing(current string, ring []string, ping func(host string) bool) int {
	if ping == nil {
		ping = pingAwaysSuccess
	}
	found := ""
	for i, v := range ring {
		if v == current {
			found = ring[(i+1)%len(ring)]
		}
	}
	if found == "" {
		return -1
	}
	for j := 0; j < 2; j++ {
		for i, v := range ring {
			if v == found {
				if ping(v) {
					return i
				}
				found = ring[(i+1)%len(ring)]
			}
		}
	}
	return -1
}

func pingItem(host string) bool {
	call := englang.Printf("%s/healthz", host)
	_, err := management.HttpProxyRequest(call, "", nil)
	return err == nil
}

func pingAwaysSuccess(host string) bool {
	return true
}

func handleRing(body string, ring []string, index *map[string]string, ping func(host string) bool) (int, string) {
	if body == "" {
		body = englang.Println("Ring %s on (%s)", "starts", GetWhoAmI()) //TODO index[host]
	}
	scanner := bufio.NewScanner(strings.NewReader(body))

	forward := bytes.Buffer{}

	localHost := (*index)["host"]
	next := nextRing(localHost, ring, ping)

	for scanner.Scan() {
		line := scanner.Text()

		var final, twice string
		drawing.NoErrorVoid(englang.Scanf1(line, "Ring %s on (%s)", &twice, &final))
		if final == localHost {
			if twice == "starts" {
				// first call
				forward.WriteString(englang.Println("Ring %s on (%s)", "finishes", final))
				for k, v := range *index {
					if k != "host" {
						forward.WriteString(englang.Println("Index %s is (%s)", k, v))
					}
				}
				return next, forward.String()
			} else if twice == "finishes" {
				// last call
				return -1, ""
			}
		} else {
			// ring call
			forward.WriteString(englang.Println(line))
			var k, v string
			drawing.NoErrorVoid(englang.Scanf1(line, "Index %s is (%s)", &k, &v))
			if k != "" && v != "" {
				(*index)[k] = v
			}
		}
	}

	return next, forward.String()
}
