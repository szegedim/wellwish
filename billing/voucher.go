package billing

import (
	"bufio"
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/metadata"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

const VoucherInvoicePointer = "%s/invoice.html?apikey=%s"

func SetupVoucher() {
	http.HandleFunc("/voucher.html", func(w http.ResponseWriter, r *http.Request) {
		if drawing.EnsureAPIKey(w, r) != nil {
			return
		}
		if drawing.ResetSession(w, r) != nil {
			return
		}
		drawing.ServeRemoteForm(w, r, "voucher")
	})
	http.HandleFunc("/voucher.png", func(w http.ResponseWriter, r *http.Request) {
		if drawing.EnsureAPIKey(w, r) != nil {
			return
		}
		drawing.ServeRemoteFrame(w, r, declareVoucherForm)
	})
	http.HandleFunc("/invoice.coin", func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.URL.Query().Get("apikey")
		if apiKey == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		ListVouchers(w, r)
	})
	http.HandleFunc("/voucher.coin/validate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if drawing.EnsureAPIKey(w, r) != nil {
			return
		}
		ret, _, _, _ := ValidateVoucher(w, r, false)
		if ret {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	})
}

func declareVoucherForm(session *drawing.Session) {
	if session.Form.Boxes == nil {
		drawing.DeclareForm(session, "./billing/res/voucher.png")

		const VoucherText = 0
		const CancelButton = 1
		const InvoiceButton = 2

		drawing.PutText(session, VoucherText, drawing.Content{Text: "", Lines: 20, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 1})
		drawing.PutText(session, CancelButton, drawing.Content{Text: " Cancel/Refund ", Lines: 1, Selectable: false, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
		drawing.PutText(session, InvoiceButton, drawing.Content{Text: "  Find invoice  ", Lines: 1, Selectable: false, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})

		session.SignalClicked = func(session *drawing.Session, i int) {
			voucher, done := getVoucher(session)
			if !done {
				return
			}
			if i == CancelButton {
				last := voucher
				voucher = CancelVoucher(last, voucher)
				if voucher != last {
					vouchers[session.ApiKey] = voucher
					session.SignalRecalculate(session)
					session.SignalPartialRedrawNeeded(session, VoucherText)
				}
			}
			if i == InvoiceButton {
				var companyHeader string
				var date string
				var invoice string
				var status string = ""
				err := englang.Scanf(voucher, metadata.VoucherPattern,
					&companyHeader, &date, &invoice, &status)
				if err == nil {
					session.Redirect = invoice
				}
			}
		}
		session.SignalRecalculate = func(session *drawing.Session) {
			voucher, done := getVoucher(session)
			if !done {
				return
			}
			var companyHeader string
			var date string
			var invoice string
			var status string = ""
			err := englang.Scanf(voucher, metadata.VoucherPattern,
				&companyHeader, &date, &invoice, &status)
			if err != nil {
				return
			}

			chg := session.Text[VoucherText]
			chg.Text = voucher
			session.Text[VoucherText] = chg
		}
		session.SignalRecalculate(session)
	}
}

func CancelVoucher(current string, voucher string) string {
	message := "Status is cancelled.\nRefund of any payment will be sent within five business days."
	if strings.Contains(current, "Status is valid.") {
		voucher = strings.Replace(current, "Status is valid.", message, 1)
	}
	return voucher
}

func NewVoucher(apiKey string, amount string, issued time.Time) {
	final, err := strconv.ParseInt(amount, 10, 64)
	if err == nil {
		for i := int64(0); i < final; i++ {
			key := drawing.GenerateUniqueKey()
			vouchers[key] = englang.Printf(metadata.VoucherPattern, metadata.CompanyInfo, issued.Format("Jan 2, 2006"),
				fmt.Sprintf(VoucherInvoicePointer, metadata.SiteUrl, apiKey), "Status is valid.")
		}
	}
}

func getVoucher(session *drawing.Session) (string, bool) {
	voucher, ok := vouchers[session.ApiKey]
	if !ok {
		return "", false
	}
	return voucher, true
}

func ListVouchers(w http.ResponseWriter, r *http.Request) {
	apiKey := r.URL.Query().Get("apikey")
	if apiKey == "" {
		http.Redirect(w, r, "/", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=\"%s %s.coin\"", metadata.CompanyName, drawing.RedactPublicKey(apiKey)))
	// ApiKey may point to an invoice key of a valid voucher
	invoiceCandidate := fmt.Sprintf(VoucherInvoicePointer, metadata.SiteUrl, apiKey)
	writer := bufio.NewWriter(w)

	GetCoinFile(invoiceCandidate, writer)
}

func GetCoinFile(invoiceCandidate string, writer *bufio.Writer) {
	for key, voucher := range vouchers {
		// ApiKey may point to a voucher directly
		if strings.Contains(voucher, invoiceCandidate) {
			var companyHeader string
			var issued string
			var invoice string
			var status string = ""
			err := englang.Scanf(voucher, metadata.VoucherPattern,
				&companyHeader, &issued, &invoice, &status)
			if err == nil && status == "Status is valid." {
				t, err := time.Parse("Jan 2, 2006", issued)
				if err == nil && t.Add(365*24*time.Hour).After(time.Now()) {
					_, _ = writer.WriteString(fmt.Sprintf("%s/voucher.html?apikey=%s\n", metadata.SiteUrl, key))
				}
			}
		}
	}
	_, _ = writer.WriteString("Used, expired, invalid, refunded vouchers:\n")
	for key, voucher := range vouchers {
		// ApiKey may point to a voucher directly
		if strings.Contains(voucher, invoiceCandidate) {
			var companyHeader string
			var issued string
			var invoice string
			var status string = ""
			err := englang.Scanf(voucher, metadata.VoucherPattern,
				&companyHeader, &issued, &invoice, &status)
			if err != nil || status != "Status is valid." {
				t, err := time.Parse("Jan 2, 2006", issued)
				if err == nil && t.Add(365*24*time.Hour).After(time.Now()) {
					_, _ = writer.WriteString(fmt.Sprintf("%s/voucher.html?apikey=%s\n", metadata.SiteUrl, key))
				}
			}
		}
	}
	_ = writer.Flush()
}

func ValidateVoucher(w http.ResponseWriter, r *http.Request, consume bool) (bool, bool, string, string) {
	apiKey := r.URL.Query().Get("apikey")
	return ValidateVoucherKey(apiKey, consume)
}

func ValidateVoucherKey(apiKey string, consume bool) (bool, bool, string, string) {
	// TODO management.QuantumGradeAuthorization()
	// ApiKey may point to an invoice key of a valid voucher
	invoiceCandidate := fmt.Sprintf(VoucherInvoicePointer, metadata.SiteUrl, apiKey)
	for key, voucher := range vouchers {
		isInvoice := strings.Contains(voucher, invoiceCandidate)
		// ApiKey may point to a voucher directly
		if apiKey == key || isInvoice {
			var companyHeader string
			var issued string
			var invoice string
			var status string = ""
			err := englang.Scanf(voucher, metadata.VoucherPattern,
				&companyHeader, &issued, &invoice, &status)
			if err == nil && status == "Status is valid." {
				t, err := time.Parse("Jan 2, 2006", issued)
				if err == nil && t.Add(365*24*time.Hour).After(time.Now()) {
					if !consume {
						return true, isInvoice, invoice, key
					}
					status = "Status is used."
					vouchers[key] = englang.Printf(metadata.VoucherPattern, companyHeader, issued, invoice, status)
					return true, isInvoice, invoice, key
				}
			}
		}
	}
	return false, false, "", ""
}
