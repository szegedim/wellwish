package activation

import (
	"fmt"
	drawing "gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/mesh"
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

func SetupActivation() {
	http.HandleFunc("/activate.html", func(w http.ResponseWriter, r *http.Request) {
		err := drawing.EnsureAPIKey(w, r)
		if err != nil {
			return
		}
		drawing.ServeRemoteForm(w, r, "activate")
	})

	http.HandleFunc("/activate.png", func(w http.ResponseWriter, r *http.Request) {
		drawing.ServeRemoteFrame(w, r, declareActivationForm)
	})

	http.HandleFunc("/activate", func(w http.ResponseWriter, r *http.Request) {
		if metadata.ActivationKey == "" {
			// Already activated
			return
		}
		adminKeyCandidate := r.URL.Query().Get("apikey")
		activationKey := r.URL.Query().Get("activationkey")
		management.QuantumGradeAuthorization()
		mesh.ForwardRoundRobinRingRequest(r)
		if activationKey == metadata.ActivationKey {
			management.UpdateAdminKey(adminKeyCandidate)
			adminKey := Activate()
			_, _ = w.Write([]byte(fmt.Sprintf("%s/management.html?apikey=%s", metadata.SiteUrl, adminKey)))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})
}

func declareActivationForm(session *drawing.Session) {
	if session.Form.Boxes == nil {
		drawing.DeclareForm(session, "./activation/res/activate.png")
		drawing.DeclareTextField(session, 0, drawing.ActiveContent{Text: drawing.Revert + "Enter the activation key", Lines: 1, Editable: true, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
		session.SignalTextChange = func(session *drawing.Session, i int, from string, to string) {
			session.SignalPartialRedrawNeeded(session, i)
			if strings.Contains(session.Text[i].Text, metadata.ActivationKey) {
				adminKey := Activate()
				session.Data = fmt.Sprintf("/management.html?apikey=%s", adminKey)
				session.SignalClosed(session)
			}
		}
		session.SignalClosed = func(session *drawing.Session) {
			session.SelectedBox = -1
			session.Redirect = session.Data
		}
	}
}

func Activate() string {
	original := metadata.ActivationKey

	metadata.ActivationKey = ""
	Activated <- "Hello World!"
	adminKey := <-Activated
	mesh.ActivateSite(original, adminKey)
	return adminKey
}
