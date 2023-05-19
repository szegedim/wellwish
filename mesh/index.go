package mesh

import (
	"bufio"
	"bytes"
	"gitlab.com/eper.io/engine/englang"
	"io"
	"sync"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

var m sync.Mutex

func UpdateIndex(r io.Reader) {
	m.Lock()
	defer m.Unlock()

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		apikey := ""
		server := ""
		t := scanner.Text()
		err := englang.Scanf(t, MeshPattern, &apikey, &server)
		if err != nil {
			continue
		}
		index[apikey] = server
	}
}

func IndexUsed() bool {
	return len(index) > 0
}

func IndexLengthForTestingOnly() int {
	return len(index)
}

func GetIndex(apiKey string) string {
	m.Lock()
	defer m.Unlock()
	return index[apiKey]
}

func SetIndex(k string, v string) {
	m.Lock()
	defer m.Unlock()
	index[k] = v
}

func FilterIndexEntries() *bytes.Buffer {
	serializedIndex := bytes.Buffer{}
	index := index
	for apiKey, server := range index {
		serializedIndex.Write([]byte(englang.Printf(MeshPattern, apiKey, server) + "\n"))
	}
	return &serializedIndex
}

func englangMergeIndex(in string) string {
	UpdateIndex(bytes.NewBufferString(in))
	ret := FilterIndexEntries().String()
	return ret
}
