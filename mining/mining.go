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
				// Want to extend? Let it delete and create a new one.
				// Reason? Newly generated ids are safer.
				// TODO cleanup
				mineTicket := MakeCryptoNuggetMine(coinToUse)
				// TODO cleanup
				// mesh.SetIndex(burst, mesh.WhoAmI)
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
		if r.Method == "GET" {
			apiKey, err := billing.IsApiKeyValid(w, r, &miningTicket, mesh.Proxy)
			if err != nil {
				return
			}
			writer := bufio.NewWriter(w)
			r := uint32(0)
			for i := 0; i < 3 || r == 0; i++ {
				r = random(apiKey)
			}
			_, _ = writer.WriteString(fmt.Sprintf("%8x", r))
			_ = writer.Flush()
			return
		}
		//if r.Method == "PUT" {
		//	ok, _, _, voucher := billing.ValidateVoucher(w, r, true)
		//	if ok {
		//
		//		http.Redirect(w, r, fmt.Sprintf("/cryptonugget?apikey=%s", voucher), http.StatusTemporaryRedirect)
		//	} else {
		//		w.WriteHeader(http.StatusPaymentRequired)
		//	}
		//	return
		//}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
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

func MakeCryptoNuggetMine(voucher string) string {
	miningTicket[voucher] = fmt.Sprintf(billing.TicketExpiry, time.Now().Add(168*time.Hour).Format("Jan 2, 2006"))
	return voucher
}
