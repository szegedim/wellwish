package stateful

import (
	"fmt"
	"gitlab.com/eper.io/engine/englang"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// This is the most basic least recently used cache logic.
// The advantage is that it works on single write memory.
// It uses a bit more CPU in exchange.
// We will typically use 1 Gbps full duplex containers.
// This allows offloading old data right away.
// However, the ideal is that the whole container just restarts at this point.
// It will load data back from backups.

func cleanupMemoryCache(data *map[string]string, lru *map[string]string) {
	for i := 0; i < 2; i++ {
		cacheSize := len(*lru)
		if cacheSize >= containerIndexLimit {
			// Make some space
			lrutime := ""
			lrukey := ""
			size := 0
			num := 0
			for k, v := range *lru {
				size = size + len(v) + len(k)
				num++
				if lrutime == "" || v < lrutime {
					lrukey = k
					lrutime = v
				}
				if num == 104 {
					break
				}
			}
			if lrutime != "" {
				// Make sure we are backed up
				if (*data)[lrukey] == string(readStatefulItem(lrukey, false)) {
					fmt.Printf("Memory usage before cleanup: %d %s\n", size, lrukey)
					delete(*data, lrukey)
					delete(*lru, lrukey)
				} else {
					// This will rarely happen, when the container is almost full
					// and the backup server is slow or offline
					fmt.Printf("out of memory")
					time.Sleep(checkpointPeriod)
				}
			}
		}
	}
}

func touchMemoryCache(lru *map[string]string, apiKey string) {
	timeStamp := englang.DecimalString(int64(time.Now().Sub(startupTime).Seconds() + 0.01))
	(*lru)[apiKey] = timeStamp
}
