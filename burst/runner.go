package burst

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
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
// an entry point. The locality is ensured by private keys distributed by mapped files called 'metal'.
// This ensures that we have a local runner.
// Each runner connects to the main site as /idle using the metal key
// The runner restarts after each run, so that any local state is lost
// Every burst fetches a new metal key from the file.

func SetupRunner() {
	InitializeNodeList()
}

func UpdateContainerWithBurst(apiKey string, update string) string {
	container := Container[apiKey]
	var metalfile, url, burstKey string
	if nil == englang.Scanf(container, ContainerPattern, &metalfile, &url, &burstKey) {
		container := fmt.Sprintf(ContainerPattern, metalfile, url, update)
		Container[apiKey] = container
		RestartFinishedContainer(apiKey, container)
	}
	return burstKey
}

func RestartFinishedContainer(key string, container string) {
	if strings.HasSuffix(container, "finished") {
		delete(Container, key)
		GenerateNewContainerKey(container)
		//TODO restart
		x := time.Duration(uint64(100000/len(Container))) * time.Microsecond
		time.Sleep(x)
	}
}

func GenerateNewContainerKey(container string) {
	currentKey := drawing.GenerateUniqueKey()
	Container[currentKey] = container
	var metalfile, url, burstKey string
	if nil == englang.Scanf(container, ContainerPattern, &metalfile, &url, &burstKey) {
		_ = os.WriteFile(metalfile, []byte(currentKey), 0700)
	}
}

func InitializeNodeList() {
	if MetalFilePattern != "" {
		MetalFiles = make([]string, 0)
		actual := []string{MetalFilePattern}
		for {
			next := make([]string, 0)
			for _, x := range actual {
				if strings.Contains(x, "*") {
					for i := 0; i < 10; i++ {
						next = append(next, strings.Replace(x, "*", englang.DecimalString(int64(i)), 1))
					}
				}
			}
			if len(next) == 0 {
				break
			}
			actual = next
		}
		for _, name := range actual {
			content := fmt.Sprintf(ContainerPattern, name, metadata.SiteUrl, "idle")
			GenerateNewContainerKey(content)
		}
		MetalFiles = actual
	}
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

func LogSnapshot(m string, w io.Writer, r io.Reader) {

}
