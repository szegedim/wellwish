package activation

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/metadata"
	"io"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

var Activated = make(chan string)

var ActivationHash = drawing.RedactPublicKey(metadata.ActivationKey)

func LogSnapshot(m string, w io.Writer, r io.Reader) {
	// Activation key is shared with multiple containers of the same version,
	// so we just return the record locator
	if m == "GET" {
		_, _ = w.Write([]byte(fmt.Sprintf("This container is running with activation key as %s ...", ActivationHash)))
	}
}
