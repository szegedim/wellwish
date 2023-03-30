package mesh

import (
	"gitlab.com/eper.io/engine/drawing"
	"io"
	"os"
	"testing"
)

func TestName(t *testing.T) {
	_ = os.Remove("/tmp/a")
	_ = os.Remove("/tmp/b")
	_ = os.Remove("/tmp/checkpoint")
	_ = os.WriteFile("/tmp/a", []byte("abc"), 0700)
	_ = os.Link("/tmp/a", "/tmp/checkpoint")

	_ = os.WriteFile("/tmp/b", []byte("def"), 0700)
	_ = os.Remove("/tmp/checkpoint")
	_ = os.Link("/tmp/b", "/tmp/checkpoint")

	x, _ := io.ReadAll(drawing.NoErrorFile(os.Open("/tmp/checkpoint")))
	if string(x) != "def" {
		t.Error(string(x))
	}
}
