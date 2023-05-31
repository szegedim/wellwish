package sack

import (
	"bufio"
	"bytes"
	"fmt"
	"gitlab.com/eper.io/engine/billing"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"gitlab.com/eper.io/engine/stateful"
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

// Storage is the core part of this cloud solution.
//
// Our engine has the following benefits over traditional implementations
//  - we use an apikey ticket per entry
//  - the client does not need to log in with an asymmetric key protocol
//  - we work with cUrl and http directly, there is no need to mess with REST
//  - complex datasets, directories can use tarballs
//  - the sack disposes itself after expiry reducing privacy risks with a magnitude
//  - do not use them as a long term backup but as a cache instead
//  - buy and upload to multiple sacks to support redundancy
//  - you can do RAID 0, 1, 2, 3 etc. from simple client scripts if needed

// Sack stands for a sack of grain for example.
//
// Benefits of sacks:
// They can be used to send large attachments to clients like DropBox.
// They can be used as interim datasets for data streaming queries.
// They can hold your code temporarily as a backup until it is committed.
// They can be a low latency part of your continuous delivery pipeline.

// TODO Seek and put into the middle of a tarball

func Setup() {
	stateful.RegisterModuleForBackup(&Sacks)

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

	http.HandleFunc("/tmp.coin", func(w http.ResponseWriter, r *http.Request) {
		// Setup burst sessions, a range of time, when a coin can be used for bursts.
		if r.Method == "PUT" {
			coinToUse := billing.ValidatedCoinContent(w, r)
			if coinToUse != "" {
				// Want to extend? Let it delete and create a new one.
				// Reason? Newly generated ids are safer.
				// TODO cleanup
				sack := makeSack(coinToUse)
				// TODO cleanup
				// mesh.SetIndex(burst, mesh.WhoAmI)
				management.QuantumGradeAuthorization()
				_, _ = w.Write([]byte(sack))
				return
			}
			management.QuantumGradeAuthorization()
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}

		if r.Method == "GET" {
			apiKey := r.URL.Query().Get("apikey")
			session, sessionValid := Sacks[apiKey]
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

	http.HandleFunc("/tmp", func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.URL.Query().Get("apikey")

		sack := apiKey
		traces := Sacks[sack]
		if traces == "" {
			management.QuantumGradeAuthorization()
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fileName := sack
		p := path.Join(fmt.Sprintf("/tmp/%s", fileName))
		if r.Method == "GET" {
			management.QuantumGradeAuthorization()
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
			drawing.NoErrorWrite(bw.WriteString(fmt.Sprintf("This is a sack storage of a single file\n")))
			drawing.NoErrorWrite(bw.WriteString(fmt.Sprintf("The current size is %d bytes.\n", size)))
			drawing.NoErrorWrite(bw.WriteString(fmt.Sprintf("Sack record follows\n%s\n", Sacks[sack])))
			drawing.NoErrorVoid(bw.Flush())
			return
		}
		if r.Method == "PUT" || r.Method == "DELETE" {
			drawing.NoErrorVoid(os.Remove(p))
		}
		if r.Method == "PUT" {
			f := drawing.NoErrorFile(os.Create(p))
			defer func() { _ = f.Close() }()
			drawing.NoErrorWrite64(io.Copy(f, r.Body))
			return
		}
		if r.Method == "DELETE" {
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	go func() {
		for {
			if len(Sacks) > 0 {
				nanos := time.Duration(metadata.CheckpointPeriod.Nanoseconds() / int64(len(Sacks)))
				for sack := range Sacks {
					CleanupExpiredSack(sack)
					time.Sleep(nanos)
				}
			}
			time.Sleep(metadata.CheckpointPeriod)
		}
	}()
}

func CleanupExpiredSack(sack string) {
	info := Sacks[sack]
	dt := ""
	begin := ""
	end := ""
	err := englang.ScanfContains(info, billing.TicketExpiry, &begin, &dt, &end)
	if err != nil {
		return
	}
	expiry, err := time.Parse("Jan 2, 2006", dt)
	if err != nil {
		return
	}
	if time.Now().After(expiry) {
		path1 := path.Join(fmt.Sprintf("/tmp/%s", sack))
		_ = os.Remove(path1)
	}
}

func MakeSackWithCoin(coinUrlList string) string {
	scanner := bufio.NewScanner(bytes.NewBufferString(coinUrlList))
	for scanner.Scan() {
		var voucher, begin, end, site string
		err := englang.ScanfContains(scanner.Text()+".", "http%s/voucher.html?apikey=%s.", &begin, &site, &voucher, &end)
		if err == nil {
			ok, isInvoice, _, valid := billing.ValidateVoucherKey(voucher, true)
			if ok {
				sack := makeSack(valid)
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

func declareForm(session *drawing.Session) {
	if session.Form.Boxes == nil {
		drawing.DeclareForm(session, "./sack/media/page.png")

		init := "Click here to pay with a coin file."
		if Sacks[session.ApiKey] != "" {
			init = "Click here to preview sack."
		}
		CommandText := drawing.PutText(session, -1, drawing.Content{Text: init, Lines: 1, Editable: false, Selectable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})

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
					drawing.PutText(session, CommandText, drawing.Content{Text: "Click here to upload content.", Lines: 1, Editable: false, Selectable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
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

func makeSack(sack string) string {
	trace := fmt.Sprintf(billing.TicketExpiry, time.Now().Add(4*168*time.Hour).Format("Jan 2, 2006"))
	Sacks[sack] = trace
	path1 := path.Join(fmt.Sprintf("/tmp/%s", sack))
	newSack := drawing.NoErrorFile(os.Create(path1))
	w := bufio.NewWriter(newSack)
	_, _ = w.WriteString(fmt.Sprintf("curl -X GET %s/tmp?apikey=%s", metadata.SiteUrl, sack))
	_ = w.Flush()
	_ = newSack.Close()
	return sack
}
