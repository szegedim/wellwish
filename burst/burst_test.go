package burst

import (
	"bytes"
	"gitlab.com/eper.io/engine/drawing"
	"io"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	code, _ := io.ReadAll(drawing.NoErrorFile(os.Open("./cgi/main.go")))
	in := bytes.NewBufferString("Hello Burst!")
	out := bytes.NewBufferString("")
	Run(code, in, out)
	t.Log(out.String())
}
