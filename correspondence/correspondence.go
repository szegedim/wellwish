package correspondence

import (
	"gitlab.com/eper.io/engine/drawing"
	"image/color"
	"net/http"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// This is just a package to show a user interface concept.
// The basic design is that you find what you need right away.
// Modern software was designed to use swipes, drags, and scrolls
// to showcase the superiority of newer and newer semiconductor
// components over the preceding generation of such products.
// Moore's Law and growth helped to fund research for a long time.
//
// Some customers on the other hand need the right information, right there.
// It is a connector that can attach people, documents, and tasks.

func SetupCorrespondence() {
	http.HandleFunc("/correspondence.html", func(w http.ResponseWriter, r *http.Request) {
		err := drawing.EnsureAPIKey(w, r)
		if err != nil {
			return
		}
		drawing.ServeRemoteForm(w, r, "correspondence")
	})

	http.HandleFunc("/correspondence.png", func(w http.ResponseWriter, r *http.Request) {
		drawing.ServeRemoteFrame(w, r, declareCorrespondenceForm)
	})
}

func declareCorrespondenceForm(session *drawing.Session) {
	if session.Form.Boxes == nil {
		drawing.DeclareForm(session, "./correspondence/res/correspondence.png")
		var MakeWidget = 0
		drawing.PutText(session, MakeWidget, drawing.Content{Text: "hello", Lines: 1, Editable: true, Selectable: true, FontColor: color.Black, BackgroundColor: color.White})
		checked := drawing.CombineImages("./correspondence/res/person.png", "./correspondence/res/check.png")
		unchecked := "./correspondence/res/person.png"
		empty := "./correspondence/res/placeholder.png"
		for i := 1; i < 65; i++ {
			if i < 9 {
				drawing.SetImage(session, i, unchecked, drawing.Content{Selectable: true, Editable: false})
				continue
			}
			drawing.SetImage(session, i, empty, drawing.Content{Selectable: false, Editable: false})
		}
		session.SignalFocusChanged = func(session *drawing.Session, from int, to int) {
			if session.Text[to].Text == "" && session.Text[to].BackgroundFile != empty {
				drawing.SetImage(session, to, checked, drawing.Content{Selectable: true, Editable: false})
				session.SignalPartialRedrawNeeded(session, to)
			}
			if session.Text[from].BackgroundFile == checked {
				drawing.SetImage(session, from, unchecked, drawing.Content{Selectable: true, Editable: false})
				session.SignalPartialRedrawNeeded(session, from)
			}
		}
		session.SignalClicked = func(session *drawing.Session, i int) {
			session.SignalPartialRedrawNeeded(session, i)
		}
		session.SignalTextChange = func(session *drawing.Session, i int, from string, to string) {
			session.SignalPartialRedrawNeeded(session, i)
		}
	}
}
