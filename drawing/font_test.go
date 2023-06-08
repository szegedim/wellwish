package drawing

import (
	"testing"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func TestConvertAll(t *testing.T) {
	// Loading a font should not take more than 20 seconds at startup
	start := time.Now()
	LoadFont(indexes, "res/defaultfont.png", "/tmp/")
	LoadSpace()
	LoadFont("�", "res/cursorwide.png", "/tmp/")
	if time.Now().Sub(start).Seconds() > 20 {
		t.Error("timeout 20s")
	}
}
