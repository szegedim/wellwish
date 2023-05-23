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
	if len(stateModules) > 0 {
		http.HandleFunc("/snapshot", func(w http.ResponseWriter, r *http.Request) {
			_, err := management.EnsureAdministrator(w, r)
			if err != nil {
				return
			}
			_, _ = io.Copy(w, bytes.NewReader(*checkpoint))
		})

		go func() {
			// Memory side snapshot
			time.Sleep(1100 * time.Millisecond)
			for {
				snapshot := captureMemorySnapshot()
				byteSnapshot := snapshot.Bytes()
				checkpoint = &byteSnapshot
				time.Sleep(checkpointPeriod)
			}
		}()
	}

	if metadata.DataRoot != "" {
		// This is running on the server with large disk space
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
			// Disk side snapshot
			time.Sleep(1200 * time.Millisecond)
			for {
				remoteCheckpoint := drawing.NoErrorResponse(http.Get(fmt.Sprintf("%s/snapshot?apikey=%s", metadata.SiteUrl, management.GetAdminKey())))
				captureDiskSnapshot(remoteCheckpoint.Body)
				_ = remoteCheckpoint.Body.Close()

				time.Sleep(checkpointPeriod)
			}
		}()
	}
}

func RegisterModuleForBackup(module *map[string]string) {
	stateModules = append(stateModules, module)
}

func SetStatefulItem(csi *map[string]string, k string, v string) {
	lock.Lock()
	defer lock.Unlock()
	cleanupMemoryCache(csi, &lru)
	(*csi)[k] = v
	touchMemoryCache(&lru, k)
}

func GetStatefulItem(csi *map[string]string, kk string) string {
	lock.Lock()
	defer lock.Unlock()
	// local memory cache
	vvv, ok := (*csi)[kk]
	if !ok {
		// remote disk cache
		cleanupMemoryCache(csi, &lru)
		vvv = string(readStatefulItem(kk, false))
		(*csi)[kk] = vvv
	}
	touchMemoryCache(&lru, kk)
	return vvv
}

func readStatefulItem(apiKey string, local bool) []byte {
	if local {
		return drawing.NoErrorBytes(os.ReadFile(path.Join(metadata.DataRoot, apiKey)))
	} else {
		resp := drawing.NoErrorResponse(http.Get(fmt.Sprintf("%s/stateful?apikey=%s", metadata.StatefulBackupUrl, apiKey)))
		return drawing.NoErrorBytes(io.ReadAll(drawing.NoNilReader(resp)))
	}
}

func captureMemorySnapshot() bytes.Buffer {
	snapshot := bytes.Buffer{}
	for _, m := range stateModules {
		for kk, vv := range *m {
			snapshot.WriteString(englang.Printf("Index %s is set to %s length value of the string next.%s", kk, englang.DecimalString(int64(len(vv))), vv))
		}
	}
	return snapshot
}

func captureDiskSnapshot(rd io.Reader) {
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
