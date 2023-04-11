package mesh

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestMesh(t *testing.T) {
	Nodes["http://127.0.0.1:7780"] = "active"
	Nodes["http://127.0.0.1:7781"] = "active"
	Nodes["http://127.0.0.1:7782"] = "active"
	SetupRing()
	go func() {
		time.Sleep(5 * time.Second)
		Englang(fmt.Sprintf("Call server %s path %s with method %s and content %s. The call expects %s.", "http://127.0.0.1:7780", "/whoami", "GET", "<init>", "success"))
		time.Sleep(5 * time.Second)
		ret := Englang("Call server http://127.0.0.1:7781 path /ring?apikey=INNABDBNSETETAKTRDOTNJSHFRKMKCQRCPRLMTNIBQPFAEESPNRPDEEIGLPNMPBC&ring=http://127.0.0.1:7781 with method GET and content <init>. The call expects englang.")
		t.Log(ret)
	}()

	go func() { _ = http.ListenAndServe(":7780", nil) }()
	go func() { _ = http.ListenAndServe(":7781", nil) }()
	go func() { _ = http.ListenAndServe(":7782", nil) }()

	time.Sleep(20 * time.Second)
}
