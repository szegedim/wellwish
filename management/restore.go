package management

import (
	"bytes"
	"gitlab.com/eper.io/engine/englang"
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

func RestoreRecords(backup []byte, pattern string, recordType string, items *map[string]string, directory string) {
	prefix := []byte(englang.ScanfPrefix(pattern))
	suffix := []byte(englang.ScanfSuffix(pattern))
	i := 0
	for {
		b := bytes.Index(backup[i:], prefix)
		if b == -1 {
			return
		}
		b = i + b
		e := bytes.Index(backup[b:], suffix)
		if e == -1 {
			return
		}
		e = b + e + len(suffix)
		slice := backup[b:e]
		var typeR string
		var apiKey string
		var length string
		if englang.Scanf(string(slice), pattern, &typeR, &apiKey, &length) != nil {
			return
		}
		n, err := strconv.ParseInt(length, 10, 64)
		if typeR == recordType &&
			len(apiKey) > 0 &&
			err == nil {
			v := backup[e : e+int(n)]
			if items != nil {
				(*items)[apiKey] = string(v)
			}
			if directory != "" {
				path1 := path.Join(directory, apiKey)
				_ = os.WriteFile(path1, v, 0700)
			}
		}
		if err == nil {
			i = e + int(n)
		} else {
			i = e
		}
	}
}
