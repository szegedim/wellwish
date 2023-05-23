package management

import (
	"fmt"
	drawing "gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/metadata"
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

func SetupSiteManagement(traces func(m string, w io.Writer, r io.Reader)) {
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

func IsAdministrator(apiKey string) error {
	time.Sleep(15 * time.Millisecond)
	administrationKey := GetAdminKey()
	if administrationKey == "" || apiKey != administrationKey {
		return fmt.Errorf("unauthorized")
	}
	return nil
}

func EnsureAdministrator(w http.ResponseWriter, r *http.Request) (string, error) {
	apiKey := r.URL.Query().Get("apikey")

	time.Sleep(15 * time.Millisecond)
	administrationKey := GetAdminKey()
	if apiKey != administrationKey {
		w.WriteHeader(http.StatusUnauthorized)
		return "", fmt.Errorf("unauthorized")
	}
	return apiKey, nil
}

func GetAdminKey() string {
	return metadata.ManagementKey
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
		const Logo = 0
		const Contact = 1
		const Logs = 2
		const PublicSite = 3
		const PrivateSite = 4
		const Backup = 5
		const Restore = 6
		drawing.DeclareForm(session, "./management/res/management.png")
		drawing.SetImage(session, Logo, "./metadata/logo.png", drawing.Content{Text: "", Lines: 1, Editable: false, FontColor: drawing.White, BackgroundColor: drawing.Black, Alignment: 1})
		drawing.SetImage(session, Contact, "./drawing/res/space.png", drawing.Content{Text: "", Lines: 1, Editable: false, FontColor: drawing.White, BackgroundColor: drawing.Black, Alignment: 1})
		drawing.PutText(session, Logs, drawing.Content{Text: "     Traces     ", Lines: 1, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
		drawing.PutText(session, PublicSite, drawing.Content{Text: "     Public     ", Lines: 1, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
		drawing.PutText(session, PrivateSite, drawing.Content{Text: "     Private    ", Lines: 1, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
		drawing.PutText(session, Backup, drawing.Content{Text: "     Backup     ", Lines: 1, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
		drawing.PutText(session, Restore, drawing.Content{Text: "     Restore    ", Lines: 1, Editable: false, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})

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
			if i == Backup {
				session.Redirect = fmt.Sprintf("/backup.checkpoint?apikey=%s", session.ApiKey)
				session.SelectedBox = -1
			}
			if i == Restore {
				session.Upload = "checkpoint"
			}
		}
		session.SignalUploaded = func(session *drawing.Session, upload drawing.Upload) {
			err := IsAdministrator(session.ApiKey)
			if err != nil {
				return
			}
			//body := bytes.NewBuffer(upload.Body)
			//mesh.HttpRequest("PUT", "/")
			//CheckpointFunc("PUT", nil, body)
		}
	}
}

func LogSnapshot(m string, w io.Writer, r io.Reader) {
	if m == "GET" {
		_, _ = w.Write([]byte(fmt.Sprintf("This container is running with management key %s ...\n\n", drawing.RedactPublicKey(GetAdminKey()))))
	}
}
