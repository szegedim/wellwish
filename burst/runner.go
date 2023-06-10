package burst

import (
	"fmt"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/metadata"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func BoxCore() {
	participationKey := Curl(englang.Printf("curl -X GET http://127.0.0.1%s/idle?apikey=%s", metadata.Http11Port, metadata.ActivationKey), "")

	started := time.Now()
	for time.Now().Before(started.Add(MaxBurstRuntime * 4)) {
		command := Curl(englang.Printf("curl -X GET http://127.0.0.1%s/idle?apikey=%s", metadata.Http11Port, participationKey), "")
		if command == "success" {
			command = ""
		}
		if command != "" {
			// TODO
			//go func() {
			//	time.Sleep(MaxBurstRuntime)
			//	os.Exit(0)
			//}()
			result := RunExternalShell(command)
			fmt.Println(command, result)
			Curl(englang.Printf("curl -X PUT http://127.0.0.1%s/idle?apikey=%s", metadata.Http11Port, participationKey), result)
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
}
