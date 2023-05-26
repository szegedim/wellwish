package box

import (
	"bufio"
	"bytes"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"os"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

var Context = map[string]string{}

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
	if nil == englang.Scanf1(acc, "Set burst timeout to ten seconds.") {
		go func() {
			time.Sleep(10 * time.Second)
			os.Exit(2)
		}()
		return "success"
	}
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
	ret = englangGenerateKey(acc)
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

func englangGenerateKey(s string) string {
	var begin, end string
	if nil != englang.Scanf(s, "a newly generated burst key", &begin, &end) {
		return ""
	}
	for i := 0; i < 120; i++ {
		ret := mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /idle?apikey=%s with method GET and content %s. The call expects englang.", metadata.Http11Port, metadata.ActivationKey, "Wait for 10 seconds for a new task."))
		if ret != "" {
			return begin + ret + end
		}
		time.Sleep(1 * time.Second)
	}
	return ""
}

func englangInitialize(s string) string {
	var content, task, key string
	if nil != englang.Scanf(s, "Fetch task %s into %s using key in %s.", &content, &task, &key) {
		return ""
	}
	containerKey := englangProcess(content)
	Context[key] = containerKey

	ret := mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /idle?apikey=%s with method GET and content %s. The call expects englang.", metadata.Http11Port, containerKey, "Wait for 10 seconds for a new task."))
	if ret != "too early" {
		Context[task] = ret
	}
	return ret
}

func englangFinish(s string) string {
	var content, apiKey string
	if nil != englang.Scanf(s, "Upload container result content %s and key %s.", &content, &apiKey) {
		return ""
	}
	content = englangProcess(content)

	containerKey := englangProcess(apiKey)
	ret := mesh.EnglangRequest(englang.Printf("Call server http://127.0.0.1%s path /idle?apikey=%s with method PUT and content %s. The call expects success.", metadata.Http11Port, containerKey, content))
	if ret == "success" {
		Context = map[string]string{}
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

	Context[content] = Context["accumulator"]

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

	Context["accumulator"] = Context[content]

	return Context["accumulator"]
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

	if Context[env] != "" {
		return Context[env]
	}

	return os.Getenv(env)
}

func englangSecurityCheck(s string) string {
	// TODO Prevent englang injection attacks
	return s
}
