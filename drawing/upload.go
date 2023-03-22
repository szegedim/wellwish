package drawing

import (
	"io"
	"net/http"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func SetupUploads() {
	http.HandleFunc("/upload.coin", func(w http.ResponseWriter, r *http.Request) {
		err := EnsureAPIKey(w, r)
		if err != nil {
			return
		}
		session := GetSession(w, r)
		dat, _ := io.ReadAll(r.Body)
		session.SignalUploaded(session, Upload{Body: dat, Type: r.Header.Get("Application-Binary-Type")})
	})
}
