package bag

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"io"
	"os"
	"path"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

var bags = map[string]string{}

const ValidPeriod = 168 * time.Hour

func LogSnapshot(m string, w *bufio.Writer, r *bufio.Reader) {
	if m == "GET" {
		for k, v := range bags {
			englang.WriteIndexedEntry(w, k, "bag", bytes.NewBufferString(v))
		}
	}
	if m == "PUT" {
		for {
			e, k, v := englang.ReadIndexedEntry(*r)
			if k == "" {
				return
			}
			if e == "bag" {
				bags[k] = v
			}
		}
	}
	// Bags are special with binary data at the end to support debugging.
	logBinaries(m, w, r)
}

func logBinaries(m string, w *bufio.Writer, r *bufio.Reader) {
	if m == "GET" {
		for k, _ := range bags {
			bag := k
			filePath := path.Join(fmt.Sprintf("/tmp/%s", bag))
			binaryData := drawing.NoErrorFile(os.Open(filePath))
			var length int64
			stat, _ := binaryData.Stat()
			if stat != nil {
				length = stat.Size()
			}
			drawing.NoErrorWrite(w.WriteString(englang.Printf("Indexed entity %s of bytes %s follows.\n", k, englang.DecimalString(length))))
			drawing.NoErrorWrite64(w.ReadFrom(binaryData))
		}
	}
	if m == "PUT" {
		for {
			line, _ := r.ReadBytes('\n')
			var bag, lengths string
			if nil == englang.Scanf1(string(line), "Indexed entity %s of bytes %s follows.\n", &bag, &lengths) {
				length := englang.Decimal(lengths)
				content := make([]byte, length)
				n, _ := io.ReadFull(r, content)
				filePath := path.Join(fmt.Sprintf("/tmp/%s", bag))
				_ = os.WriteFile(filePath, content[0:n], 0700)
			} else {
				return
			}
		}
	}
}
