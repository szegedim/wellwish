package main

import (
	"fmt"
	"gitlab.com/eper.io/engine/activation"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/correspondence"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/entry"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"gitlab.com/eper.io/engine/mining"
	"gitlab.com/eper.io/engine/sack"
	"net/http"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// A simple billing experiment
func main() {
	drawing.SetupDrawing()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if metadata.ActivationKey == "" {
			w.Header().Set("Location", "/index.html")
			w.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			w.Header().Set("Location", "/activate.html")
			w.WriteHeader(http.StatusTemporaryRedirect)
		}
	})

	activation.SetupActivation()

	go func() {
		<-activation.Activated
		administrationKey := drawing.GenerateUniqueKey()

		management.SetupSiteManagement(administrationKey)
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
	}()
	err := http.ListenAndServe(":7777", nil)
	if err != nil {
		fmt.Println(err)
	}
}
