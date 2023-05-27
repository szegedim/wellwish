package activation

import (
	"fmt"
	drawing "gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/englang"
	"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/mesh"
	"gitlab.com/eper.io/engine/metadata"
	"net/http"
	"os"
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

// If the activation key is empty, no activation is needed.
// Containers start right away. However, you cannot access any management features either.

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
		// Activation key is private updated in metadata.go
		// We make sure it is used only once to retrieve the management key for all the data
		if activationKey == metadata.ActivationKey {
			if ActivationNeeded {
				adminKey := startActivation()
				time.Sleep(activationPeriod)
				_, _ = w.Write([]byte(adminKey))
			} else {
				// TODO is this secure? Give them just a hint, if they have the private key
				_, _ = w.Write([]byte(drawing.RedactPublicKey(metadata.ManagementKey)))
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})

	go func() {
		key, ok := os.LookupEnv("KEY")
		if ok && key != "" {
			metadata.ActivationKey = key
		}
		fmt.Println(englang.Printf("Activate with %s .", metadata.ActivationKey))
		if metadata.ActivationKey == "" {
			activate()
			return
		}
		for {
			if metadata.ActivationKey == "" {
				break
			}

			adminKey := mesh.GetIndex(metadata.ActivationKey)
			if adminKey != "" {
				// This happens, if another node in the mesh was activated
				metadata.ManagementKey = adminKey
				activate()
				break
			}
			time.Sleep(activationPeriod)
		}
	}()
}

// startActivation happens once in a cluster and the key gets propagated.
func startActivation() string {
	metadata.ManagementKey = drawing.GenerateUniqueKey()
	mesh.SetIndex(metadata.ActivationKey, metadata.ManagementKey)
	return metadata.ManagementKey
}

func declareActivationForm(session *drawing.Session) {
	if session.Form.Boxes == nil {
		drawing.DeclareForm(session, "./activation/res/activate.png")
		drawing.SetImage(session, 0, "./metadata/logo.png", drawing.Content{Text: "", Lines: 1, Editable: false, FontColor: drawing.White, BackgroundColor: drawing.Black, Alignment: 1})
		drawing.PutText(session, 1, drawing.Content{Text: drawing.RevertAndReturn + "Click here and enter the activation key", Lines: 1, Editable: true, FontColor: drawing.Black, BackgroundColor: drawing.White, Alignment: 0})
		session.SignalTextChange = func(session *drawing.Session, i int, from string, to string) {
			session.SignalPartialRedrawNeeded(session, i)
			if strings.Contains(session.Text[i].Text, metadata.ActivationKey) {
				// Activation key is private updated in metadata.go
				// We make sure it is used only once to retrieve the management key for all the data
				if ActivationNeeded {
					adminKey := startActivation()
					time.Sleep(activationPeriod)
					session.Data = fmt.Sprintf("/management.html?apikey=%s", adminKey)
					session.SignalClosed(session)
				} else {
					session.Data = fmt.Sprintf("/")
					session.SignalClosed(session)
				}
			}
		}
		session.SignalClosed = func(session *drawing.Session) {
			session.SelectedBox = -1
			session.Redirect = session.Data
		}
	}
}

func activate() {
	Activated <- "Hello World!"
	<-Activated
	// TODO is this secure?
	fmt.Println("Activated.", fmt.Sprintf("%s/management.html?apikey=%s...", metadata.SiteUrl, drawing.RedactPublicKey(metadata.ManagementKey)))
	ActivationNeeded = false
}
