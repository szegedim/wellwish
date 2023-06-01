package burst

import (
	"bytes"
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

// Because of this UI should be low latency using just bags and direct code
// Bursts should scale out. They are okay to be located elsewhere than the data bags.
// The reason is that large computation will require streaming, and
// streaming is driven by pipelined steps without replies and feedbacks.
// Streaming bandwidth is not affected by co-location of data and code.
// Example: 1million 100ms reads followed by 100ms compute will last 200000 seconds
// Example: 1million 100ms reads streamed into 100ms compute will last 100000 seconds,
// even if there is an extra network latency of 100ms

func Setup() {
	stateful.RegisterModuleForBackup(&BurstSession)

	http.HandleFunc("/run", func(writer http.ResponseWriter, request *http.Request) {
		lock.Lock()
		defer lock.Unlock()
		apiKey := request.URL.Query().Get("apikey")
		_, call := BurstSession[apiKey]
		if !call {
			management.QuantumGradeAuthorization()
			writer.WriteHeader(http.StatusPaymentRequired)
			drawing.NoErrorWrite(writer.Write([]byte("Payment required with a PUT to /run.coin")))
			return
		}

		input := drawing.NoErrorString(io.ReadAll(request.Body))

		var instruction string
		var key string
		for instruction == "" {
			for k, v := range ContainerRunning {
				if v == "I am idle." {
					key = k
					instruction = englang.Printf("Run this %s and return in http://127.0.0.1%s/idle?apikey=%s.", input, metadata.Http11Port, key)
					ContainerRunning[k] = instruction
					break
				}
			}
			time.Sleep(100 * time.Millisecond)
		}

		var output string
		for output == "" {
			current := ContainerRunning[key]
			if current != instruction {
				output = current
				delete(ContainerRunning, apiKey)
			}
			time.Sleep(100 * time.Millisecond)
		}

		drawing.NoErrorWrite64(io.Copy(writer, bytes.NewBuffer([]byte(output))))
	})
	http.HandleFunc("/idle", func(writer http.ResponseWriter, request *http.Request) {
		lock.Lock()
		defer lock.Unlock()
		apiKey := request.URL.Query().Get("apikey")
		if request.Method == "GET" {
			for k, v := range ContainerRunning {
				var instruction, port, key string
				if nil == englang.Scanf1(v, "Run this %s and return in http://127.0.0.1%s/idle?apikey=%s.", &instruction, &port, &key) {
					if k == key {
						cmd1 := bytes.NewBufferString(v)
						drawing.NoErrorWrite64(io.Copy(writer, cmd1))
						return
					}
				}
			}
			apiKey = drawing.GenerateUniqueKey()
			query := englang.Printf("Idle.")
			ContainerRunning[apiKey] = "I am idle."
			cmd1 := englang.Printf("Run this %s and return in http://127.0.0.1%s/idle?apikey=%s.", query, metadata.Http11Port, apiKey)
			ret := bytes.NewBufferString(cmd1)
			drawing.NoErrorWrite64(io.Copy(writer, ret))
			return
		}
		if request.Method == "PUT" {
			result := drawing.NoErrorString(io.ReadAll(request.Body))

			var instruction, port, key string
			if nil == englang.Scanf1(result, "Run this %s and return in http://127.0.0.1%s/idle?apikey=%s.", &instruction, &port, &key) {
				if key == apiKey {
					idle := ContainerRunning[apiKey]
					if idle != "I am idle." {
						ContainerRunning[apiKey] = instruction
					}
				}
			}
		}
	})
	http.HandleFunc("/run.coin", func(w http.ResponseWriter, r *http.Request) {
		lock.Lock()
		defer lock.Unlock()
		// Setup burst sessions, a range of time, when a coin can be used for bursts.
		if r.Method == "PUT" {
			coinToUse := billing.ValidatedCoinContent(w, r)
			if coinToUse != "" {
				// TODO generate new?
				burst := coinToUse
				// TODO cleanup
				BurstSession[burst] = englang.Printf(fmt.Sprintf("Burst chain api created from %s is %s/run.coin?apikey=%s. Chain is valid until %s.", coinToUse, metadata.SiteUrl, burst, time.Now().Add(24*time.Hour).String()))
				mesh.RegisterIndex(burst)
				// TODO cleanup
				// mesh.SetIndex(burst, mesh.WhoAmI)
				management.QuantumGradeAuthorization()
				_, _ = w.Write([]byte(burst))
				return
			}
			management.QuantumGradeAuthorization()
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}

		if r.Method == "GET" {
			apiKey := r.URL.Query().Get("apikey")
			session, sessionValid := BurstSession[apiKey]
			if !sessionValid {
				management.QuantumGradeAuthorization()
				_, _ = w.Write([]byte("payment required"))
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}
			management.QuantumGradeAuthorization()
			_, _ = w.Write([]byte(session))
			return
		}
	})
}
