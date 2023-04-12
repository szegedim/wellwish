package mesh

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"net/http"
	"testing"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func TestMesh(t *testing.T) {
	Nodes["http://127.0.0.1:7780"] = "active"
	Nodes["http://127.0.0.1:7781"] = "active"
	Nodes["http://127.0.0.1:7782"] = "active"

	Index[drawing.GenerateUniqueKey()] = "http://127.0.0.1:7780"
	Index[drawing.GenerateUniqueKey()] = "http://127.0.0.1:7781"
	Index[drawing.GenerateUniqueKey()] = "http://127.0.0.1:7782"

	SetupRing()

	go func() {
		time.Sleep(5 * time.Second)
		update := "<init>\n" + englang.Println(MeshPattern, drawing.GenerateUniqueKey(), "http://127.0.0.1:7780") + "<init>\n"
		EnglangRequest(fmt.Sprintf("Call server %s path %s with method %s and content %s. The call expects %s.", "http://127.0.0.1:7780", "/whoami", "GET", "<init>", "success"))
		time.Sleep(5 * time.Second)
		ret := EnglangRequest(englang.Printf("Call server http://127.0.0.1:7781 path /ring?apikey=INNABDBNSETETAKTRDOTNJSHFRKMKCQRCPRLMTNIBQPFAEESPNRPDEEIGLPNMPBC&ring=http://127.0.0.1:7781 with method GET and content %s. The call expects englang.", update))
		t.Log(ret)
	}()

	go func() { _ = http.ListenAndServe(":7780", nil) }()
	go func() { _ = http.ListenAndServe(":7781", nil) }()
	go func() { _ = http.ListenAndServe(":7782", nil) }()

	time.Sleep(20 * time.Second)
	if len(Index) != 4 {
		t.Error(Index)
	}
	fmt.Println(Index)
}
