package bag

import (
	"bufio"
	"bytes"
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
//  - the bag disposes itself after expiry reducing privacy risks with a magnitude
//  - do not use them as a long term backup but as a cache instead
//  - buy and upload to multiple bags to support redundancy
//  - you can do RAID 0, 1, 2, 3 etc. from simple client scripts if needed

// bag stands for a bag of grain for example.
//
// Benefits of bags:
// They can be used to send large attachments to clients like DropBox.
// They can be used as interim datasets for data streaming queries.
// They can hold your code temporarily as a backup until it is committed.
// They can be a low latency part of your continuous delivery pipeline.
// Want to extend a bag period? Let it delete and create a new one. Reason? Newly generated ids are safer.

// TODO Seek and put into the middle of a tarball

func Setup() {
	stateful.RegisterModuleForBackup(&bags)

	http.HandleFunc("/bag.html", func(w http.ResponseWriter, r *http.Request) {
		if nil == mesh.RedirectToPeerServer(w, r) {
			return
		}
		err := drawing.EnsureAPIKey(w, r)
		if err != nil {
			return
		}
		drawing.ServeRemoteForm(w, r, "bag")
	})

	http.HandleFunc("/bag.png", func(w http.ResponseWriter, r *http.Request) {
		if nil == mesh.RedirectToPeerServer(w, r) {
			return
		}
		drawing.ServeRemoteFrame(w, r, declareForm)
	})

	http.HandleFunc("/tmp.coin", func(w http.ResponseWriter, r *http.Request) {
		if nil == mesh.RedirectToPeerServer(w, r) {
			return
		}
		// Setup burst sessions, a range of time, when a coin can be used for bursts.
		if r.Method == "PUT" {
			coinToUse := billing.ValidatedCoinContent(w, r)
			if coinToUse != "" {
				bag := MakeBagInternal(coinToUse)
				management.QuantumGradeAuthorization()
				_, _ = w.Write([]byte(bag))
				return
			}
			management.QuantumGradeAuthorization()
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}

		if r.Method == "GET" {
			apiKey := r.URL.Query().Get("apikey")
			session, sessionValid := bags[apiKey]
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
		if nil == mesh.RedirectToPeerServer(w, r) {
			return
		}
		apiKey := r.URL.Query().Get("apikey")

		bag := apiKey
		if !mesh.CheckExpiry(bag) {
			r.Method = "DELETE"
		}

		traces := bags[bag]
		if traces == "" || mesh.GetIndex(bag) == "" {
			management.QuantumGradeAuthorization()
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		fileName := bag
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
			drawing.NoErrorWrite(bw.WriteString(fmt.Sprintf("This is a bag storage of a single file\n")))
			drawing.NoErrorWrite(bw.WriteString(fmt.Sprintf("The current size is %d bytes.\n", size)))
			drawing.NoErrorWrite(bw.WriteString(fmt.Sprintf("bag record follows\n%s\n", bags[bag])))
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
			delete(bags, bag)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	go func() {
		for {
			if len(bags) > 0 {
				nanos := time.Duration(metadata.CheckpointPeriod.Nanoseconds() / int64(len(bags)))
				for bag := range bags {
					CleanupExpiredbag(bag)
					time.Sleep(nanos)
				}
			}
			time.Sleep(metadata.CheckpointPeriod)
		}
	}()
}

func CleanupExpiredbag(bag string) {
	valid := mesh.GetIndex(bag)
	if valid == "" {
		path1 := path.Join(fmt.Sprintf("/tmp/%s", bag))
		_ = os.Remove(path1)
		delete(bags, bag)
	}
}

func MakebagWithCoin(coinUrlList string) string {
	scanner := bufio.NewScanner(bytes.NewBufferString(coinUrlList))
	for scanner.Scan() {
		var voucher, begin, end, site string
		err := englang.ScanfContains(scanner.Text()+".", "http%s/voucher.html?apikey=%s.", &begin, &site, &voucher, &end)
		if err == nil {
			ok, isInvoice, _, valid := billing.ValidateVoucherKey(voucher, true)
			if ok {
				bag := MakeBagInternal(valid)
				if isInvoice {
					bags[bag] = bags[bag] + fmt.Sprintf("\nInvoice used: %s\n", drawing.RedactPublicKey(voucher))
				}
				bags[bag] = bags[bag] + fmt.Sprintf("\nVoucher used: %s\n", drawing.RedactPublicKey(valid))
				return bag
			}
		}
	}
	return ""
}

func declareForm(session *drawing.Session) {
	if session.Form.Boxes == nil {
		drawing.DeclareForm(session, "./bag/media/page.png")

		init := "Click here to pay with a coin file."
		if bags[session.ApiKey] != "" {
			init = "Click here to preview bag."
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
				if session.Text[CommandText].Text == "Click here to preview bag." && bags[session.ApiKey] != "" {
					session.Redirect = fmt.Sprintf("/tmp?apikey=%s", session.ApiKey)
					session.SelectedBox = -1
				}
				session.SignalPartialRedrawNeeded(session, i)
			}
		}

		session.SignalUploaded = func(session *drawing.Session, upload drawing.Upload) {
			if session.Text[CommandText].Text == "Click here to pay with a coin file." {
				// session.Data is going to the voucher id
				session.Data = MakebagWithCoin(string(upload.Body))
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
				bag := session.Data
				fileName := bag
				p := path.Join(fmt.Sprintf("/tmp/%s", fileName))
				_ = os.WriteFile(p, upload.Body, 0700)
				session.Data = ""

				data := session.Text[CommandText]
				data.Text = "Redirecting..."
				session.Text[CommandText] = data
				session.SignalPartialRedrawNeeded(session, CommandText)

				session.SelectedBox = -1
				session.Redirect = fmt.Sprintf("/bag.html?apikey=%s", bag)
			}
		}

		session.SignalTextChange = func(session *drawing.Session, i int, from string, to string) {
			session.SignalPartialRedrawNeeded(session, i)
		}
	}
}

func MakeBagInternal(bag string) string {
	bags[bag] = "Bag is valid."
	mesh.RegisterIndex(bag)
	mesh.SetExpiry(bag, ValidPeriod)
	path1 := path.Join(fmt.Sprintf("/tmp/%s", bag))
	bagFile := drawing.NoErrorFile(os.Create(path1))
	w := bufio.NewWriter(bagFile)
	_, _ = w.WriteString(fmt.Sprintf("curl -X GET %s/tmp?apikey=%s", metadata.SiteUrl, bag))
	_ = w.Flush()
	_ = bagFile.Close()
	return bag
}

func GetBagPathInternal(bag string) string {
	fileName := bag
	return path.Join(fmt.Sprintf("/tmp/%s", fileName))
}

func GetBagInternal(bag string) []byte {
	return drawing.NoErrorBytes(os.ReadFile(GetBagPathInternal(bag)))
}
