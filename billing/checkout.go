package billing

import (
	"bytes"
	"fmt"
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

var sample = fmt.Sprintf(metadata.OrderPattern, "\vExample Buyer Inc.\v", "\v111 S Ave\v, \vSan Fransisco\v, \vCA\v, \v55555\v, \vUnited States\v", "\vinfo\v@\vexample.com\v", "\v10\v", metadata.UnitPrice, "USD 10", "0")

func setupCheckout() {
	http.HandleFunc("/checkout.html", func(w http.ResponseWriter, r *http.Request) {
		if drawing.EnsureAPIKey(w, r) != nil {
			return
		}
		if drawing.ResetSession(w, r) != nil {
			return
		}
		drawing.ServeRemoteForm(w, r, "checkout")
	})
	http.HandleFunc("/checkout.png", func(w http.ResponseWriter, r *http.Request) {
		if drawing.EnsureAPIKey(w, r) != nil {
			return
		}
		drawing.ServeRemoteFrame(w, r, declareCheckoutForm)
	})
	http.HandleFunc("/checkout", func(w http.ResponseWriter, r *http.Request) {
		// See tests.Me for a good example to put as a body.
		order := drawing.NoErrorString(io.ReadAll(r.Body))
		var company string
		var address string
		var email string
		var amount string = "10"
		var unit string = metadata.UnitPrice
		var total string = "USD 10"
		var tax string = "0"
		err := englang.Scanf(order, metadata.OrderPattern, &company, &address, &email, &amount, &unit, &total, &tax)
		if err == nil {
			orderId := drawing.GenerateUniqueKey()
			IssueOrder(orderId, amount, company, address, email, unit)
			ret := bytes.NewBuffer([]byte(orderId))
			_, _ = io.Copy(w, ret)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			ret := bytes.NewBufferString("Invalid request. Please use this pattern.\n" + strings.ReplaceAll(sample, "\v", ""))
			_, _ = io.Copy(w, ret)
		}
	})
}

func declareCheckoutForm(session *drawing.Session) {
	if session.Form.Boxes == nil {
		drawing.DeclareForm(session, "./billing/res/checkout.png")

		const Logo = 0
		const OrderText = 1
		const BackButton = 2
		const OrderButton = 3

		pattern := metadata.OrderPattern
		drawing.SetImage(session, Logo, "./metadata/logo.png", drawing.Content{Text: "", Lines: 1, Editable: false, FontColor: drawing.White, BackgroundColor: drawing.Black, Alignment: 1})
		drawing.PutText(session, OrderText, drawing.Content{Text: "�" + sample, Lines: 20, Editable: true, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 1})
		drawing.PutText(session, BackButton, drawing.Content{Text: "    Cancel    ", Lines: 1, Selectable: false, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
		drawing.PutText(session, OrderButton, drawing.Content{Text: "    Submit    ", Lines: 1, Selectable: false, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})

		session.SignalClicked = func(session *drawing.Session, i int) {
			if i == BackButton {
				session.Redirect = "/"
				session.SelectedBox = -1
			}
			if i == OrderButton {
				s := session.Text[OrderText].Text
				s = strings.ReplaceAll(s, "�", "")
				var company string
				var address string
				var email string
				var amount string = "10"
				var unit string = metadata.UnitPrice
				var total string = "USD 10"
				var tax string = "0"
				err := englang.Scanf(s, pattern, &company, &address, &email, &amount, &unit, &total, &tax)
				if err == nil {
					IssueOrder(session.ApiKey, amount, company, address, email, unit)
					session.Redirect = fmt.Sprintf("/invoice.html?apikey=%s", session.ApiKey)
					session.SignalClosed(session)
				}
			}
		}
		session.SignalTextChange = func(session *drawing.Session, i int, from string, to string) {
			session.Data = from
			session.SignalRecalculate(session)
			if strings.HasPrefix(session.Data, drawing.RevertAndReturn) {
				last := session.Text[OrderText]
				last.Text = session.Data[1:]
				session.Text[OrderText] = last
			}
			session.SignalPartialRedrawNeeded(session, i)
		}
		session.SignalRecalculate = func(session *drawing.Session) {
			var company string
			var address string
			var email string
			var amount string = "10"

			var unit string = "USD 1"
			var total string = "USD 10"
			var tax string = "0"
			s := session.Text[OrderText].Text
			err := englang.Scanf(s, pattern, &company, &address, &email, &amount, &unit, &total, &tax)
			if err != nil {
				s = strings.ReplaceAll(s, "�", "")
				err = englang.Scanf(s, pattern, &company, &address, &email, &amount, &unit, &total, &tax)
			}
			change := false
			if amount == "\v�" {
				amount = "\v0�"
				change = true
			}
			if !englang.IsEmail(email) {
				err = fmt.Errorf("not an email")
			}
			if !englang.IsAddress(&address) {
				err = fmt.Errorf("not an address")
			}
			if !englang.IsCompany(company) {
				err = fmt.Errorf("not a company")
			}
			if !englang.IsNumber(amount) {
				err = fmt.Errorf("not an amount")
			}
			if strings.ReplaceAll(unit, "�", "") != metadata.UnitPrice {
				err = fmt.Errorf("cannot change unit price")
			}
			if strings.ReplaceAll(tax, "�", "") != "0" {
				err = fmt.Errorf("cannot change stales tax")
			}
			if err != nil {
				session.Data = drawing.RevertAndReturn + session.Data
				return
			}
			newTotal := englang.Evaluate(fmt.Sprintf("%s multiplied by %s", amount, unit))
			if newTotal != strings.ReplaceAll(total, "�", "") {
				change = true
			}
			if change {
				s = fmt.Sprintf(pattern, company, address, email, amount, unit, newTotal, tax)
				err = englang.Scanf(s, pattern, &company, &address, &email, &amount, &unit, &total, &tax)
				if err == nil {
					if !strings.Contains(s, "�") && !strings.Contains(session.Text[OrderText].Text, "�") {
						s = s + "�"
					}
					saved := session.Text[OrderText]
					saved.Text = s
					session.Text[OrderText] = saved
				}
			}
		}
		session.SignalRecalculate(session)
	}
}

func IssueOrder(apiKey string, amount string, company string, address string, email string, unit string) {
	invoice := IssueVouchers(apiKey, amount, company, address, email, unit)
	orders[apiKey] = invoice
}

func IssueVouchers(apiKey string, amount string, company string, address string, email string, unit string) string {
	if apiKey == "" {
		return ""
	}

	issued := time.Now()
	amount = strings.TrimSpace(amount)
	NewVoucher(apiKey, amount, issued)
	total := englang.Evaluate(fmt.Sprintf("%s multiplied by %s", amount, unit))

	invoice := englang.Printf(metadata.InvoicePattern,
		metadata.CompanyInfo, issued.Format("Jan 2, 2006"), drawing.RedactPublicKey(apiKey),
		company, address, email, amount, unit, total,
		"Status is due.")

	return invoice
}
