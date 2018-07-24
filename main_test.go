package main

import (
	"strings"
	"testing"
)

func TestTrimToken(t *testing.T) {
	str := "XYZ-------------456"
	trim := trimToken(str)
	correct := "XYZ--...--456"
	if strings.EqualFold(correct, trim) {
		t.Error("Expect", correct, "got", trim)
	}
}

func TestDistance(t *testing.T) {

	if DistanceForStrings([]rune(""), []rune("")) != 0 {
		t.Error("TestDistance fail 1")
	}

	if DistanceForStrings([]rune("1"), []rune("1")) != 0 {
		t.Error("TestDistance fail 2")
	}

	if DistanceForStrings([]rune("1111"), []rune("1111")) != 0 {
		t.Error("TestDistance fail 2")
	}

	if DistanceForStrings([]rune("ёлка"), []rune("елка")) != 1 {
		t.Error("TestDistance fail 3")
	}

	if DistanceForStrings([]rune("ёлко"), []rune("елка")) == 1 {
		t.Error("TestDistance fail 4")
	}
}
