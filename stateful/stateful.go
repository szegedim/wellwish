package stateful

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Stateful modules use this module to persist their state
// The way this works is that they stream any changes out to a backup server
// Cloud specific solutions can do backups and restores.
// Snapshots, and restores are out of our scope as organizations like to unify this tasks.
// Enforcing a specific solution would limit our customer base.

func SetupStateful() {
	if len(stateModules) > 0 {
		http.HandleFunc("/backup", func(w http.ResponseWriter, r *http.Request) {
			_, err := management.EnsureAdministrator(w, r)
			if err != nil {
				return
			}
			wr := bufio.NewWriter(w)
			remoteCheckpoint := drawing.NoErrorResponse(http.Get(fmt.Sprintf("%s/snapshot?apikey=%s", metadata.StatefulBackupUrl, management.GetAdminKey())))
			scanner := bufio.NewScanner(remoteCheckpoint.Body)
			defer drawing.NoErrorVoid(remoteCheckpoint.Body.Close())
			for scanner.Scan() {
				remoteData := drawing.NoErrorResponse(http.Get(fmt.Sprintf("%s/persisted?apikey=%s", metadata.StatefulBackupUrl, scanner.Text())))
				drawing.NoErrorWrite64(io.Copy(wr, remoteData.Body))
				drawing.NoErrorVoid(remoteData.Body.Close())
			}
			drawing.NoErrorVoid(wr.Flush())
		})
	}

	if metadata.DataRoot != "" {
		// This is running on the server with large disk space
		SetupPersistence()
	}

	go func() {
		for {
			time.Sleep(checkpointPeriod)
			RegularCleanup()
		}
	}()
}

func RegularCleanup() {
	for i := range stateModules {
		lock.Lock()
		cleanupMemoryCache(stateModules[i], &lru)
		lock.Unlock()
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
	go func() {
		touchMemoryCache(&lru, k)
		drawing.NoErrorResponse(http.Post(fmt.Sprintf("%s/persisted?apikey=%s&key=%s", metadata.StatefulBackupUrl, metadata.ManagementKey, k), "text/plain", bytes.NewBufferString(v)))
	}()
}

func put(c *http.Client, url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

func GetStatefulItem(csi *map[string]string, kk string) string {
	lock.Lock()
	defer lock.Unlock()
	// local memory cache
	vvv, ok := (*csi)[kk]
	if !ok {
		// remote disk cache
		cleanupMemoryCache(csi, &lru)
		remoteData := drawing.NoErrorResponse(http.Get(fmt.Sprintf("%s/persisted?apikey=%s&key=%s", metadata.StatefulBackupUrl, metadata.ManagementKey, kk)))
		vvv = drawing.NoErrorString(io.ReadAll(remoteData.Body))
		(*csi)[kk] = vvv
		drawing.NoErrorVoid(remoteData.Body.Close())
	}
	touchMemoryCache(&lru, kk)
	return vvv
}
