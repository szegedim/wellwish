package crypto

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"gitlab.com/eper.io/engine/englang"
	"net/http"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func SetupCryptoMining() {
	http.HandleFunc("/cryptonugget", func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := IsApiKeyValid(w, r)
		if err != nil {
			return
		}
		writer := bufio.NewWriter(w)
		_, _ = writer.WriteString(fmt.Sprintf("%8x", random(apiKey)))
		_ = writer.Flush()
	})
}

func random(string2 string) uint32 {
	buf := make([]byte, 4)
	n, err := rand.Read(buf)
	if err != nil || n != 4 {
		return 0
	}
	x := []byte(string2[5 : 5+4])
	y := uint32(buf[0]^x[0])<<24 | uint32(buf[1]^x[1])<<16 | uint32(buf[2]^x[2])<<8 | uint32(buf[3]^x[3])<<0
	return y
}

func IsApiKeyValid(w http.ResponseWriter, r *http.Request) (string, error) {
	apiKey := r.URL.Query().Get("apikey")
	if Tickets[apiKey] == "" {
		w.WriteHeader(http.StatusPaymentRequired)
		return "", fmt.Errorf("no payment")
	}
	expiry := ""
	err := englang.Scanf(Tickets[apiKey], TicketExpiry, &expiry)
	if err != nil {
		w.WriteHeader(http.StatusPaymentRequired)
		return "", fmt.Errorf("expired apikey")
	}
	expired, err := time.Parse("Jan 2, 2006", expiry)
	if err != nil {
		w.WriteHeader(http.StatusPaymentRequired)
		return "", fmt.Errorf("expiry misformatted apikey")
	}
	if time.Now().After(expired) {
		w.WriteHeader(http.StatusPaymentRequired)
		return "", fmt.Errorf("expired apikey")
	}
	return apiKey, nil
}

func MakeTicket(voucher string) {
	Tickets[voucher] = fmt.Sprintf(TicketExpiry, time.Now().Add(168*time.Hour).Format("Jan 2, 2006"))
}
