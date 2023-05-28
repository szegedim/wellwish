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

	SetupBurstLambdaEndpoint("/run", true)
	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		// Setup burst sessions, a range of time, when a coin can be used for bursts.
		if r.Method == "PUT" {
			payment := drawing.NoErrorString(io.ReadAll(r.Body))
			coinToUse, err := billing.RedeemCoin(payment)
			if err != nil {
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}

			burst := drawing.GenerateUniqueKey()
			// TODO cleanup
			BurstSession[burst] = englang.Printf(fmt.Sprintf("Burst chain api created from %s is %s/api?apikey=%s. Chain is valid until %s.", coinToUse, metadata.SiteUrl, burst, time.Now().Add(24*time.Hour).String()))
			mesh.SetIndex(burst, mesh.WhoAmI)
			management.QuantumGradeAuthorization()
			_, _ = w.Write([]byte(burst))
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
