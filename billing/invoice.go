package billing

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/metadata"
	"net/http"
	"strings"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func SetupInvoice() {
	http.HandleFunc("/invoice.html", func(w http.ResponseWriter, r *http.Request) {
		if drawing.EnsureAPIKey(w, r) != nil {
			return
		}
		if drawing.ResetSession(w, r) != nil {
			return
		}
		drawing.ServeRemoteForm(w, r, "invoice")
	})
	http.HandleFunc("/invoice.png", func(w http.ResponseWriter, r *http.Request) {
		if drawing.EnsureAPIKey(w, r) != nil {
			return
		}
		drawing.ServeRemoteFrame(w, r, declareinvoiceForm)
	})
}

func declareinvoiceForm(session *drawing.Session) {
	if session.Form.Boxes == nil {
		drawing.DeclareForm(session, "./billing/res/invoice.png")

		const Logo = 0
		const InvoiceText = 1
		const CancelButton = 2
		const PaymentButton = 3
		const VoucherButton = 4

		drawing.SetImage(session, Logo, "./metadata/logo.png", drawing.Content{Text: "", Lines: 1, Editable: false, FontColor: drawing.White, BackgroundColor: drawing.Black, Alignment: 1})
		drawing.PutText(session, InvoiceText, drawing.Content{Text: "", Lines: 25, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 1})
		drawing.PutText(session, CancelButton, drawing.Content{Text: "    Refund     ", Lines: 1, Selectable: false, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
		drawing.PutText(session, PaymentButton, drawing.Content{Text: "      Pay      ", Lines: 1, Selectable: false, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
		drawing.PutText(session, VoucherButton, drawing.Content{Text: "    Vouchers   ", Lines: 1, Selectable: false, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})

		session.SignalClicked = func(session *drawing.Session, i int) {
			invoice, done := GetInvoice(session)
			if !done {
				return
			}
			if i == CancelButton {
				last := invoice
				message := "Status is cancelled.\nRefund of any payment will be sent within five business days."
				if strings.Contains(last, "Status is due.") {
					invoice = strings.Replace(last, "Status is due.", message, 1)
				}
				if strings.Contains(last, "Status is paid.") {
					invoice = strings.Replace(last, "Status is paid.", message, 1)
				}
				if invoice != last {
					orders[session.ApiKey] = invoice
					invoiceText := fmt.Sprintf(VoucherInvoicePointer, metadata.SiteUrl, session.ApiKey)
					for key, voucher := range vouchers {
						// ApiKey may point to a voucher directly
						// We also accept an invoice
						if session.ApiKey == key || strings.Contains(voucher, invoiceText) {
							last := voucher
							voucher = CancelVoucher(last, voucher)
							if voucher != last {
								vouchers[key] = voucher
							}
						}
					}

					session.SignalRecalculate(session)
					session.SignalPartialRedrawNeeded(session, InvoiceText)
				}
			}
			if i == PaymentButton {
				if strings.Contains(invoice, "Status is due.") {
					//Paypal/Yatta/Paychex/etc.
					session.Redirect = fmt.Sprintf(metadata.PaymentPattern, drawing.RedactPublicKey(session.ApiKey))
				}
			}
			if i == VoucherButton {
				//Paypal/Yatta/Paychex/etc.
				session.Redirect = fmt.Sprintf("%s/invoice.coin?apikey=%s", metadata.SiteUrl, session.ApiKey)
			}
		}
		session.SignalRecalculate = func(session *drawing.Session) {
			invoice, done := GetInvoice(session)
			if !done {
				return
			}
			var companyHeader string
			var date string
			var invoiceID string
			var company string
			var address string
			var email string
			var amount string = "10"
			var unit string = "USD 1"
			var total string = "USD 10"
			var status string = ""
			err := englang.Scanf(invoice, metadata.InvoicePattern,
				&companyHeader, &date, &invoiceID,
				&company, &address, &email, &amount, &unit, &total, &status)
			if err != nil {
				return
			}

			chg := session.Text[InvoiceText]
			chg.Text = invoice
			session.Text[InvoiceText] = chg
		}
		session.SignalRecalculate(session)
	}
}

func GetInvoice(session *drawing.Session) (string, bool) {
	order, ok := orders[session.ApiKey]
	if !ok {
		return "", false
	}
	return order, true
}

func RedeemCoin(upload string) (string, error) {
	scanner := bufio.NewScanner(bytes.NewBufferString(upload))
	for scanner.Scan() {
		var voucher, begin, end, site string
		err := englang.ScanfContains(scanner.Text()+".", "http%s/voucher.html?apikey=%s.", &begin, &site, &voucher, &end)
		if err == nil {
			ok, isInvoice, _, valid := ValidateVoucherKey(voucher, true)
			if ok {
				buf := bytes.NewBufferString("")
				if isInvoice {
					buf.WriteString(fmt.Sprintf("\nInvoice used: %s\n", drawing.RedactPublicKey(voucher)))
				}
				buf.WriteString(fmt.Sprintf("\nVoucher used: %s\n", drawing.RedactPublicKey(valid)))
				return buf.String(), nil
			}
		}
	}
	return "", fmt.Errorf("not found")
}
