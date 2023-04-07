package main

import (
	"fmt"
	"net/http"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://botanical.eper.io", http.StatusPermanentRedirect)
	})

	err := http.ListenAndServe(":7777", nil)
	if err != nil {
		fmt.Println(err)
	}
}
