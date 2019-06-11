package models

import (
	"testing"
)

func TestGetMidpoint(t *testing.T) {
	p1 := FloatPoint{x: 1, y: 1}
	p2 := FloatPoint{x: 3, y: 4}
	actual := *GetMidpoint(p1, p2)
	expect := FloatPoint{x: 2, y: 2.5}
	if actual != expect {
		t.Errorf("mid point of 1,1 and 3,4 is %v. but actual is %v", expect, actual)
	}
}

func TestAdd(t *testing.T) {
	p1 := FloatPoint{x: 1, y: 4}
	p2 := FloatPoint{x: 3, y: 2}
	actual := *(p1.Add(p2))
	expect := FloatPoint{x: 4, y: 6}
	if actual != expect {
		t.Errorf("sum of 1,4 and 3,2 is %v. but actual is %v", expect, actual)
	}
}

func TestDifference(t *testing.T) {
	p1 := FloatPoint{x: 1, y: 4}
	p2 := FloatPoint{x: 3, y: 2}
	actual := *(p1.Difference(p2))
	expect := FloatPoint{x: -2, y: 2}
	if actual != expect {
		t.Errorf("difference of 1,4 and 3,2 is %v. but actual is %v", expect, actual)
	}
}

func TestDistance(t *testing.T) {
	p1 := FloatPoint{x: 1, y: 1}
	p2 := FloatPoint{x: 4, y: 5}
	actual := p1.Distance(p2)
	expect := 5.0
	if actual != expect {
		t.Errorf("distance of 1,1 and 4,5 is %v. but actual is %v", expect, actual)
	}
}
