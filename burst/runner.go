package burst

import (
	"crypto/tls"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"gitlab.com/eper.io/engine/sack"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func SetupRunner() {
	// writes keys to metal files
	// listens to burst supply

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
	var stdin io.ReadCloser
	if r.ContentLength != 0 {
		stdin = r.Body
	}
	// You can run a burst, if you pay a sack underneath
	// TODO Change this to pay bursts separately.
	burstCode, err := DownloadBurst(apiKey)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if burstCode == nil {
		w.WriteHeader(http.StatusGone)
		return
	}
	Run(burstCode, stdin, w)
}

func DownloadBurst(burst string) ([]byte, error) {
	ret, err := DownloadCode(fmt.Sprintf("%s/sack?apikey=%s", metadata.SiteUrl, burst))
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func DownloadCode(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Use a client not associated with the Server.
	var c http.Client
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	burstCode, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	_ = resp.Body.Close()

	return burstCode, nil
}

func Run(code []byte, stdin io.ReadCloser, stdout io.Writer) {
	if len(sack.Sacks) > 0 || len(mesh.Index) > 0 {
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

func BurstRunner(client string, codeUrl string) {
	burstKey := drawing.GenerateUniqueKey()
	metalKey := drawing.NoErrorString(io.ReadAll(drawing.NoErrorFile(os.Open("/tmp/apikey"))))
	if len(metalKey) != len(burstKey) {
		fmt.Println("missing /tmp/apikey")
	}
	// Get key (same machine?)
	a, err := net.ResolveTCPAddr("tcp", client)
	if err != nil {
		fmt.Println(err)
		return
	}
	tc, err := net.DialTCP("tcp", nil, a)
	if err != nil {
		fmt.Println(err)
		return
	}
	w, r := TlsClient(tc, client)

	// So I do not check for errors on the stream...
	// Does the server need to check for errors anyway? Yes.
	// Is it cheaper to have one fallback solution docker --restart=always? Yes.
	_, err = w.Write([]byte(metalKey))
	if err != nil {
		fmt.Println(err)
		return
	}

	dummyCode, err := DownloadCode(codeUrl)
	if err != nil {
		fmt.Println(err)
		return
	}
	Run(dummyCode, r, w)
	_ = w.Close()
}

func TlsServer(c net.Conn, client string) (io.WriteCloser, io.Reader) {
	if strings.HasPrefix(client, "127") {
		return io.WriteCloser(c), io.Reader(c)
	}
	var ts = tls.Server(c, tlsConfig(client))
	return io.WriteCloser(ts), io.Reader(ts)
}

func TlsClient(c net.Conn, client string) (io.WriteCloser, io.ReadCloser) {
	if strings.HasPrefix(client, "127") {
		return io.WriteCloser(c), io.ReadCloser(c)
	}
	var ts = tls.Client(c, tlsConfig(client))
	return io.WriteCloser(ts), io.ReadCloser(ts)
}

func tlsConfig(server string) *tls.Config {
	c := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         server,
		CipherSuites:       []uint16{tls.TLS_RSA_WITH_AES_128_CBC_SHA256, tls.TLS_RSA_WITH_AES_128_GCM_SHA256},
	}
	c.MinVersion = tls.VersionTLS13
	return c
}
