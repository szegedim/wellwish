package mesh

import (
	"fmt"
	"gitlab.com/eper.io/engine/englang"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Index contains key indexes that are propagated to all servers.
// This allows us to use even random load balancer without any sticky setting by IP, cookie or apikey
// Note: the reason we use indexes is not to use cookies that require annoying prompts
// These are temporary stateless indexes

// Stateful indexes are keys and values that are backed by a stateful disk server backup
// They can be cleaned in our memory but fetch again from backups
// Stateful indexes can also be used to support load balancing

// Finally cleaned up indexes can have a rule to clean up periodically
// Stateful indexes are cleaned up by design

func IndexLengthForTestingOnly() string {
	i := 0
	for k, v := range index {
		if k != "" && v != "" {
			i++
		}
		fmt.Println(k, v)
	}
	return englang.DecimalString(int64(i))
}

func GetIndex(k string) string {
	indexLock.Lock()
	defer indexLock.Unlock()
	return index[k]
	//return stateful.GetStatefulItem(&index, k)
}

func SetIndex(k string, v string) {
	indexLock.Lock()
	defer indexLock.Unlock()
	index[k] = v
	//stateful.SetStatefulItem(&index, k, v)
}

func RegisterIndex(index string) {
	SetIndex(index, WhoAmI)
}
