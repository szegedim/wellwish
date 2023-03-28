package burst

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
)

func Setup() {
	http.HandleFunc("/burst", func(w http.ResponseWriter, r *http.Request) {
		if drawing.EnsureAPIKey(w, r) != nil {
			return
		}
		RunBurst(w, r)
	})
}

func RunBurst(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("apikey")
	if apiKey == "" {
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}
	burstCode, done := DownloadBurst(w, apiKey)
	if done {
		return
	}
	var stdin io.Reader
	if r.ContentLength != 0 {
		stdin = r.Body
	}
	Run(burstCode, stdin, w)
}

func DownloadBurst(w http.ResponseWriter, burst string) ([]byte, bool) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/sack?apikey=%s", metadata.SiteUrl, burst), nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	// Use a client not associated with the Server.
	var c http.Client
	resp, err := c.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return nil, true
	}
	burstCode, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusGone)
		return nil, true
	}
	_ = resp.Body.Close()

	return burstCode, false
}

func Run(code []byte, stdin io.Reader, stdout io.Writer) {
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
		Stdout: stdout,
		Stdin:  stdin,
	}

	err := cmd.Start()
	if err != nil {
		_, _ = stdout.Write([]byte(err.Error()))
		return
	}

	err = cmd.Wait()
	if err != nil {
		_, _ = stdout.Write([]byte(err.Error()))
		return
	}
	defer func(proc *os.Process) {
		if proc != nil {
			_ = proc.Kill()
		}
	}(cmd.Process)
}

func DebuggingInformation(w http.ResponseWriter, r *http.Request) {

}
