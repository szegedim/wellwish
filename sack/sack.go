package sack

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/crypto"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/metadata"
	"io"
	"net/http"
	"os"
	"path"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Storage is an important part of any cloud solution.
//
// Our engine has the following benefits over traditional implementations
//  - we use an apikey ticket per entry
//  - the client does not need to log in with an asymmetric key protocol
//  - we work with cUrl and http directly, there is no need to mess with REST
//  - complex data sets like directories can use tarballs
//  - the sack disposes itself after expiry reducing privacy risks with a magnitude
//  - do not use them as a backup
//  - buy and upload to multiple sacks to support redundancy

// They can be used to send large attachments to clients like DropBox.
// They can be used as interim datasets for data streaming queries.
// They can hold your code temporarily as a backup until it is committed.
// They can be a significant part of your continuous delivery pipeline.

//TODO Seek and put into the middle of a tarball

func Setup() {
	http.HandleFunc("/sack.html", func(w http.ResponseWriter, r *http.Request) {
		err := drawing.EnsureAPIKey(w, r)
		if err != nil {
			return
		}
		drawing.ServeRemoteForm(w, r, "sack")
	})

	http.HandleFunc("/sack.png", func(w http.ResponseWriter, r *http.Request) {
		drawing.ServeRemoteFrame(w, r, declareForm)
	})

	http.HandleFunc("/tmp", func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.URL.Query().Get("apikey")
		if apiKey == "" {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		sack := apiKey
		redirect := ""
		if Sacks[sack] == "" && r.Method == "PUT" {
			invoice := sack
			ok, isInvoice, _, voucher := billing.ValidateVoucherKey(invoice, true)
			if !ok {
				w.WriteHeader(http.StatusPaymentRequired)
				return
			}
			sack = MakeSack(voucher)
			if isInvoice {
				redirect = fmt.Sprintf("/tmp?apikey=%s", voucher)
			}
		}
		if Sacks[sack] == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fileName := sack
		p := path.Join(fmt.Sprintf("/tmp/%s", fileName))
		if r.Method == "GET" {
			http.ServeFile(w, r, p)
			return
		}
		if r.Method == "TRACE" {
			bw := bufio.NewWriter(w)
			size := int64(0)
			stat, err := os.Stat(p)
			if err != nil {
				size = 0
			}
			size = stat.Size()
			_, _ = bw.WriteString(fmt.Sprintf("This is a sack storage of a single file\n"))
			_, _ = bw.WriteString(fmt.Sprintf("The current size is %d bytes.\n", size))
			_, _ = bw.WriteString(fmt.Sprintf("The sack record is the following\n%s\n", Sacks[sack]))
			_ = bw.Flush()
			return
		}
		if r.Method == "PUT" || r.Method == "DELETE" {
			_ = os.Remove(p)
		}
		if r.Method == "PUT" {
			f := drawing.NoErrorFile(os.Create(p))
			defer func() { _ = f.Close() }()
			_, _ = io.Copy(f, r.Body)
			if redirect != "" {
				http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
				return
			}
			return
		}
		if r.Method == "DELETE" {
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
}

func declareForm(session *drawing.Session) {
	if session.Form.Boxes == nil {
		drawing.DeclareForm(session, "./sack/media/page.png")

		init := "Click here to pay with a coin file."
		if Sacks[session.ApiKey] != "" {
			init = "Click here to preview sack."
		}
		CommandText := drawing.DeclareTextField(session, -1, drawing.ActiveContent{Text: init, Lines: 1, Editable: false, Selectable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})

		session.SignalClicked = func(session *drawing.Session, i int) {
			if i == CommandText {
				if session.Text[CommandText].Text == "Click here to pay with a coin file." {
					session.Upload = "coin"
				}
				if session.Text[CommandText].Text == "Click here to upload content." {
					session.Upload = "*.*"
				}
				if session.Text[CommandText].Text == "Click here to preview sack." && Sacks[session.ApiKey] != "" {
					session.Redirect = fmt.Sprintf("/tmp?apikey=%s", session.ApiKey)
					session.SelectedBox = -1
				}
				session.SignalPartialRedrawNeeded(session, i)
			}
		}

		session.SignalUploaded = func(session *drawing.Session, upload drawing.Upload) {
			if session.Text[CommandText].Text == "Click here to pay with a coin file." {
				// session.Data is going to the voucher id
				session.Data = MakeSackWithCoin(string(upload.Body))
				if session.Data != "" {
					drawing.DeclareTextField(session, CommandText, drawing.ActiveContent{Text: "Click here to upload content.", Lines: 1, Editable: false, Selectable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
					session.SignalPartialRedrawNeeded(session, CommandText)
					return
				} else {
					data := session.Text[CommandText]
					data.Text = "The coin did not contain any valid vouchers. Click refresh."
					session.Text[CommandText] = data
					session.SignalPartialRedrawNeeded(session, CommandText)
				}
			}
			if session.Text[CommandText].Text == "Click here to upload content." && session.Data != "" {
				// session.Data is the voucher id
				sack := session.Data
				fileName := sack
				p := path.Join(fmt.Sprintf("/tmp/%s", fileName))
				_ = os.WriteFile(p, upload.Body, 0700)
				session.Data = ""

				data := session.Text[CommandText]
				data.Text = "Redirecting..."
				session.Text[CommandText] = data
				session.SignalPartialRedrawNeeded(session, CommandText)

				session.SelectedBox = -1
				session.Redirect = fmt.Sprintf("/sack.html?apikey=%s", sack)
			}
		}

		session.SignalTextChange = func(session *drawing.Session, i int, from string, to string) {
			session.SignalPartialRedrawNeeded(session, i)
		}
	}
}

func MakeSackWithCoin(upload string) string {
	scanner := bufio.NewScanner(bytes.NewBufferString(upload))
	for scanner.Scan() {
		var voucher, begin, end, site string
		err := englang.ScanfContains(scanner.Text()+".", "http%s/voucher.html?apikey=%s.", &begin, &site, &voucher, &end)
		if err == nil {
			ok, isInvoice, _, valid := billing.ValidateVoucherKey(voucher, true)
			if ok {
				sack := MakeSack(valid)
				if isInvoice {
					Sacks[sack] = Sacks[sack] + fmt.Sprintf("\nInvoice used: %s\n", drawing.RedactPublicKey(voucher))
				}
				Sacks[sack] = Sacks[sack] + fmt.Sprintf("\nVoucher used: %s\n", drawing.RedactPublicKey(valid))
				return sack
			}
		}
	}
	return ""
}

func MakeSack(sack string) string {
	trace := fmt.Sprintf(crypto.TicketExpiry, time.Now().Add(4*168*time.Hour).Format("Jan 2, 2006"))
	Sacks[sack] = trace
	path1 := path.Join(fmt.Sprintf("/tmp/%s", sack))
	newSack := drawing.NoErrorFile(os.Create(path1))
	w := bufio.NewWriter(newSack)
	_, _ = w.WriteString(fmt.Sprintf("curl -X GET %s/tmp?apikey=%s", metadata.SiteUrl, sack))
	_ = w.Flush()
	_ = newSack.Close()
	return sack
}