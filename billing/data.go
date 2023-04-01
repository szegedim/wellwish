package billing

import (
	"bytes"
	"gitlab.com/eper.io/engine/englang"
	"io"
	"strconv"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

var orders = map[string]string{}

var vouchers = map[string]string{}

func LogSnapshot(m string, w io.Writer, r io.Reader) {
	if m == "GET" {
		for k, v := range orders {
			buf := bytes.NewBufferString("")
			bufv := []byte(v)
			buf.WriteString(englang.Printf("Record with type %s, apikey %s, and length %s bytes.", "order", k, strconv.FormatUint(uint64(len(bufv)), 10)))
			buf.Write(bufv)
			_, _ = w.Write(buf.Bytes())
		}
		for k, v := range vouchers {
			buf := bytes.NewBufferString("")
			bufv := []byte(v)
			buf.WriteString(englang.Printf("Record with type %s, apikey %s, and length %s bytes.", "voucher", k, strconv.FormatUint(uint64(len(bufv)), 10)))
			buf.Write(bufv)
			_, _ = w.Write(buf.Bytes())
		}
	}
	return
}

const TicketExpiry = "Validated until %s."
