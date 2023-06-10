package server

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/activation"
	"gitlab.com/eper.io/engine/bag"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/burst"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/entry"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"gitlab.com/eper.io/engine/mining"
	"gitlab.com/eper.io/engine/stateful"
	"io"
	"net/http"
	"strings"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Main server
// - It starts up a page where you can activate with a key in metadata/data.go ActivationKey
// - It propagates to local servers matching the pattern NodePattern
// - The service is redirected to a management and administration page using a new api key
// - Bookmark this page with the admin key to return and get backup or traces
// - It has a pointer to public pages that can create bags, bursts, etc.

// ## Design.
// Why activate?
// Such a solution allows a unique private distribution to be propagated and started in a clean state.
// The activation key can also help to give access to clusters in a few hundred milliseconds to customers who just paid.

func Main(args []string) {
	port := metadata.Http11Port
	port = customizePort(args, port)
	metadata.Http11Port = port

	go func() {
		drawing.SetupDrawing()
	}()

	stateful.SetupStateful()
	mesh.SetupRing()
	activation.SetupActivation()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !activation.ActivationNeeded {
			w.Header().Set("Location", "/index.html")
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			if strings.HasSuffix(r.URL.Path, "html") || r.URL.Path == "/" {
				w.Header().Set("Location", "/activate.html")
				w.WriteHeader(http.StatusTemporaryRedirect)
			} else {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	})

	go setupSite()

	err := http.ListenAndServe(port, nil)
	printUsage(err)
}

func customizePort(args []string, port string) string {
	if len(args) > 1 {
		port = args[1]
	}
	return port
}

func printUsage(err error) {
	if err != nil {
		fmt.Println(err)
		fmt.Println("usage: go run main.go :7777")
	}
}

func setupSite() {
	<-activation.Activated
	management.SetupSiteManagement(func(m string, w bufio.Writer, r io.Reader) {
		fullRestore := bytes.NewBuffer(drawing.NoErrorBytes(io.ReadAll(r)))
		// We could try in parallel, but we will probably be ram bound anyway.
		activation.LogSnapshot(m, w, bufio.NewReader(fullRestore))
		management.LogSnapshot(m, w, bufio.NewReader(fullRestore))
		billing.LogSnapshot(m, w, bufio.NewReader(fullRestore))
		mining.LogSnapshot(m, w, bufio.NewReader(fullRestore))
		bag.LogSnapshot(m, w, bufio.NewReader(fullRestore))
		burst.LogSnapshot(m, w, bufio.NewReader(fullRestore))
		mesh.LogSnapshot(m, w, bufio.NewReader(fullRestore))
	})
	activation.Activated <- "Hello Moon!"

	management.SetupSiteRoot()

	// It is reliable to do these in order
	entry.Setup()
	mesh.SetupExpiry()
	bag.Setup()
	burst.Setup()
	mining.Setup()
	drawing.SetupUploads()
	billing.Setup()
}
