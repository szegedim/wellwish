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

var containerIndexLimit = 100000

var checkpointPeriod = 10 * time.Second

var lock = sync.Mutex{}
var lru = map[string]string{}
var stateModules = make([]*map[string]string, 0)

var checkpoint *[]byte = nil

var startupTime = time.Now()