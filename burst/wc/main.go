package main

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"os"
	"path"
	"strings"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

var wc = 0
var lc = 0

func main() {
	arr := make([]string, 0)
	Readdir(&arr, ".")
	fmt.Printf("wc:%d\n", wc)
	fmt.Printf("lc:%d\n", lc)
}

func Readdir(arr *[]string, s string) {
	items1, err := os.ReadDir(s)
	if err != nil {
		return
	}
	for _, vv := range items1 {
		p := path.Join(s, vv.Name())
		if vv.IsDir() && !strings.HasPrefix(vv.Name(), ".") {
			Readdir(arr, p)
		} else {
			*arr = append(*arr, p)
			fmt.Println(p)
			if strings.HasSuffix(p, ".go") {
				q := drawing.NoErrorString(os.ReadFile(p))
				wc = wc + len(strings.Split(q, " "))
				lc = lc + len(strings.Split(q, "\n"))
			}
		}
	}
}
