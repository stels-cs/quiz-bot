package main

import (
	"testing"
	"strings"
)

func TestTrimToken(t *testing.T) {
	str := "XYZ-------------456"
	trim := trimToken(str)
	correct := "XYZ--...--456"
	if strings.EqualFold(correct, trim) {
		t.Error("Expect",correct,"got",trim)
	}
}