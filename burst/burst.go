package burst

import (
	"fmt"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"gitlab.com/eper.io/engine/stateful"
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

func SetupBurst() {
	stateful.RegisterModuleForBackup(&BurstSession)

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			ok, _, _, voucher := billing.ValidateVoucher(w, r, true)
			if ok {
				burst := drawing.GenerateUniqueKey()
				// TODO cleanup
				BurstSession[burst] = englang.Printf(fmt.Sprintf("Burst chain api created from %s is %s/api?apikey=%s. Chain is valid until %s.", voucher, metadata.SiteUrl, burst, time.Now().Add(24*time.Hour).String()))
				mesh.SetIndex(burst, mesh.WhoAmI)
				management.QuantumGradeAuthorization()
				_, _ = w.Write([]byte(burst))
				return
			} else {
				management.QuantumGradeAuthorization()
				_, _ = w.Write([]byte("payment required"))
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}
		}

		if r.Method == "HEAD" {
			apiKey := r.URL.Query().Get("apikey")
			session, sessionValid := BurstSession[apiKey]
			burst, burstOk := Burst[apiKey]
			if !sessionValid && !burstOk {
				management.QuantumGradeAuthorization()
				_, _ = w.Write([]byte("payment required"))
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}
			if sessionValid {
				management.QuantumGradeAuthorization()
				_, _ = w.Write([]byte(session))
				return
			}
			if burstOk {
				management.QuantumGradeAuthorization()
				_, _ = w.Write([]byte(burst))
				return
			}
		}

		if r.Method == "GET" {
			apiKey := r.URL.Query().Get("apikey")
			_, call := BurstSession[apiKey]
			_, result := Burst[apiKey]
			if !call && !result {
				management.QuantumGradeAuthorization()
				_, _ = w.Write([]byte("payment required"))
				w.WriteHeader(http.StatusPaymentRequired)
			}

			if call {
				body := string(drawing.NoErrorBytes(io.ReadAll(r.Body)))
				burst := drawing.GenerateUniqueKey()
				Burst[burst] = englang.Printf("Request paid with %s %s", apiKey, body)

				// This is just a perf improvement that can be eliminated
				go func() { NewTask <- burst }()

				management.QuantumGradeAuthorization()
				_, _ = w.Write([]byte(burst))
				return
			}

			if result {
				burst, call := Burst[apiKey]
				if call {
					if strings.HasPrefix(burst, "Response ") {
						management.QuantumGradeAuthorization()
						_, _ = io.Copy(w, strings.NewReader(burst[len("Response "):]))
						return
					} else {
						w.WriteHeader(http.StatusTooEarly)
						management.QuantumGradeAuthorization()
						_, _ = w.Write([]byte("too early"))
						return
					}
				}
			}

			management.QuantumGradeAuthorization()
			w.WriteHeader(http.StatusNotFound)
			return
		}
		management.QuantumGradeAuthorization()
	})

	http.HandleFunc("/idle", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.RemoteAddr, "127.0.0.1") {
			// local burst runner allowed only
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		apiKey := r.URL.Query().Get("apikey")
		if apiKey == metadata.ActivationKey {
			time.Sleep(0 * time.Second)
			currentKey := drawing.GenerateUniqueKey()
			localRunnerEndpoint := fmt.Sprintf("http://127.0.0.1%s", metadata.Http11Port)
			content := fmt.Sprintf(ContainerPattern, "any", localRunnerEndpoint, "idle")
			Container[currentKey] = content
			_, _ = w.Write([]byte(currentKey))
			return
		}
		containerKey := apiKey

		if r.Method == "GET" {
			_, ok := Container[containerKey]
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
						UpdateContainerWithBurst(containerKey, k)
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
			burst := UpdateContainerWithBurst(containerKey, "finished")
			if burst == "idle" || len(burst) != len(drawing.GenerateUniqueKey()) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if strings.HasPrefix(Burst[burst], "running") {
				response := string(drawing.NoErrorBytes(io.ReadAll(r.Body)))
				Burst[burst] = "Response " + response
			}
		}
	})
}
