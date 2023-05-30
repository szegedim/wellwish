package main

import (
	"gitlab.com/eper.io/engine/burst"
	"gitlab.com/eper.io/engine/drawing"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Box is a container code that waits for a single burst and exits
// Box can run in a container launched as
// docker run -d --rm --restart=always --name box1 wellwish go run burst/box/main.go
// There are two ways to input and output data to and from boxes
// One is the burst `/run.coin` body request and return.
// Keep this small as it contains Englang that is duplicated in memory two times at least.
// The other way is to pass a sack url or cloud bucket url where the box streams any input or results.
// We do not log runtime or errors, the server takes care of that.
// TODO add timeout logic on paid vouchers

func main() {
	burst.ShutdownOnFinish = true
	drawing.NoErrorVoid(burst.RunBox())
}
