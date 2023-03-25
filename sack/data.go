package sack

import (
	"bytes"
	"gitlab.com/eper.io/engine/englang"
	"net/http"
	"strconv"
)

var Sacks = map[string]string{}

func DebuggingInformation(w http.ResponseWriter, r *http.Request) {
	for k, v := range Sacks {
		buf := bytes.NewBufferString("")
		bufv := []byte(v)
		buf.WriteString(englang.Printf("Record with type %s, apikey %s, and length %s bytes.", "sack", k, strconv.FormatUint(uint64(len(bufv)), 10)))
		buf.Write(bufv)
		_, _ = w.Write(buf.Bytes())
	}
	return
}
