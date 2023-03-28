package sack

import (
	"bytes"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"io"
	"net/http"
	"strconv"
)

var Sacks = map[string]string{}

const RecordPattern = "Record with type %s, apikey %s, and length %s bytes."

const RecordType = "sack"

func DebuggingInformation(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		for k, v := range Sacks {
			buf := bytes.NewBufferString("")
			bufv := []byte(v)
			buf.WriteString(englang.Printf(RecordPattern, RecordType, k, strconv.FormatUint(uint64(len(bufv)), 10)))
			buf.Write(bufv)
			_, _ = w.Write(buf.Bytes())
		}
	}
	if r.Method == "PUT" {
		backup, err := io.ReadAll(r.Body)
		if err != nil {
			return
		}
		management.RestoreRecords(backup, RecordPattern, RecordType, &Sacks, "")
		management.RestoreRecords(backup, RecordPattern, RecordType, nil, "/tmp")
	}
}
