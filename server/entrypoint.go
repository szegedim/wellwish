package server

import (
	"fmt"
	"gitlab.com/eper.io/engine/activation"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/burst"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/entry"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"gitlab.com/eper.io/engine/mining"
	"gitlab.com/eper.io/engine/sack"
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
// - It has a pointer to public pages that can create sacks bursts, etc.

// ## Design.
// Why activate?
// Such a solution allows a unique private distribution to be propagated and started in a clean state.
// The activation key can also help to open up clusters in a few hundred milliseconds to customers who just paid.

func Main(args []string) {
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

	port := metadata.Http11Port

	port = customizePort(args, port)
	err := http.ListenAndServe(port, nil)
	printUsage(err)
}

func customizePort(args []string, port string) string {
	if len(args) > 1 {
		port = args[1]
	}
	if strings.HasPrefix(metadata.SiteUrl, "http://127.") {
		if strings.HasPrefix(metadata.SiteUrl, ":") {
			metadata.SiteUrl = fmt.Sprintf("http://127.0.0.1%s", port)
		}
	}
	return port
}

func printUsage(err error) {
	if err != nil {
		fmt.Println(err)
		fmt.Println("usage: go run main.go 127.0.0.1:7777")
	}
}

func setupSite() {
	<-activation.Activated
	management.SetupSiteManagement(func(m string, w io.Writer, r io.Reader) {
		management.LogSnapshot(m, w, r)
		activation.LogSnapshot(m, w, r)
		billing.LogSnapshot(m, w, r)
		mining.LogSnapshot(m, w, r)
		sack.LogSnapshot(m, w, r)
		burst.LogSnapshot(m, w, r)
		mesh.LogSnapshot(m, w, r)
	})
	activation.Activated <- "Hello Moon!"

	management.SetupSiteRoot()

	// It is reliable to do these in order
	entry.Setup()
	sack.Setup()
	mining.Setup()
	drawing.SetupUploads()
	billing.SetupVoucher()
	billing.SetupCheckout()
	billing.SetupInvoice()
}
