package mining

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/stateful"
	"net/http"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func Setup() {
	stateful.RegisterModuleForBackup(&miningTicket)

	http.HandleFunc("/cryptonugget.coin", func(w http.ResponseWriter, r *http.Request) {
		// Setup burst sessions, a range of time, when a coin can be used for bursts.
		if r.Method == "PUT" {
			coinToUse := billing.ValidatedCoinContent(w, r)
			if coinToUse != "" {
				mineTicket := makeCryptoNuggetMine(coinToUse)
				management.QuantumGradeAuthorization()
				_, _ = w.Write([]byte(mineTicket))
				return
			}
			management.QuantumGradeAuthorization()
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}

		if r.Method == "GET" {
			apiKey := r.URL.Query().Get("apikey")
			session, sessionValid := miningTicket[apiKey]
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

	http.HandleFunc("/cryptonugget", func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.URL.Query().Get("apikey")
		if r.Method == "GET" {
			cryptoNuggetMine := apiKey
			if !mesh.CheckExpiry(cryptoNuggetMine) {
				delete(miningTicket, cryptoNuggetMine)
				management.QuantumGradeAuthorization()
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			traces := miningTicket[cryptoNuggetMine]
			if traces == "" {
				management.QuantumGradeAuthorization()
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			writer := bufio.NewWriter(w)
			for j := 0; j < 4096/8/4; j++ {
				r := uint32(0)
				for i := 0; i < 3 || r == 0; i++ {
					r = Random(apiKey)
				}
				_, _ = writer.WriteString(fmt.Sprintf("%08x", r))
			}
			_ = writer.Flush()
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

func Random(salt string) uint32 {
	buf := make([]byte, 4)
	n, err := rand.Read(buf)
	if err != nil || n != 4 {
		return 0
	}
	x := []byte(salt[5 : 5+4])
	y := uint32(buf[0]^x[0])<<24 | uint32(buf[1]^x[1])<<16 | uint32(buf[2]^x[2])<<8 | uint32(buf[3]^x[3])<<0
	return y
}

func makeCryptoNuggetMine(voucher string) string {
	mesh.RegisterIndex(voucher)
	mesh.SetExpiry(voucher, ValidPeriod)
	miningTicket[voucher] = fmt.Sprintf("Mine expires on %s.", time.Now().Add(ValidPeriod).Format("Jan 2, 2006"))
	return voucher
}
