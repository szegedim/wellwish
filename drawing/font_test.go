package drawing

import (
	"testing"
	"time"
)

func TestConvertAll(t *testing.T) {
	// Loading a font should not take more than 20 seconds at startup
	start := time.Now()
	LoadFont(indexes, "res/courier.png", "/tmp/")
	LoadSpace()
	LoadFont("�", "res/cursorwide.png", "/tmp/")
	if time.Now().Sub(start).Seconds() > 20 {
		t.Error("timeout 20s")
	}
}
