package management

import (
	"fmt"
	drawing "gitlab.com/eper.io/engine/drawing"
	"io"
	"net/http"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func SetupSiteManagement(administrationKeySet string, traces func(m string, w io.Writer, r io.Reader)) {
	administrationKey = administrationKeySet
	CheckpointFunc = traces

	http.HandleFunc("/management.html", func(w http.ResponseWriter, r *http.Request) {
		_, err := EnsureAdministratorSession(w, r)
		if err != nil {
			return
		}
		drawing.ServeRemoteForm(w, r, "management")
	})

	http.HandleFunc("/management.png", func(w http.ResponseWriter, r *http.Request) {
		_, err := EnsureAdministratorSession(w, r)
		if err != nil {
			return
		}
		drawing.ServeRemoteFrame(w, r, declareForm)
	})
	http.HandleFunc("/logs.md", func(w http.ResponseWriter, r *http.Request) {
		_, err := EnsureAdministratorSession(w, r)
		if err != nil {
			return
		}

		if r.Method == "GET" {
			w.Header().Set("Content-Type", "text/plain")
			CheckpointFunc("GET", w, nil)
		}
		if r.Method == "PUT" {
			CheckpointFunc("PUT", nil, r.Body)
		}
	})
}

func EnsureAdministrator(w http.ResponseWriter, r *http.Request) (string, error) {
	apiKey := r.URL.Query().Get("apikey")

	time.Sleep(15 * time.Millisecond)
	if apiKey != administrationKey {
		w.WriteHeader(http.StatusUnauthorized)
		return "", fmt.Errorf("unauthorized")
	}
	return apiKey, nil
}

func EnsureAdministratorSession(w http.ResponseWriter, r *http.Request) (*drawing.Session, error) {
	_, err := EnsureAdministrator(w, r)
	if err != nil {
		return nil, err
	}
	session := drawing.GetSession(w, r)
	if session == nil {
		return nil, fmt.Errorf("no session")
	}
	return session, nil
}

func declareForm(session *drawing.Session) {
	if session.Form.Boxes == nil {
		const Contact = 0
		const Logs = 1
		const PublicSite = 2
		const PrivateSite = 3
		drawing.DeclareForm(session, "./management/res/management.png")
		drawing.DeclareImageField(session, Contact, "./drawing/res/space.png", drawing.ActiveContent{Text: "", Lines: 1, Editable: false, FontColor: drawing.White, BackgroundColor: drawing.Black, Alignment: 1})
		drawing.DeclareTextField(session, Logs, drawing.ActiveContent{Text: "     Traces     ", Lines: 1, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
		drawing.DeclareTextField(session, PublicSite, drawing.ActiveContent{Text: "     Public     ", Lines: 1, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
		drawing.DeclareTextField(session, PrivateSite, drawing.ActiveContent{Text: "     Private    ", Lines: 1, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})

		session.SignalClicked = func(session *drawing.Session, i int) {
			if i == Contact {
				session.Redirect = fmt.Sprintf("mailto:hq@schmied.us")
				session.SelectedBox = -1
			}
			if i == Logs {
				session.Redirect = fmt.Sprintf("/logs.md?apikey=%s", session.ApiKey)
				session.SelectedBox = -1
			}
			if i == PublicSite {
				session.Redirect = fmt.Sprintf("/index.html")
				session.SelectedBox = -1
			}
			if i == PrivateSite {
				session.Redirect = fmt.Sprintf("/checkout.html?apikey=%s", session.ApiKey)
				session.SelectedBox = -1
			}
		}
	}
}

func LogSnapshot(m string, w io.Writer, r io.Reader) {
	_, _ = w.Write([]byte(fmt.Sprintf("This container is running with management key %s ...\n\n", drawing.RedactPublicKey(administrationKey))))
}

func Restore(w http.ResponseWriter, r *http.Request) {
	//apiKey := r.URL.Query().Get("apikey")
	//if apiKey != "" {
	//	_, _ = w.Write([]byte(fmt.Sprintf("admin:%s\n\n", drawing.RedactPublicKey(apiKey))))
	//}
	//activation.LogSnapshot(w, r)
	//billing.LogSnapshot(w, r)
	//mining.LogSnapshot(w, r)
	//sack.LogSnapshot(w, r)
}
