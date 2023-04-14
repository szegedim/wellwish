package sack

import (
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/englang"
	"io"
	"os"
	"path"
	"strconv"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

var Sacks = map[string]string{}

const RecordPattern = "Record with type %s, apikey %s, info %s, file name %s and length of %s bytes."

const RecordType = "sack"

func LogSnapshot(m string, w io.Writer, r io.Reader) {
	if m == "GET" {
		for k, v := range Sacks {
			sack := k
			filePath := path.Join(fmt.Sprintf("/tmp/%s", k))
			info := v

			content, _ := os.ReadFile(filePath)

			buf := bytes.NewBuffer([]byte{})
			buf.WriteString(englang.Printf(RecordPattern, RecordType, sack, info, filePath, strconv.FormatInt(int64(len(content)), 10)))
			buf.Write(content)
			_, _ = w.Write(buf.Bytes())
		}
	}
	if m == "PUT" {
		buf, err := io.ReadAll(r)
		if err != nil {
			return
		}
		i := 0
		for {
			record := ""
			sack := ""
			info := ""
			filePath := ""
			length := ""
			n, err := englang.ScanfStream(buf, i, RecordPattern, &record, &sack, &info, &filePath, &length)
			if err != nil {
				break
			}
			l, err := strconv.ParseInt(length, 10, 32)
			if err != nil {
				break
			}
			// Storing the length ensures to avoid Englang injection vulnerabilities
			// if the file contains Englang of sacks.
			Sacks[sack] = info
			if l > 0 {
				_ = os.WriteFile(path.Join("/tmp", sack), buf[n:n+int(l)], 0700)
			}
			i = n
		}
	}
}
