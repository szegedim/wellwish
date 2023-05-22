package stateful

import (
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
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

func SetupStateful() {
	http.HandleFunc("/snapshot", func(w http.ResponseWriter, r *http.Request) {
		_, err := management.EnsureAdministrator(w, r)
		if err != nil {
			return
		}
		_, _ = io.Copy(w, bytes.NewReader(*checkpoint))
	})

	http.HandleFunc("/stateful", func(writer http.ResponseWriter, request *http.Request) {
		apiKey := request.URL.Query().Get("apikey")
		if apiKey == "" {
			return
		}
		time.Sleep(15 * time.Millisecond)
		// Remote disk cache local stub
		content := readStatefulItem(apiKey, true)
		_, _ = io.Copy(writer, bytes.NewReader(content))
	})

	go func() {
		// Stateless side snapshot
		time.Sleep(3 * time.Second)
		for {
			snapshot := saveSnapshot()
			byteSnapshot := snapshot.Bytes()
			checkpoint = &byteSnapshot

			time.Sleep(checkpointPeriod)
		}
	}()

	go func() {
		// Stateful side snapshot
		time.Sleep(6 * time.Second)
		for {
			// TODO admin key
			remoteCheckpoint := drawing.NoErrorResponse(http.Get(fmt.Sprintf("%s/snapshot?apikey=%s", metadata.StatefulBackupUrl, management.GetAdminKey())))
			loadSnapshot(remoteCheckpoint.Body)
			_ = remoteCheckpoint.Body.Close()

			time.Sleep(checkpointPeriod)
		}
	}()

	go func() {
		// Clean cache periodically even if there is no input
		time.Sleep(1 * time.Second)
		for {
			lock.Lock()
			cleanupMemoryCache(&csi, &lru)
			lock.Unlock()
			time.Sleep(checkpointPeriod)
		}
	}()
}

func readStatefulItem(apiKey string, local bool) []byte {
	if local {
		return drawing.NoErrorBytes(os.ReadFile(path.Join(metadata.DataRoot, apiKey)))
	} else {
		resp := drawing.NoErrorResponse(http.Get(fmt.Sprintf("%s/stateful?apikey=%s", metadata.StatefulBackupUrl, apiKey)))
		return drawing.NoErrorBytes(io.ReadAll(resp.Body))
	}
}

func SetStatefulItem(k string, v string) {
	lock.Lock()
	defer lock.Unlock()
	cleanupMemoryCache(&csi, &lru)
	csi[k] = v
	touch(&csi, &lru, k)
}

func GetStatefulItem(kk string) string {
	lock.Lock()
	defer lock.Unlock()
	// local memory cache
	vvv, ok := csi[kk]
	if !ok {
		// remote disk cache
		cleanupMemoryCache(&csi, &lru)
		vvv = string(readStatefulItem(kk, false))
		csi[kk] = vvv
	}
	touch(&csi, &lru, kk)
	return vvv
}

func saveSnapshot() bytes.Buffer {
	snapshot := bytes.Buffer{}
	for kk, vv := range csi {
		snapshot.WriteString(englang.Printf("Index %s is set to %s length value of the string next.%s", kk, englang.DecimalString(int64(len(vv))), vv))
	}
	return snapshot
}

func loadSnapshot(rd io.Reader) {
	for {
		index := englang.ReadWith(rd, " length value of the string next.")
		if index == "" {
			break
		}
		var apiKey string
		var length string
		err := englang.Scanf1(index, "Index %s is set to %s length value of the string next.", &apiKey, &length)
		if err == nil {
			contentLength := englang.Decimal(length)
			rdl := io.LimitReader(rd, contentLength)
			content, _ := io.ReadAll(rdl)
			//fmt.Println(string(apiKey), string(content))
			drawing.NoErrorVoid(os.WriteFile(path.Join(metadata.DataRoot, apiKey), content, 0700))
		}
	}
}

func touch(data *map[string]string, lru *map[string]string, apiKey string) {
	(*lru)[apiKey] = timeStamp()
}

func timeStamp() string {
	return englang.DecimalString(int64(time.Now().Sub(startupTime).Seconds() + 0.01))
}

func cleanupMemoryCache(data *map[string]string, lru *map[string]string) {
	for i := 0; i < 2; i++ {
		if len(*lru) >= containerIndexLimit {
			// Make some space
			lrutime := ""
			lrukey := ""
			size := 0
			for k, v := range *lru {
				size = size + len(v) + len(k)
				if lrutime == "" || v < lrutime {
					lrukey = k
					lrutime = v
				}
			}
			if lrutime != "" {
				// Make sure we are backed up
				if (*data)[lrukey] == string(readStatefulItem(lrukey, false)) {
					fmt.Printf("Memory usage before cleanup: %d %s\n", size, lrukey)
					delete(*data, lrukey)
					delete(*lru, lrukey)
				} else {
					// This will rarely happen, when the container is almost full
					// and the backup server is slow or offline
					fmt.Printf("out of memory")
					time.Sleep(checkpointPeriod)
				}
			}
		}
	}
}
