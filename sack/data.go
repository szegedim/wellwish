package sack

import (
	"bytes"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"io"
	"strconv"
)

var Sacks = map[string]string{}

const RecordPattern = "Record with type %s, apikey %s, and length %s bytes."

const RecordType = "sack"

func LogSnapshot(m string, w io.Writer, r io.Reader) {
	if m == "GET" {
		for k, v := range Sacks {
			buf := bytes.NewBufferString("")
			bufv := []byte(v)
			buf.WriteString(englang.Printf(RecordPattern, RecordType, k, strconv.FormatUint(uint64(len(bufv)), 10)))
			buf.Write(bufv)
			_, _ = w.Write(buf.Bytes())
		}
	}
	if m == "PUT" {
		backup, err := io.ReadAll(r)
		if err != nil {
			return
		}
		management.RestoreRecords(backup, RecordPattern, RecordType, &Sacks, "")
		management.RestoreRecords(backup, RecordPattern, RecordType, nil, "/tmp")
	}
}
