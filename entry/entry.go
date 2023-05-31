package entry

import (
	"fmt"
	"gitlab.com/eper.io/engine/drawing"
	"gitlab.com/eper.io/engine/metadata"
	"net/http"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// This is the entry screen of the applet.
// It is the first point when the customer enters.
//
// The basic design is what property investors suggest, when selling a house.
// Color and painting is the cheapest option to upgrade your product.
// We use the botanical style that was randomly selected by rolling a die.
// It is colorful to attract and get noticed.
//
// Customers have not paid yet.
// Clicking to a paid option will lead them to the payment tab.

func Setup() {
	http.HandleFunc("/entry.html", func(w http.ResponseWriter, r *http.Request) {
		err := drawing.EnsureAPIKey(w, r)
		if err != nil {
			return
		}
		drawing.ServeRemoteForm(w, r, "entry")
	})

	http.HandleFunc("/entry.png", func(w http.ResponseWriter, r *http.Request) {
		drawing.ServeRemoteFrame(w, r, declareCorrespondenceForm)
	})
}

func declareCorrespondenceForm(session *drawing.Session) {
	if session.Form.Boxes == nil {
		drawing.DeclareForm(session, "./entry/media/entry.png")

		drawing.PutText(session, -1, drawing.Content{Text: metadata.SiteName, Lines: 2, Editable: false, Selectable: false, FontColor: drawing.White, BackgroundColor: drawing.Black, Alignment: 1})
		DocumentButton := drawing.SetImage(session, -1, "./entry/media/document.png", drawing.Content{Selectable: false, Editable: false})
		_ = drawing.SetImage(session, -1, "./entry/media/mine.png", drawing.Content{Selectable: false, Editable: false})
		CheckoutButton := drawing.SetImage(session, -1, "./entry/media/cart.png", drawing.Content{Selectable: false, Editable: false})
		TermsButton := drawing.SetImage(session, -1, "./entry/media/terms.png", drawing.Content{Selectable: false, Editable: false})
		ContactButton := drawing.SetImage(session, -1, "./entry/media/contact.png", drawing.Content{Selectable: false, Editable: false})

		session.SignalClicked = func(session *drawing.Session, i int) {
			if i == CheckoutButton {
				session.Redirect = fmt.Sprintf("/checkout.html?apikey=%s", drawing.GenerateUniqueKey())
				session.SignalClosed(session)
			}
			if i == TermsButton {
				session.Redirect = "/terms.txt"
				session.SignalClosed(session)
			}
			if i == ContactButton {
				session.Redirect = fmt.Sprintf("mailto:hq@schmied.us?subject=Question%%20About%%20%s", metadata.SiteUrl)
				session.SignalClosed(session)
			}
			if i == DocumentButton {
				session.Redirect = fmt.Sprintf("/bag.html?apikey=%s", drawing.GenerateUniqueKey())
				session.SignalClosed(session)
			}
		}
		session.SignalTextChange = func(session *drawing.Session, i int, from string, to string) {
			session.SignalPartialRedrawNeeded(session, i)
		}
	}
}
