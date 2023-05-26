package php

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

func IsPhpAvailable() bool {
	_, err := os.Stat(PhpPath)
	return err == nil
}

func EnglangPhp(key string, code string, timeout time.Duration) string {
	if strings.HasPrefix(code, "Run the following php code.") {
		php := code[len("Run the following php code."):]
		if php == MockPhp {
			return MockPhpResult
		}
		return englangPhp(key, php, timeout)
	}
	return ""
}

func englangPhp(key string, code string, timeout time.Duration) string {
	php := path.Join("/tmp", key)

	_ = os.WriteFile(php, []byte(code), 0700)

	cmd := exec.Command(PhpPath, php)
	output, err := cmd.Output()
	if err != nil {
		output = []byte(err.Error())
	} else {
		go func() {
			time.Sleep(timeout)
			_ = cmd.Process.Kill()
		}()
	}
	if len(output) == 0 {
		return "No php result returned."
	}

	return string(output)
}
