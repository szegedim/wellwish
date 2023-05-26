package burst

import (
	"fmt"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"gitlab.com/eper.io/engine/sack"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// This is a module code that runs burst containers.
// The big difference between these and other modules is that it actually does not have
// an entry point.
// The locality is ensured by private keys distributed early called 'metal'.
// This ensures that we have a local runner.
// What does this mean?
// - /idle responds to local endpoints only like a co-located container in the same pod
// - idle returns a task and a key to complete the task
// - malicious tasks may go for idle again
// - we protect against this by letting bursts run for a term e.g. ten seconds
// - we protect against this also by not issuing a new key until the previous one finishes
// - Each runner connects to the main site as /idle using the activation key
// - The activation key is deleted from the container once used
// - The init task of the container is our burst runner. It should be set debuggable by workload.
// - The init task kills the container, if the workload tries to kill it.
// - The final column is time fencing allowing /idle calls only once every minute when workloads are already gone.
// - The runner restarts after each run, so that any local state and code is lost disabling double /idle calls.

func SetupRunner() {
	fmt.Println("Initializing burst runners on 127.0.0.1")
}

func UpdateContainerWithBurst(containerKey string, update string) string {
	container := ContainerRunning[containerKey]
	var metalfile, url, burstKey string
	if nil == englang.Scanf(container, ContainerPattern, &metalfile, &url, &burstKey) {
		container := fmt.Sprintf(ContainerPattern, metalfile, url, update)
		ContainerRunning[containerKey] = container
		if strings.HasSuffix(container, "finished") {
			delete(ContainerRunning, containerKey)
		}
	}
	return burstKey
}

func Run(code []byte, stdin io.ReadCloser, stdout io.Writer) {
	if len(sack.Sacks) > 0 || mesh.IndexUsed() {
		_, _ = stdout.Write([]byte("isolation error running burst on sack/mesh"))
		return
	}
	goPath := path.Join(os.Getenv("GOROOT"), "bin", "go")
	workDir := path.Join("/tmp")
	mainGo := path.Join(workDir, "main.go")
	_ = os.WriteFile(mainGo, code, 0700)
	cmd := &exec.Cmd{
		Path: path.Join(goPath),
		Args: []string{"", "run", mainGo},
		Dir:  workDir,
		Env: []string{fmt.Sprintf("HOSTNAME=%s", metadata.SiteUrl),
			fmt.Sprintf("HOME=%s", workDir),
			fmt.Sprintf("GOROOT=%s", os.Getenv("GOROOT"))},
		Stderr: stdout,
		Stdin:  stdin,
	}

	stdoutRead, err := cmd.StdoutPipe()
	if err != nil {
		_, _ = stdout.Write([]byte(err.Error()))
		return
	}

	err = cmd.Start()
	if err != nil {
		_, _ = stdout.Write([]byte(err.Error()))
		return
	}

	defer func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	go func(r io.ReadCloser, w io.ReadCloser) {
		_, err = io.Copy(stdout, r)
		if err != nil {
			_, _ = stdout.Write([]byte(err.Error()))
			return
		}

		_ = r.Close()
		_ = w.Close()
	}(stdoutRead, stdin)

	err = cmd.Wait()
	if err != nil {
		_, _ = stdout.Write([]byte(err.Error()))
		return
	}
}

func runRunner(ready chan int, timeout time.Duration) {
	goRoot := os.Getenv("GOROOT")
	goroot := path.Join(goRoot, "bin", "go")
	box := "./burst/box1/main.go"
	_, err := os.Stat(box)
	if err != nil {
		fmt.Println("cannot find " + box)
	}
	fmt.Println(goroot)
	runC := []string{goroot, "help"}
	fmt.Println(strings.Join(runC, " "))
	p := exec.Cmd{
		Dir:  ".",
		Path: goroot,
		Args: runC,
		Env:  []string{"GOROOT=" + goRoot, "GOPATH=/tmp"},
	}
	err = p.Start()
	if err != nil {
		fmt.Println(err)
	}
	go func() {
		time.Sleep(timeout)
		_ = p.Process.Kill()
	}()
	err = p.Wait()
	if err != nil && err.Error() != "signal: killed" {
		fmt.Println(err)
	}
	b, _ := p.CombinedOutput()
	if len(b) > 0 {
		fmt.Println(string(b))
	}
	if p.ProcessState.ExitCode() != 0 && p.ProcessState.ExitCode() != -1 {
		fmt.Println(p.ProcessState.ExitCode())
	}
	ready <- 1
}

func LogSnapshot(m string, w io.Writer, r io.Reader) {

}
