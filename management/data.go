package management

import (
	"io"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// TODO need a way to reset/renew this
var administrationKey = ""
var SiteActivationKey = ""

var CheckpointFunc func(m string, w io.Writer, r io.Reader)
