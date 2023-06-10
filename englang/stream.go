package englang

import (
	"bufio"
	"bytes"
	"gitlab.com/eper.io/engine/drawing"
	"io"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func ReadIndexedEntry(r bufio.Reader) (string, string, string) {
	line, _ := r.ReadBytes('\n')
	var entity, key, lengths string
	if nil == Scanf1(string(line), "Indexed %s entity %s of bytes %s follows.\n", &entity, &key, &lengths) {
		length := Decimal(lengths)
		content := make([]byte, length)
		n, _ := io.ReadFull(&r, content)
		return entity, key, string(content[0:n])
	} else {
		return "", "", ""
	}
}

func WriteIndexedEntry(w bufio.Writer, entity string, k string, buf *bytes.Buffer) {
	drawing.NoErrorWrite(w.WriteString(Printf("Indexed %s entity %s of bytes %s follows.\n", entity, k, DecimalString(int64(buf.Len())))))
	_, _ = w.Write(buf.Bytes())
}
