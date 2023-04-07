package mesh

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"io"
	"net/http"
	"sort"
	"strings"
)

func SetupRing() {
	http.HandleFunc("/ring", func(w http.ResponseWriter, r *http.Request) {
		next := ""
		ring := r.URL.Query().Get("ring")
		last := r.URL.Query().Get("next")
		if last == ring {
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
			q := r.URL.String()
			begin, _, _ := strings.Cut(q, "&next=")
			q = fmt.Sprintf("%s&next=%s", begin, next)

			expected := "success"
			if r.ContentLength > 0 {
				expected = "englang"
			}

			buf := fmt.Sprintf("Ring running on %s\n", last)
			_, _ = w.Write([]byte(buf))
			body := drawing.NoErrorBytes(io.ReadAll(r.Body))
			_, _ = w.Write([]byte(EnglangHandler(string(body)) + "\n"))
			ret := Englang(fmt.Sprintf("Call server %s path %s with method %s and content %s. The call expects %s.", next, q, r.Method, body, expected))
			_, _ = w.Write([]byte(ret))
		}
	})

	http.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		if MeshId == "" {
			MeshId = drawing.GenerateUniqueKey()
		} else {
			body := r.URL.Query().Get("apikey")
			if MeshId == body {
				WhoAmI = string(drawing.NoErrorBytes(io.ReadAll(r.Body)))
				fmt.Println(WhoAmI)
				return
			}
		}
		for node, status := range Nodes {
			if status != "This node got an eviction notice." {
				path := fmt.Sprintf("/whoami?apikey=%s", MeshId)
				Englang(fmt.Sprintf("Call server %s path %s with method %s and content %s. The call expects %s.", node, path, "GET", node, "success"))
			}
		}
	})
}

func EnglangHandler(in string) string {
	return englang.Printf("Answer to %s is %s.", in, "<done>")
}

func Englang(e string) string {
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
