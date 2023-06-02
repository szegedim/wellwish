package mesh

import (
	"gitlab.com/eper.io/engine/englang"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func SetupExpiry() {
	go func() {
		for {
			verifyExpiry()

			time.Sleep(updateFrequency)
		}
	}()
}

func verifyExpiry() {
	indexLock.Lock()
	defer indexLock.Unlock()
	for k, v := range expiry {
		if v == "" {
			continue
		}

		del := false
		expiry := ""
		err := englang.Scanf(v, "Validated until %s.", &expiry)
		if err != nil {
			del = true
		}
		expired, err := time.Parse("Jan 2, 2006", expiry)
		if err != nil {
			del = true
		}
		if del || time.Now().After(expired) {
			DeleteIndex(k)
		}
	}
}

func SetExpiry(key string, period time.Duration) {
	indexLock.Lock()
	defer indexLock.Unlock()
	n := time.Now()
	willExpire := englang.Printf("Validated until %s.", n.Add(period).Format("Jan 2, 2006"))
	expiry[key] = willExpire
}

func CheckExpiry(key string) bool {
	indexLock.Lock()
	defer indexLock.Unlock()
	_, ok := index[key]
	return ok
}
