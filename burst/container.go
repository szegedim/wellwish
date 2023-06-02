package burst

import (
	"gitlab.com/eper.io/engine/burst/php"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/metadata"
	"os/exec"
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
// The big difference between these and other modules is that bursts do not have an api endpoint.

func RunExternalShell(task string) string {
	var ret string
	if task == "Idle." {
		return "Idle."
	}
	ret = php.EnglangPhp(drawing.GenerateUniqueKey(), task, MaxBurstRuntime+500*time.Millisecond)
	if ret != "" {
		return ret
	}
	task = runCommandInBox(task)
	if ret != "" {
		return ret
	}
	return "This is the result." + task
}

func runCommandInBox(task string) string {
	var command string
	if nil == englang.Scanf1(task+"DZPSOTHXAYZMZSJQEFMAD", "Run the following command line.%s"+"DZPSOTHXAYZMZSJQEFMAD", &command) {
		cmds := strings.Split(command, "")
		cmd := exec.Command(cmds[0], cmds[1:]...)
		go func() {
			if !metadata.Simplify {
				time.Sleep(MaxBurstRuntime + 500*time.Millisecond)
				if cmd.Process != nil {
					_ = cmd.Process.Kill()
				}
			}
		}()
		ret, _ := cmd.Output()
		task = string(ret)
	}
	return task
}

func FinishCleanup() {
	ContainerRunning = map[string]string{}
	BurstSession = map[string]string{}
}
