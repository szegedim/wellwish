package mesh

import (
	"gitlab.com/eper.io/engine/metadata"
	"testing"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func TestCidr(t *testing.T) {
	metadata.NodePattern = ""

	InitializeNodeList()

	metadata.NodePattern = "10.55.0.0/21"
	InitializeNodeList()

	if len(Nodes) != 2048 {
		t.Error("cidr not parsed", len(Nodes))
	}
}
