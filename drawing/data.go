package drawing

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

import (
	"image"
	"image/color"
	"sync"
)

const RevertAndReturn = "\r"
const InitialHint = RevertAndReturn

var Loaded sync.Mutex

var Transparent = color.RGBA64{R: 0xFFFF, G: 0xFFFF, B: 0xFFFF, A: 0x0}
var White = color.RGBA64{R: 0xFFFF, G: 0xFFFF, B: 0xFFFF, A: 0xFFFF}
var Black = color.RGBA64{R: 0, G: 0, B: 0, A: 0xFFFF}

var sessions = map[string]*Session{}

type quickbmp struct {
	bounds image.Rectangle
	bits   []bool
}

var fontCache = map[rune]quickbmp{}

// Copy this into a fontCache png file.
// Make sure that you leave enough padding
var indexes = `Qabcdefghijklmnopqrstuvwxyz
ABCDEFGHIJKLMNOPQRSTUVWXYZ
Q()[]{}<>/\
Q0123456789
Q?!@#$%^&*-
Q,';:=+~
Q._
QẞäÄ
`
