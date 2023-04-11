package burst

import (
	"fmt"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
	"strings"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// The main design behind burst runners is that they to be scalable.
// Data locality means that data is co-located with burst containers.
// Data locality is important in some cases, especially UI driven code like ours.
// However, bursts are designed to handle the longer running process.

// Because of this UI should be low latency using just sacks and direct code
// Bursts should scale out. They are okay to be located elsewhere than the data sacks.
// The reason is that large computation will require streaming, and
// streaming is driven by pipelined steps without replies and feedbacks.
// Streaming bandwidth is not affected by co-location of data and code.
// Example: 1million 100ms reads followed by 100ms compute will last 200000 seconds
// Example: 1million 100ms reads streamed into 100ms compute will last 100000 seconds,
// even if there is an extra network latency of 100ms

func Setup() {
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			ok, _, _, voucher := billing.ValidateVoucher(w, r, true)
			if ok {
				burst := drawing.GenerateUniqueKey()
				BurstSession[burst] = englang.Printf(fmt.Sprintf("Burst chain api created from %s is %s/api?apikey=%s. Chain has %s left.", voucher, metadata.SiteUrl, burst, (24 * time.Hour).String()))
				_, _ = w.Write([]byte(burst))
				return
			} else {
				_, _ = w.Write([]byte("payment required"))
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}
		}

		if r.Method == "GET" {
			apiKey := r.URL.Query().Get("apikey")
			_, call := BurstSession[apiKey]
			_, result := Burst[apiKey]
			if !call && !result {
				_, _ = w.Write([]byte("payment required"))
				w.WriteHeader(http.StatusPaymentRequired)
			}

			if call {
				body := string(drawing.NoErrorBytes(io.ReadAll(r.Body)))
				burst := drawing.GenerateUniqueKey()
				Burst[burst] = englang.Printf("Request paid with %s %s", apiKey, body)

				// This is just a perf improvement that can be eliminated
				go func() { NewTask <- burst }()

				_, _ = w.Write([]byte(burst))
				return
			}

			if result {
				burst, call := Burst[apiKey]
				if call {
					if strings.HasPrefix(burst, "Response ") {
						_, _ = io.Copy(w, strings.NewReader(burst[len("Response "):]))
						return
					} else {
						w.WriteHeader(http.StatusTooEarly)
						_, _ = w.Write([]byte("too early"))
						return
					}
				}
			}

			w.WriteHeader(http.StatusNotFound)
			return
		}
	})

	http.HandleFunc("/idle", func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.URL.Query().Get("apikey")
		if len(apiKey) != len(drawing.GenerateUniqueKey()) {
			_, _ = w.Write([]byte("payment required"))
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}

		if r.Method == "GET" {
			// TODO get paid
			_, ok := Container[apiKey]
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			in := drawing.NoErrorString(io.ReadAll(r.Body))
			var seconds string
			_ = englang.Scanf(in, "Wait for %s for a new task.", &seconds)
			retryFor, _ := time.ParseDuration(seconds)
			start := time.Now()

			for {
				started := false
				for k, v := range Burst {
					prefix := englang.Printf("Request paid with %s ", drawing.GenerateUniqueKey())
					if strings.HasPrefix(v, "Request paid with ") {
						_, _ = w.Write([]byte(v[len(prefix):]))
						Burst[k] = "running"
						UpdateContainerWithBurst(apiKey, k)
						started = true
						return
					}
				}
				if !started && time.Now().After(start.Add(retryFor)) {
					w.WriteHeader(http.StatusTooEarly)
					_, _ = w.Write([]byte("too early"))
					return
				}
				// This is just a perf improvement that can be eliminated
				select {
				case <-time.After(time.Now().Sub(start.Add(retryFor))):
					continue
				case <-NewTask:
					continue
				}
			}
		}
		if r.Method == "PUT" {
			burst := UpdateContainerWithBurst(apiKey, "finished")
			if burst == "idle" || len(burst) != len(drawing.GenerateUniqueKey()) {
				_, _ = w.Write([]byte("unauthorized"))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if strings.HasPrefix(Burst[burst], "running") {
				Burst[burst] = "Response " + string(drawing.NoErrorBytes(io.ReadAll(r.Body)))
			}
		}
	})
}
