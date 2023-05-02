package activation

import (
	"fmt"
	drawing "gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
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

// The activation module does not allow any production modules to run until it gets activated.
// This helps to make sure we can always get into a consistent state at startup.
// Each deployment should update the activation key in metadata.ActivationKey
// Activation is propagated automatically to all servers participating in the mesh
// The activation key can be entered from the UI on startup, which redirects to the management UI.
// Bookmark the management url with ApiKey to manage the site. There is no other way to access backups, etc.
// TODO management apikey rotation.

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
		management.QuantumGradeAuthorization()
		if metadata.ActivationKey == "" {
			// Already activated
			return
		}
		activationKey := r.URL.Query().Get("apikey")
		if activationKey == metadata.ActivationKey {
			adminKey := startActivation()
			_, _ = w.Write([]byte(fmt.Sprintf("%s/management.html?apikey=%s", metadata.SiteUrl, adminKey)))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

	go func() {
		for {
			if metadata.ActivationKey == "" {
				break
			}
			if mesh.Index[metadata.ActivationKey] != "" {
				management.UpdateAdminKey(mesh.Index[metadata.ActivationKey])
				activate()
				break
			}
			time.Sleep(time.Second)
		}
	}()
}

func startActivation() string {
	adminKey := drawing.GenerateUniqueKey()
	mesh.Index[metadata.ActivationKey] = adminKey
	return adminKey
}

func declareActivationForm(session *drawing.Session) {
	if session.Form.Boxes == nil {
		drawing.DeclareForm(session, "./activation/res/activate.png")
		drawing.DeclareTextField(session, 0, drawing.ActiveContent{Text: drawing.Revert + "Enter the activation key", Lines: 1, Editable: true, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
		session.SignalTextChange = func(session *drawing.Session, i int, from string, to string) {
			session.SignalPartialRedrawNeeded(session, i)
			if strings.Contains(session.Text[i].Text, metadata.ActivationKey) {
				adminKey := startActivation()
				session.Data = fmt.Sprintf("/management.html?apikey=%s", adminKey)
				session.SignalClosed(session)
			}
		}
		session.SignalClosed = func(session *drawing.Session) {
			for {
				if metadata.ActivationKey == "" {
					break
				}
			}
			session.SelectedBox = -1
			session.Redirect = session.Data
		}
	}
}

func activate() string {
	Activated <- "Hello World!"
	adminKey := <-Activated
	fmt.Println("Activated.", adminKey)
	metadata.ActivationKey = ""
	return adminKey
}
