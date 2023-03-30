package main

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Box is a container code that waits for a single burst and exits
// Box can run in a container launched as
// docker run -d --rm --restart=always --name box1 wellwish go run burst/box/main.go

func main() {
	// We do not log runtime or errors, the server takes care of that.
	// TODO add timeout logic on paid vouchers
	// burst.Burst(metadata.SiteUrl)
}
