package mesh

import (
	"gitlab.com/eper.io/engine/metadata"
	"testing"
)

func TestCidr(t *testing.T) {
	metadata.NodePattern = ""

	InitializeNodeList()

	metadata.NodePattern = "10.55.0.0/21"
	InitializeNodeList()

	if len(Nodes) != 2048 {
		t.Error("cidr not parsed", len(Nodes))
	}
}
