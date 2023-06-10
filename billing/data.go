package billing

import (
	"bufio"
	"bytes"
	"gitlab.com/eper.io/engine/englang"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Order requests that can be used, cancelled, refunded.
var orders = map[string]string{}

// Vouchers issued from orders.
// Vouchers are just valid for this site as a proof of order and/or payment.
// They cannot be resold, they have no value.
// They are also not a coin or digital currency, but they can be reworked as such with minimal effors.
var vouchers = map[string]string{}

func LogSnapshot(m string, w bufio.Writer, r *bufio.Reader) {
	if m == "GET" {
		for k, v := range orders {
			englang.WriteIndexedEntry(w, k, "order", bytes.NewBufferString(v))
		}
		for k, v := range vouchers {
			englang.WriteIndexedEntry(w, k, "voucher", bytes.NewBufferString(v))
		}
	}
	if m == "PUT" {
		for {
			e, k, v := englang.ReadIndexedEntry(*r)
			if k == "" {
				return
			}
			if e == "order" {
				orders[k] = v
			}
			if e == "voucher" {
				vouchers[k] = v
			}
		}
	}
}
