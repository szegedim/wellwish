package mining

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

import (
	"bytes"
	"gitlab.com/eper.io/engine/englang"
	"io"
	"strconv"
	"time"
)

var miningTicket = map[string]string{}

const ValidPeriod = 4 * 168 * time.Hour

func LogSnapshot(m string, w io.Writer, r io.Reader) {
	if m == "GET" {
		for k, v := range miningTicket {
			buf := bytes.NewBufferString("")
			bufv := []byte(v)
			buf.WriteString(englang.Printf("Record with type %s, apikey %s, and length %s bytes.", "miningticket", k, strconv.FormatUint(uint64(len(bufv)), 10)))
			buf.Write(bufv)
			_, _ = w.Write(buf.Bytes())
		}
	}
	return
}
