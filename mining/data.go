package mining

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

import (
	"bufio"
	"bytes"
	"gitlab.com/eper.io/engine/englang"
	"time"
)

var miningTicket = map[string]string{}

const ValidPeriod = 4 * 168 * time.Hour

func LogSnapshot(m string, w *bufio.Writer, r *bufio.Reader) {
	if m == "GET" {
		for k, v := range miningTicket {
			englang.WriteIndexedEntry(w, k, "mining", bytes.NewBufferString(v))
		}
	}
	if m == "PUT" {
		for {
			e, k, v := englang.ReadIndexedEntry(*r)
			if k == "" {
				return
			}
			if e == "mining" {
				miningTicket[k] = v
			}
		}
	}
}
