package mesh

import (
	"bufio"
	"fmt"
	"gitlab.com/eper.io/engine/management"
	"io"
)

var Nodes = map[string]string{}

var Rings = map[string]string{}

var Index = map[string]string{}

var NodePattern = ""

var MeshPattern = "Stateful item %s stored on %s server.\n"

func LogSnapshot(m string, w io.Writer, r io.Reader) {
	if m == "GET" && w != nil {
		ww := bufio.NewWriter(w)
		_, _ = ww.Write([]byte("\n"))
		for k, v := range Nodes {
			s := "unavailable"
			_, err := management.HttpProxyRequest(fmt.Sprintf("%s/healthz", k), "GET", nil)
			if err == nil {
				s = "ready"
			}
			_, _ = ww.WriteString(fmt.Sprintf("Node %s has status %s. Health result is %s\n", k, v, s))
		}
		for k, v := range Index {
			_, _ = ww.WriteString(fmt.Sprintf(MeshPattern, k, v))
		}
		for k, v := range Rings {
			_, _ = ww.WriteString(fmt.Sprintf("Ring %s has status %s.\n", k, v))
		}
		_ = ww.Flush()
	}
}
