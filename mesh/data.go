package mesh

import (
	"bufio"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/management"
	"io"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

var MeshId = drawing.GenerateUniqueKey()

var WhoAmI = ""

var Nodes = map[string]string{}

var Rings = map[string]string{}

var index = map[string]string{}

var NodePattern = ""

var MeshPattern = "Stateful item %s is stored on %s server."

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
		index := index
		for k, v := range index {
			_, _ = ww.WriteString(fmt.Sprintf(MeshPattern, k, v) + "\n")
		}
		for k, v := range Rings {
			_, _ = ww.WriteString(fmt.Sprintf("Ring %s has status %s.\n", k, v))
		}
		_ = ww.Flush()
	}
}
