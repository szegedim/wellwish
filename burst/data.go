package burst

import (
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
// - Bursts are heavy on CPU compute and do not wait much on events
// - Bursts can easily be triggered by other bursts making waiting unnecessary
// - Serverless bursts will take over compute heavy workloads from functions
// - Any waiting logic will be split by the mesh network
// - Mesh networks will go beyond TCP, and they become session heavy
// - Communication will shift from REST/XML to Englang processed by ChatGPT like agents
// - Bursts will be completely transparent allowing bare metal error resolution

var lock = sync.Mutex{}

var BurstSession = map[string]string{}
var ContainerRunning = map[string]string{}

var MaxBurstRuntime = 3 * time.Second
