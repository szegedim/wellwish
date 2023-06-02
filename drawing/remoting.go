package drawing

import (
	"fmt"
	"time"

	//"gitlab.com/eper.io/engine/management"
	"gitlab.com/eper.io/engine/metadata"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// This is the actual remote desktop library implemented with pngs.
// It is a bit slower, but compatible, but we have a paid version that is faster.

type Session struct {
	ApiKey                    string
	Mutex                     sync.Mutex
	BackgroundFile            string
	Text                      map[int]Content
	DirtyFrame                image.Image
	SelectedBox               int
	BaseIndex                 int
	Payment                   string
	Form                      ImageSlice
	Upload                    string
	Redirect                  string
	Data                      string
	SignalFullRedrawNeeded    func(*Session)
	SignalPartialRedrawNeeded func(*Session, int)
	SignalFocusLost           func(*Session, int)
	SignalShiftFocus          func(*Session)
	SignalFocusGot            func(*Session, int)
	SignalFocusChanged        func(*Session, int, int)
	SignalTextChange          func(*Session, int, string, string)
	SignalClicked             func(*Session, int)
	SignalUploaded            func(*Session, Upload)
	SignalRecalculate         func(*Session)
	SignalClosed              func(*Session)
}

type Upload struct {
	Type string
	Body []byte
}

func ResetSession(w http.ResponseWriter, r *http.Request) error {
	apiKey := r.URL.Query().Get("apikey")
	if apiKey != "" {
		session, ok := sessions[apiKey]
		if ok {
			session.Mutex.Lock()
		}
		sessions[apiKey] = &Session{ApiKey: apiKey, Mutex: sync.Mutex{}, SelectedBox: 0, BaseIndex: 0}
		if ok {
			session.Mutex.Unlock()
		}
	}
	return nil
}

func EnsureAPIKey(w http.ResponseWriter, r *http.Request) error {
	apiKey := r.URL.Query().Get("apikey")
	if apiKey == "" || len(apiKey) != len(GenerateUniqueKey()) {
		//management.QuantumGradeAuthorization()
		time.Sleep(15 * time.Millisecond)
		w.Header().Set("Location", r.URL.EscapedPath()+fmt.Sprintf("?apikey=%s", GenerateUniqueKey()))
		w.WriteHeader(http.StatusTemporaryRedirect)
		return fmt.Errorf("redirect")
	}
	return nil
}

func ServeRemoteForm(w http.ResponseWriter, r *http.Request, name string) {
	w.Header().Set("Cache-Control", "no-cache")
	if ResetSession(w, r) != nil {
		return
	}
	w.Header().Set("Cache-Control", "no-cache")
	raw := NoErrorBytes(os.ReadFile("./drawing/res/remote.html"))
	html := strings.ReplaceAll(string(raw), "remote.html", name+".html")
	html = strings.ReplaceAll(html, "<title>opensource.eper.io</title>", "<title>"+metadata.SiteName+"</title>")
	html = strings.ReplaceAll(html, "remote.png", name+".png")
	_, _ = w.Write([]byte(html))
}

func ServeRemoteFrame(w http.ResponseWriter, r *http.Request, formFunc func(session *Session)) {
	session := GetActivatedSession(w, r)
	if session != nil {
		w.Header().Set("Cache-Control", "no-cache")
		session.Mutex.Lock()
		Loaded.Lock()
		Loaded.Unlock()
		init := session.Form.Boxes == nil
		formFunc(session)
		if init {
			session.SignalFullRedrawNeeded(session)
		}
		ProcessInputs(w, r)
		session.Mutex.Unlock()
	}
}

func GetActivatedSession(w http.ResponseWriter, r *http.Request) *Session {
	err := EnsureAPIKey(w, r)
	if err != nil {
		return nil
	}
	return GetSession(w, r)
}

func GetSession(w http.ResponseWriter, r *http.Request) *Session {
	apiKey := r.URL.Query().Get("apikey")
	_, found := sessions[apiKey]
	if !found {
		if ResetSession(w, r) != nil {
			return nil
		}
	}
	return sessions[apiKey]
}

func RecalculateSession(apiKey string) {
	session, found := sessions[apiKey]
	if found {
		if session.SignalRecalculate != nil {
			session.SignalRecalculate(session)
		}
	}
}

func ProcessInputs(w http.ResponseWriter, r *http.Request) {
	session := GetSession(w, r)

	fullRedraw := true
	xs := r.URL.Query().Get("X")
	x, errX := strconv.ParseFloat(xs, 64)
	ys := r.URL.Query().Get("Y")
	y, errY := strconv.ParseFloat(ys, 64)

	if errX == nil && errY == nil {
		for i, b := range session.Form.Boxes {
			clicked := image.Point{X: int(x * float64(session.Form.Rect.Dx())), Y: int(y * float64(session.Form.Rect.Dy()))}
			if clicked.In(b) {
				ChangeFocus(session, i)
				break
			}
		}
		fullRedraw = false
	}

	t := r.URL.Query().Get("T")
	c := session.SelectedBox
	textChanged := false
	_, ok := session.Text[c]
	if ok && t != "" {
		textNew := session.Text[c].Text
		for t != "" {
			t = strings.ReplaceAll(t, "Space", " ")
			t = strings.ReplaceAll(t, "Plus", "+")
			if strings.HasPrefix(t, "Backspace") {
				before, after, ok := strings.Cut(textNew, "�")
				if ok && len(before) > 0 {
					textNew = before[0:len(before)-1] + "�" + after
				}
				textChanged = true
				t = t[len("Backspace"):]
			} else if strings.HasPrefix(t, "Delete") {
				before, after, ok := strings.Cut(textNew, "�")
				if ok && len(after) > 0 {
					textNew = before + "�" + after[1:]
				}
				textChanged = true
				t = t[len("Delete"):]
			} else if strings.HasPrefix(t, "Enter") {
				if strings.Contains(textNew, "\v") {
					begin, middle, _ := strings.Cut(strings.ReplaceAll(textNew, "�", ""), "\v")
					middle, end, _ := strings.Cut(middle, "\v")
					textNew = begin + "�" + end
				} else if len(strings.Split(textNew, "\n")) < session.Text[c].Lines {
					before, after, ok := strings.Cut(textNew, "�")
					if ok {
						textNew = before + "\n" + "�" + after
					} else {
						textNew = textNew + "\n" + "�"
					}
				} else {
					before, after, ok := strings.Cut(textNew, "�")
					if ok {
						textNew = before + after
					}
				}
				textChanged = true
				t = t[len("Enter"):]
			} else if strings.HasPrefix(t, "Help") ||
				strings.HasPrefix(t, "Insert") {
				if strings.HasPrefix(t, "Help") {
					t = t[len("Help"):]
				}
				if strings.HasPrefix(t, "Insert") {
					t = t[len("Insert"):]
				}
				textChanged = true
				begin, end, _ := strings.Cut(textNew, "�")
				middle, end, _ := strings.Cut(end, "\v")
				left, end, _ := strings.Cut(end, "\v")
				if end == "" {
					end = left
				}
				if len(end) > 0 {
					textNew = begin + middle + "\v" + "�" + end
				} else {
					textNew = "�" + begin + middle + end
				}
			} else if strings.HasPrefix(t, "Escape") {
				t = t[len("Escape"):]
				textChanged = true
				begin, end, _ := strings.Cut(textNew, "�")
				middle, end, _ := strings.Cut(end, "\v")
				if len(end) > 0 {
					textNew = begin + middle + "\v" + "�" + end
				} else {
					textNew = "�" + begin + middle + end
				}
			} else if strings.HasPrefix(t, "Tab") {
				t = t[len("Tab"):]
				textChanged = true
				session.SignalShiftFocus(session)
			} else if strings.HasPrefix(t, "ArrowDown") ||
				strings.HasPrefix(t, "End") ||
				strings.HasPrefix(t, "PageDown") {
				i := strings.IndexRune(textNew, []rune("�")[0])
				s := []rune(textNew)
				if i >= 0 && i < len(s)-2 {
					for {
						if s[i+1] == '\n' && strings.HasPrefix(t, "End") {
							break
						}
						s[i] = s[i+1]
						s[i+1] = []rune("�")[0]
						if i == len(s)-2 || s[i] == '\n' {
							break
						}
						i++
						if s[i-1] != ' ' && s[i+1] == ' ' && strings.HasPrefix(t, "PageDown") {
							break
						}
					}
					textNew = string(s)
				}
				textChanged = true
				if strings.HasPrefix(t, "ArrowDown") {
					t = t[len("ArrowDown"):]
				}
				if strings.HasPrefix(t, "PageDown") {
					t = t[len("PageDown"):]
				}
				if strings.HasPrefix(t, "End") {
					t = t[len("End"):]
				}
			} else if strings.HasPrefix(t, "ArrowUp") ||
				strings.HasPrefix(t, "Home") ||
				strings.HasPrefix(t, "PageUp") {
				i := strings.IndexRune(textNew, []rune("�")[0])
				if i > 1 {
					s := []rune(textNew)
					for {
						if s[i-1] == '\n' && strings.HasPrefix(t, "Home") {
							break
						}
						s[i] = s[i-1]
						s[i-1] = []rune("�")[0]
						if i <= 1 || s[i] == '\n' {
							break
						}
						i--
						if s[i+1] == ' ' && s[i-1] != ' ' && strings.HasPrefix(t, "PageUp") {
							break
						}
					}
					textNew = string(s)
				}
				textChanged = true
				if strings.HasPrefix(t, "ArrowUp") {
					t = t[len("ArrowUp"):]
				}
				if strings.HasPrefix(t, "PageUp") {
					t = t[len("PageUp"):]
				}
				if strings.HasPrefix(t, "Home") {
					t = t[len("Home"):]
				}
			} else if strings.HasPrefix(t, "ArrowLeft") {
				i := strings.IndexRune(textNew, []rune("�")[0])
				if i > 0 {
					s := []rune(textNew)
					s[i] = s[i-1]
					s[i-1] = []rune("�")[0]
					textNew = string(s)
				}
				textChanged = true
				t = t[len("ArrowLeft"):]
			} else if strings.HasPrefix(t, "ArrowRight") {
				i := strings.IndexRune(textNew, []rune("�")[0])
				if i == -1 {
					break
				}
				s := []rune(textNew)
				if i+1 < len(s) {
					s[i] = s[i+1]
					s[i+1] = []rune("�")[0]
					textNew = string(s)
				}
				textChanged = true
				t = t[len("ArrowRight"):]
			} else {
				if len(t) > 100 {
					fmt.Printf("invalid string %c", t[0])
				} else if len(t) >= 1 {
					if strings.HasPrefix(textNew, RevertAndReturn) {
						textNew = t + "�"
					} else {
						before, after, ok := strings.Cut(textNew, "�")
						if ok {
							textNew = before + t + "�" + after
						} else {
							textNew = textNew + t + "�"
						}
					}
					textChanged = true
				}
			}
			break
		}
		if textChanged && session.Text[c].Editable {
			replace := session.Text[c]
			changeFrom := replace.Text
			replace.Text = textNew
			session.Text[c] = replace
			fullRedraw = false
			session.SignalTextChange(session, c, changeFrom, textNew)
		}
	}

	q := r.URL.Query().Get("progressive")
	if q != "" {
		if session.DirtyFrame != nil {
			fullRedraw = false
		} else {
			w.WriteHeader(200)
			return
		}
	}

	if fullRedraw && session.Redirect == "" {
		session.SignalFullRedrawNeeded(session)
	}

	streamResponse(w, session)
}

func streamResponse(w http.ResponseWriter, session *Session) {
	if session.Upload != "" {
		w.Header().Set("Extension-Requested", session.Upload)
		session.Upload = ""
		w.WriteHeader(http.StatusPaymentRequired)
		return
	} else if session.SelectedBox == -1 {
		w.Header().Set("Cache-Control", "no-store")
		if session.Redirect != "" {
			w.Header().Set("Location", session.Redirect)
		}
		w.WriteHeader(http.StatusGone)
		return
	} else if session.Redirect != "" {
		w.Header().Set("Cache-Control", "no-store")
		if session.Redirect != "" {
			w.Header().Set("Location", session.Redirect)
		}
		session.Redirect = ""
		w.WriteHeader(http.StatusConflict)
		return
	}

	framePng := session.DirtyFrame
	if framePng != nil {
		session.DirtyFrame = nil
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-store")
		if framePng.Bounds().Min.X > 0 || framePng.Bounds().Min.Y > 0 {
			w.Header().Set("X", fmt.Sprintf("%d", framePng.Bounds().Min.X))
			w.Header().Set("Y", fmt.Sprintf("%d", framePng.Bounds().Min.Y))
			w.Header().Set("W", fmt.Sprintf("%d", framePng.Bounds().Dx()))
			w.Header().Set("H", fmt.Sprintf("%d", framePng.Bounds().Dy()))
		}
		_ = png.Encode(w, framePng)
	}
}

func RenderFullRedraw(session *Session) {
	for i := 0; i < len(session.Form.Boxes); i++ {
		input, ok := session.Text[i]
		if !ok {
			EraseBorder(ImageSlice{Rgb: session.Form.Rgb, Rect: session.Form.Boxes[i]}) //TODO
			continue
		}
		if input.Background.Rgb != nil {
			DrawImage(ImageSlice{Rgb: session.Form.Rgb, Rect: session.Form.Boxes[i]}, input.Background.Rgb.SubImage(input.Background.Rect))
			input.BackgroundColor = color.Transparent
		} else {
			background := session.Form.Rgb.At(session.Form.Boxes[i].Min.X-1, session.Form.Boxes[i].Min.Y-1)
			input.BackgroundColor = background
		}
		DrawTextWithFillWithErase(session, ImageSlice{Rgb: session.Form.Rgb, Rect: session.Form.Boxes[i]}, input)
	}
	session.DirtyFrame = session.Form.Rgb
}

func RenderSingleBoxChange(session *Session, changedBox int) {
	if changedBox < 0 {
		return
	}
	selectedBox := session.Form.Boxes[changedBox]
	formatted := session.Text[changedBox]
	RenderBox(session, formatted, selectedBox)
}

func RenderBox(session *Session, formatted Content, selectedBox image.Rectangle) {
	background := session.Form.Rgb.At(selectedBox.Min.X+1, selectedBox.Min.Y+1)
	formatted.BackgroundColor = background

	cache := SampleImage(session.Form, selectedBox)
	DrawTextWithFillWithErase(session, ImageSlice{Rgb: session.Form.Rgb, Rect: selectedBox}, formatted)
	changes := ChangedRectangle(cache.Rgb, session.Form.Rgb.SubImage(selectedBox))
	if !changes.Empty() {
		if session.DirtyFrame == nil {
			session.DirtyFrame = session.Form.Rgb.SubImage(changes)
		} else {
			changes = session.DirtyFrame.Bounds().Union(changes)
			session.DirtyFrame = session.Form.Rgb.SubImage(changes)
		}
	}
}
