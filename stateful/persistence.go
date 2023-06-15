package stateful

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
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

// The persistence layer can be part of an office cluster.
// However, the rooms are stateless, so that they can scale fast up and down.
// This means that the backup solution has to be very quick.

// Also, there are numerous solutions for snapshots.
// We can actually just rely on the solution of the cloud vendor.
// Users can update metadata.DataRoot to store blobs in a more persistent directory
// A cloud vendor can just do a snapshot of the entire virtual machine.
// Restoring is just recreating the persistence machine.

// The architecture is a plain Englang map just like Redis.
// This diamond like architecture is a very common pattern.
// Consultants and developers can ramp up fast.

// We write in bursts (?) of key value pairs.
// We read individual key value pair items.
// Consistency is taken care of by the application layer of each module.
// We can just write items in the order as they come.

// We do not log much as it is a distraction
// Logs may be the source for hackers
// We expect the infrastructure to be ready

func SetupPersistence() {
	// This is running on an external service

	http.HandleFunc("/setkey", func(writer http.ResponseWriter, request *http.Request) {
		apiKey := request.URL.Query().Get("apikey")
		if apiKey == "" || metadata.ManagementKey != "" {
			time.Sleep(15 * time.Millisecond)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		metadata.ManagementKey = apiKey
	})

	http.HandleFunc("/persisted", func(writer http.ResponseWriter, request *http.Request) {
		StreamedBackup(writer, request)
	})

	http.HandleFunc("/snapshot", func(writer http.ResponseWriter, request *http.Request) {
		StreamedSnapshot(writer, request)
	})
}

func StreamedBackup(writer http.ResponseWriter, request *http.Request) {
	apiKey := request.URL.Query().Get("apikey")
	if apiKey == "" || apiKey != metadata.ManagementKey {
		time.Sleep(15 * time.Millisecond)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	apiKey = request.URL.Query().Get("key")
	if apiKey == "" {
		time.Sleep(15 * time.Millisecond)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	fileName := path.Join(metadata.DataRoot, apiKey)
	if request.Method == "PUT" || request.Method == "POST" || request.Method == "DELETE" {
		drawing.NoErrorVoid(os.Remove(fileName))
		content := bytes.NewBuffer(drawing.NoErrorBytes(io.ReadAll(request.Body)))
		if content.Len() == 0 {
			// Deletion request for GDPR, etc.
			return
		}
		drawing.NoErrorVoid(os.WriteFile(fileName, content.Bytes(), 0700))
		return
	}
	if request.Method == "GET" {
		time.Sleep(15 * time.Millisecond)
		body := bytes.NewBuffer(drawing.NoErrorBytes(os.ReadFile(fileName)))
		drawing.NoErrorWrite64(io.Copy(writer, body))
		return
	}
	if request.Method == "HEAD" {
		_, err := os.Stat(fileName)
		time.Sleep(15 * time.Millisecond)
		if err != nil {
			writer.WriteHeader(http.StatusNotFound)
		}
		return
	}
}

func StreamedSnapshot(writer http.ResponseWriter, request *http.Request) {
	apiKey := request.URL.Query().Get("apikey")
	if apiKey == "" || apiKey != metadata.ManagementKey {
		time.Sleep(15 * time.Millisecond)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}

	fileName := path.Join(metadata.DataRoot)
	list, err := os.ReadDir(fileName)
	if err != nil {
		time.Sleep(15 * time.Millisecond)
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	// A million entries will consume about 100 megabytes.
	// Snapshots will be rare, so this is fine for medium-sized organizational office clusters.
	// Alternatives are streaming from backup machines and distributing cloud snapshots.
	w := bufio.NewWriter(writer)
	for _, i := range list {
		drawing.NoErrorWrite(w.WriteString(fmt.Sprintln(i)))
	}
	drawing.NoErrorVoid(w.Flush())
}
