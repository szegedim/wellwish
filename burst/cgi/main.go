package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Hello World!")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://botanical.eper.io", http.StatusPermanentRedirect)
	})

	err := http.ListenAndServe(":7777", nil)
	if err != nil {
		fmt.Println(err)
	}
}
