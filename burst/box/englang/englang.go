package box

import (
	"bufio"
	"bytes"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/mesh"
	"os"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

var Vars = map[string]string{}

func Englang(s string) {
	scanner := bufio.NewScanner(bytes.NewBufferString(s))
	for scanner.Scan() {
		s := scanner.Text()
		r := englangBurst(s)
		if r == "" {
			break
		}
	}
}

func englangBurst(acc string) string {
	var ret string
	ret = englangReadfile(acc)
	if ret != "" {
		return ret
	}
	ret = englangGetEnvironment(acc)
	if ret != "" {
		return ret
	}
	ret = englangVariableSet(acc)
	if ret != "" {
		return ret
	}
	ret = englangVariableGet(acc)
	if ret != "" {
		return ret
	}
	ret = englangInitialize(acc)
	if ret != "" {
		return ret
	}
	ret = englangFinish(acc)
	if ret != "" {
		return ret
	}
	return ""
}

func englangProcess(s string) string {
	content := s
	for {
		processed := englangBurst(content)
		if processed == "" {
			break
		}
		content = processed
	}
	return content
}

func englangInitialize(s string) string {
	var content string
	if nil != englang.Scanf(s, "Initialize container with key %s.", &content) {
		return ""
	}
	content = englangProcess(content)

	containerKey := content
	ret := mesh.Englang(englang.Printf("Call server http://127.0.0.1:7777 path /idle?apikey=%s with method GET and content %s. The call expects englang.", containerKey, "Wait for 10 seconds for a new task."))
	if ret == "success" {
		return "I initialized the burst."
	}
	if ret != "too early" {
		Vars["accumulator"] = ret
	}
	return ret
}

func englangFinish(s string) string {
	var content, apiKey string
	if nil != englang.Scanf(s, "Finish container with content %s and key %s.", &content, &apiKey) {
		return ""
	}
	content = englangProcess(content)

	containerKey := englangProcess(apiKey)
	ret := mesh.Englang(englang.Printf("Call server http://127.0.0.1:7777 path /idle?apikey=%s with method PUT and content %s. The call expects success.", containerKey, content))
	if ret == "success" {
		return "I finished the burst."
	}
	return ""
}

func englangVariableSet(s string) string {
	var content string
	if nil != englang.Scanf1(s, "into %s.", &content) {
		return ""
	}
	content = englangSecurityCheck(content)
	if content == "" {
		return ""
	}

	Vars[content] = Vars["accumulator"]

	return "ok"
}

func englangVariableGet(s string) string {
	var content string
	if nil != englang.Scanf1(s+".", "from %s.", &content) {
		return ""
	}
	content = englangSecurityCheck(content)
	if content == "" {
		return ""
	}

	Vars["accumulator"] = Vars[content]

	return Vars["accumulator"]
}

func englangReadfile(s string) string {
	var content string
	if nil != englang.Scanf1(s+".", "Read file %s.", &content) &&
		nil != englang.Scanf1(s+".", "from file %s.", &content) {
		return ""
	}
	content = englangSecurityCheck(content)
	if content == "" {
		return ""
	}
	content = englangProcess(content)
	return drawing.NoErrorString(os.ReadFile(content))
}

func englangGetEnvironment(s string) string {
	var env string
	if nil != englang.Scanf1(s+".", "Get environment variable %s.", &env) &&
		nil != englang.Scanf1(s+".", "from environment variable %s.", &env) {
		return ""
	}

	if Vars[env] != "" {
		return Vars[env]
	}

	return os.Getenv(env)
}

func englangSecurityCheck(s string) string {
	// TODO Prevent englang injection attacks
	return s
}
