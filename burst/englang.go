package burst

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
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

// This is a module code that runs burst containers.
// Containers typically run as a docker container or other isolated process.
// They can be implemented as a co-located container in the same pod
// They communicate through local 127.0.0.1:2121 UDP calls
// This makes them safer
// Security design considerations.
//
// The locality is ensured by private keys distributed early starting with the activation key.
// This ensures that we have a local runner and user code cannot get the bust call of other user code.
// What does this mean?
// - Idle process is exposed on 127.0.0.1:2121, and it responds to local endpoints only.
// - Idle process returns a task and a key to complete the task.
// - Malicious tasks may go for idle again.
// - We protect against this by letting bursts run for a term e.g. ten seconds
// - We protect against this also by not issuing a new key until the previous one finishes. //TODO
// - Each runner connects to the Idle process using the activation key
// - It returns a unique burst session key for the container.
// - The activation key is deleted from the container once used. //TODO
// - The init task of the container is our burst runner. It should not be set debuggable by workload.
// - Do not use secrets, these can be read by the Idle process stub, but by workload as well.
// - The runner restarts after each run, so that any local state and code is lost disabling double /idle calls.
// - The init task also terminates the container, if the workload tries to kill it.
// - The final column is time fencing allowing /idle calls only once every minute when workloads are already gone. //TODO

func IsEmptyMessage(finish string) bool {
	var message string
	err := englang.Scanf1(finish, "Message has started.%sMessage has finished.", &message)
	if err == nil && message == "" {
		return true
	}
	return false
}

func ProcessBoxMessageEnglang(input string) string {
	lock.Lock()
	defer lock.Unlock()

	var message string
	err := englang.Scanf1(input, "Message has started.%sMessage has finished.", &message)
	if err == nil {
		var containerKey string
		if nil == englang.Scanf1(message, "Generate a burst container with key %s.", &containerKey) {
			if containerKey != metadata.ActivationKey {
				message = "We need the activation key."
			} else {
				containerKey = drawing.GenerateUniqueKey()
				message = englang.Printf("Burst container has key %s and it is running %s.", containerKey, "idle")
				ContainerRunning[containerKey] = message
				go func(key string) {
					time.Sleep(MaxBurstRuntime + 5*time.Second)
					lock.Lock()
					//TODO
					defer lock.Unlock()
					delete(ContainerRunning, key)
				}(containerKey)
			}
		}
		if nil == englang.Scanf1(message, "Get a task for the idle container with key %s.", &containerKey) {
			containerContent, ok := ContainerRunning[containerKey]
			var status, key string
			if ok && nil == englang.Scanf1(containerContent, "Burst container has key %s and it is running %s.", &key, &status) &&
				status == "idle" {
				message = englang.Printf("Burst container has key %s and no new task has arrived.", key)
			} else {
				var code string
				if ok && nil == englang.Scanf1(containerContent+"DOWMXU", "Burst container has key %s and it is running %s.Run this."+"%sDOWMXU", &key, &status, &code) &&
					status == "code" {
					message = containerContent
				}
			}
		}
		var result string
		if nil == englang.Scanf1(message+"EOEJNEEIM", "Return the results for container with key %s.Return this.%s"+"EOEJNEEIM", &containerKey, &result) {
			message = englang.Printf("Burst container has key %s and it is finished with the following result %s", containerKey, result)
			ContainerRunning[containerKey] = message
			go func() {
				time.Sleep(3 * time.Second)
				// TODO
				lock.Lock()
				defer lock.Unlock()
				delete(ContainerRunning, containerKey)
			}()
			message = ""
		}
		return englang.Printf("Message has started.%sMessage has finished.", message)
	}
	return ""
}

func ProcessBurstMessageEnglang(input string) string {
	lock.Lock()
	defer lock.Unlock()

	if input == "" {
		input = englang.Printf("Message has started.%sMessage has finished.", "")
	}
	var message string
	err := englang.Scanf1(input, "Message has started.%sMessage has finished.", &message)
	if err == nil {
		if message == "" {
			message = englang.Printf("Generate a burst container with key %s.", metadata.ActivationKey)
		}
		var key, status string
		if nil == englang.Scanf1(message, "Burst container has key %s and it is running %s.", &key, &status) {
			message = englang.Printf("Get a task for the idle container with key %s.", key)
		}
		if nil == englang.Scanf1(message, "Burst container has key %s and no new task has arrived.", &key, &status) {
			message = englang.Printf("Get a task for the idle container with key %s.", key)
		}
		var task string
		if nil == englang.Scanf1(message+"RHBABDCLF", "Burst container has key %s and it is running %s.Run this.%sRHBABDCLF", &key, &status, &task) {
			// HOTSPOT
			result := RunExternalShell(task)
			message = englang.Printf("Return the results for container with key %s.Return this.%s", key, result)
			if ShutdownOnFinish {
				// The container must shut down at this point to be restarted by Docker/Kubernetes/etc. clean.
				go func() {
					time.Sleep(500 * time.Millisecond)
					os.Exit(0)
				}()
			}
		}
		if message == "We need the activation key." {
			fmt.Println(fmt.Errorf(message))
			return ""
		}
		message = englang.Printf("Message has started.%sMessage has finished.", message)
	}
	return message
}
