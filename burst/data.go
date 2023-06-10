package burst

import (
	"bufio"
	"bytes"
	"gitlab.com/eper.io/engine/englang"
	"sync"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Burst is a reference implementation of serverless functions
// We believe the following trends in personal clouds
// - Bursts are heavy on CPU compute and do not wait much on events saving resources even more than serverless
// - Bursts can easily be triggered by other bursts making waiting unnecessary
// - Serverless bursts will take over compute heavy workloads from functions
// - Any waiting logic will be split by the mesh network
// - Mesh networks will go beyond TCP/IP, and they become session heavy instead of endpoint heavy
// - Communication will shift from REST/XML to Englang processed by ChatGPT like agents
// - Bursts will be completely transparent allowing bare metal error resolution
// - Bursts are short enough that they make dynamic memory and garbage collection unnecessary
// - Bursts are sized and sliced, so that their bandwidth can be efficiently leveraged in their runtime cap.

var lock = sync.Mutex{}

var BurstSession = map[string]string{}
var ContainerRunning = map[string]string{}
var ContainerResults = map[string]chan string{}

const ValidPeriod = 168 * time.Hour

// Use DummyBroker, if this is 0
var BurstRunners = 0
var MaxBurstRuntime = 3 * time.Second

func LogSnapshot(m string, w bufio.Writer, r *bufio.Reader) {
	if m == "GET" {
		for k, v := range BurstSession {
			englang.WriteIndexedEntry(w, k, "burst", bytes.NewBufferString(v))
		}
	}
	if m == "PUT" {
		for {
			e, k, v := englang.ReadIndexedEntry(*r)
			if k == "" {
				return
			}
			if e == "burst" {
				BurstSession[k] = v
			}
		}
	}
}
