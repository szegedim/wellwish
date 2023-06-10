package activation

import (
	"bufio"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/metadata"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Shows whether activation was done or not.
var ActivationNeeded = true

// Triggers enabling all features
var Activated = make(chan string)

// Remove activation key from logs
var ActivationHashLog = drawing.RedactPublicKey(metadata.ActivationKey)

// Protects against brute force attacks
var activationPeriod = time.Second

func LogSnapshot(m string, w bufio.Writer, r *bufio.Reader) {
	// Activation key is shared with multiple containers of the same version,
	// so we just return the record locator
	if m == "GET" {
		if metadata.ActivationKey == "" {
			_, _ = w.Write([]byte("The container is activated."))
		} else {
			_, _ = w.Write([]byte(fmt.Sprintf("This container is running with activation key as %s ...", ActivationHashLog)))
		}
	}
}
