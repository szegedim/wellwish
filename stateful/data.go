package stateful

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

var ContainerIndexLimit = 100000

// The list of modules that are backed up and restored on startup
var stateModules = make([]*map[string]string, 0)

// We take a checkpoint every period that contains all data in the node
var checkpointPeriod = 10 * time.Second
var checkpoint *[]byte = nil

// We clean up least recently used items that are backed up,
// if there is memory pressure.
var startupTime = time.Now()
var lru = map[string]string{}

var lock = sync.Mutex{}
