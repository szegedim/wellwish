package stateful

import (
	rand "crypto/rand"
	"fmt"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/metadata"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// The typical scenario is that transactions require 10k items in the cache.
// A new transaction uses these and adds 500 more.
// 250 will be needed later.

// This is the most basic least recently used cache logic.
// The advantage is that it works on single write memory.
// It uses a bit more CPU in exchange.
// We will typically use 1 Gbps full duplex containers.
// This allows offloading old data right away.
// However, the ideal is that the whole container just restarts at this point.
// It will load data back from backups.

// cleanupMemoryCache makes some space.
func cleanupMemoryCache(data *map[string]string, lru *map[string]string) {
	start := time.Now()
	var n = 0
	cacheSize := len(*data)
	for cacheSize > ContainerIndexLimit {
		overheadPercentage := (cacheSize - ContainerIndexLimit) * 100 / ContainerIndexLimit
		for i := 0; i < overheadPercentage; i++ {
			selected := englang.DecimalString(int64(random(metadata.RandomSalt) % 100))
			for k, v := range *lru {
				if v == selected {
					delete(*data, k)
					delete(*lru, k)
				}
				n++
				if n == 100 {
					n = 0
					if time.Now().Sub(start).Milliseconds() > 1000 {
						return
					}
				}
			}
		}
		n++
		if n == 100 {
			n = 0
			if time.Now().Sub(start).Milliseconds() > 1000 {
				return
			}
		}
		cacheSize = len(*data)
	}
	if cacheSize > ContainerIndexLimit {
		fmt.Println("insufficient cleanup")
	}
}

func touchMemoryCache(lru *map[string]string, apiKey string) {
	percentage := englang.DecimalString(int64(random(metadata.RandomSalt) % 100))
	(*lru)[apiKey] = percentage
}

func random(salt string) uint32 {
	buf := make([]byte, 4)
	n, err := rand.Read(buf)
	if err != nil || n != 4 {
		return 0
	}
	x := []byte(salt[5 : 5+4])
	y := uint32(buf[0]^x[0])<<24 | uint32(buf[1]^x[1])<<16 | uint32(buf[2]^x[2])<<8 | uint32(buf[3]^x[3])<<0
	return y
}
