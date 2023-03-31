package mesh

import "io"

var Nodes = map[string]string{}

var Index = map[string]string{}

var MeshPattern = "Stateful item %s stored on %s server."

func LogSnapshot(m string, w io.Writer, r io.Reader) {

}
