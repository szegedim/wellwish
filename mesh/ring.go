package mesh

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
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

func SetupRing() {
	http.HandleFunc("/ring", func(w http.ResponseWriter, r *http.Request) {
		next := ""
		ring := r.URL.Query().Get("ring")
		last := r.URL.Query().Get("next")
		if last == ring {
			buf := fmt.Sprintf("Ring finished on %s%s\n", last, r.URL.String())
			_, _ = w.Write([]byte(buf))
			return
		}
		if last == "" {
			last = ring
		}

		sample := Nodes
		nodes := make([]string, 0)
		for node, status := range sample {
			if status != "This node got an eviction notice." {
				nodes = append(nodes, node)
			}
		}
		sort.Strings(nodes)

		for i := 0; i < len(nodes); i++ {
			if nodes[i] == last {
				next = nodes[(i+1)%len(nodes)]
				break
			}
		}

		if next != "" {
			RingNext(w, r, next, last)
		}
	})

	http.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(MeshId))
	})

	go func() {
		GetWhoAmI()
		fmt.Printf("whoami:%s\n", WhoAmI)

		for {
			update := "Propagating indexes\n"
			ret := EnglangRequest(englang.Printf("Call server %s path /ring?apikey=INNABDBNSETETAKTRDOTNJSHFRKMKCQRCPRLMTNIBQPFAEESPNRPDEEIGLPNMPBC&ring=%s with method GET and content %s. The call expects englang.", WhoAmI, WhoAmI, update))
			fmt.Println(ret)
			time.Sleep(10 * time.Second)
		}
	}()
}

func RingNext(w http.ResponseWriter, r *http.Request, next string, last string) bool {
	q := r.URL.String()
	begin, _, _ := strings.Cut(q, "&next=")
	q = fmt.Sprintf("%s&next=%s", begin, next)

	expected := "success"
	if r.ContentLength > 0 {
		expected = "englang"
	}

	buf := fmt.Sprintf("Ring running on %s %s%s\n", last, next, q)
	_, _ = w.Write([]byte(buf))
	body := string(drawing.NoErrorBytes(io.ReadAll(r.Body)))
	updateSent := Englang(body)
	//fmt.Println("<" + last + updateSent + ">")
	ret := EnglangRequest1(fmt.Sprintf("Call server %s path %s with method %s and content %s. The call expects %s.", next, q, r.Method, updateSent, expected))
	if ret == "" {
		return false
	}
	update := Englang(ret)
	_, _ = w.Write([]byte(fmt.Sprintf("%sThe call included server %s with %d indexes.\n", update, last, len(Index))))
	return true
}

func GetWhoAmI() string {
	if WhoAmI != "" {
		return WhoAmI
	}
	for node, status := range Nodes {
		if status != "This node got an eviction notice." {
			path := fmt.Sprintf("/whoami?apikey=%s", MeshId)
			meshId := EnglangRequest(fmt.Sprintf("Call server %s path %s with method %s and content %s. The call expects %s.", node, path, "GET", node, "englang"))
			if meshId == MeshId {
				WhoAmI = node
				return node
			}
		}
	}
	return ""
}

func Englang(in string) string {
	var ret, res string
	update := englangMergeIndex(in)
	if update != "" {
		ret = update
	}
	res = EnglangHealthz(in)
	if res != "" {
		ret = update + res
	}

	return ret
}

func EnglangHealthz(in string) string {
	scanner := bufio.NewScanner(strings.NewReader(in))
	ret := bytes.NewBufferString("")
	for scanner.Scan() {
		x := scanner.Text()
		if x == "Test ring code" {
			ret.WriteString(englang.Printf("Answer to %s is %s.\n", in, "Done"))
		}
		if strings.HasPrefix(x, "Ring finished on") {
			ret.WriteString(x + "\n")
		}
		if strings.HasPrefix(x, "Ring running on ") {
			ret.WriteString(x + "\n")
		}
		if strings.HasPrefix(x, "The call included server") {
			ret.WriteString(x + "\n")
		}
	}
	return ret.String()
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
			return ""
		}
		response, err := management.HttpProxyRequest(fmt.Sprintf("%s%s", server, path), method, strings.NewReader(content))
		if expect == "englang" {
			if err == nil {
				return string(response)
			} else {
				return ""
			}
		}
		return "success"
	}
	fmt.Println("I do not understand " + e)
	return ""
}
