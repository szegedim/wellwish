package server

import (
	"fmt"
	"gitlab.com/eper.io/engine/activation"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/burst"
	"gitlab.com/eper.io/engine/correspondence"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/entry"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"gitlab.com/eper.io/engine/mining"
	"gitlab.com/eper.io/engine/sack"
	"io"
	"net/http"
	"os"
	"strings"
)

func Main(args []string) {
	mesh.SetupRing()
	activation.SetupActivation()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if metadata.ActivationKey == "" {
			w.Header().Set("Location", "/index.html")
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			w.Header().Set("Location", "/activate.html")
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	})

	go func() {
		drawing.SetupDrawing()
	}()

	go setupSite()

	port := ":7777"

	if len(os.Args) > 1 {
		port = args[1]
	}
	if strings.HasPrefix(metadata.SiteUrl, "http://127.") {
		if strings.HasPrefix(metadata.NodeUrl, ":") {
			metadata.NodeUrl = fmt.Sprintf("http://127.0.0.1%s", port)
			metadata.SiteUrl = fmt.Sprintf("http://127.0.0.1%s", port)
		}
	}
	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println(err)
		fmt.Println("usage: go run main.go 127.0.0.1:7777")
	}
}

func setupSite() {
	<-activation.Activated

	administrationKey := management.SetupSiteManagement(func(m string, w io.Writer, r io.Reader) {
		management.LogSnapshot(m, w, r)
		activation.LogSnapshot(m, w, r)
		billing.LogSnapshot(m, w, r)
		mining.LogSnapshot(m, w, r)
		sack.LogSnapshot(m, w, r)
		burst.LogSnapshot(m, w, r)
		mesh.LogSnapshot(m, w, r)
	})
	activation.Activated <- administrationKey

	management.SetupSiteRoot()
	entry.Setup()
	sack.Setup()
	mining.Setup()
	drawing.SetupUploads()
	billing.SetupVoucher()
	billing.SetupCheckout()
	billing.SetupInvoice()
	correspondence.SetupCorrespondence()
}
