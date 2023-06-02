package burst

import (
	"fmt"
	"gitlab.com/eper.io/engine/metadata"
	"os"
	"os/exec"
	"path"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Box is a container code that waits for a single burst and exits
// Box can run in a container launched as
// docker run -d --rm --restart=always --name box1 wellwish go run burst/box/main.go
// There are two ways to input and output data to and from boxes
// One is the burst `/run` request body and return.
// Keep this small as it transfers through requests.
// The other way is to pass a bag url or cloud bucket url where the box streams any input or results.
// We do not log runtime or errors, the server takes care of that.
// The design is that it runs for 1-10 seconds.
// This is the bandwidth of a single 1 vcpu+1 gigabyte container streamed entirely.
// It is similar to serverless lambdas, but it is a bit better.
// Lambdas can wait but bursts typically use cpu bursts and they continue in another burst.
// This makes bursts less expensive to the cloud provider using the cpu as much as possible.
// Also, bursts are not called but they run.
// They are not an api endpoint, the api gateway is the wellwish server.
// This makes burst more secure and easier to use just like a bash script.
// Bursts are typically docker containers with php/java/node preloaded by the taste of the cloud farm.
// They keep checking the frontend for new tasks and they restart when done.

func RunBox() error {
	for {
		time.Sleep(100 * time.Millisecond)

		goRoot := os.Getenv("GOROOT")
		goroot := path.Join(goRoot, "bin", "go")
		p := path.Join("..", "burst", "box", "main.go")
		fmt.Println(p, metadata.Http11Port)
		err := exec.Command(goroot, "run", p, metadata.Http11Port).Run()
		if err != nil {
			fmt.Println("local result", err)
		}
	}
}
