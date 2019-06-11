package models

import (
	"testing"
)

func TestNewLine(t *testing.T) {
	p1 := FloatPoint{x: 1, y: 1}
	p2 := FloatPoint{x: 3, y: 4}
	actual := *NewLine(p1, p2)
	expect := Line{p1: p1, p2: p2, a: -3, b: 2, c: 1}
	if actual != expect {
		t.Errorf("line from %v and %v must be %v. actual is %v", p1, p2, expect, actual)
	}
}

func TestGetMidperpendicular(t *testing.T) {
	p1 := FloatPoint{x: 1, y: 1}
	p2 := FloatPoint{x: 3, y: 4}
	actual := *GetMidperpendicular(p1, p2)
	expect := Line{p1: FloatPoint{x: 2, y: 2.5}, p2: FloatPoint{x: -1, y: 4.5}, a: -2, b: -3, c: 11.5}
	if actual != expect {
		t.Errorf("midperpendicular from %v and %v must be %v. actual is %v", p1, p2, expect, actual)
	}
}

func TestGetIntersection(t *testing.T) {
	// TODO not implemented
}
