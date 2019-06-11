package models

import (
	"reflect"
	"testing"
)

func TestNewKyouenStage(t *testing.T) {
	stage := "000000001000000000000000000000000000"
	s := NewKyouenStage(6, stage)
	if len(s.stonePointList) != 1 {
		t.Errorf("%q must have 1 stone. len = %q", stage, len(s.stonePointList))
	}
	if s.stonePointList[0].x != 2 || s.stonePointList[0].y != 1 {
		t.Errorf("%q stone point must be {x:2, y:1}. stone = %+v", stage, s.stonePointList[0])
	}
}

func TestToString(t *testing.T) {
	s := KyouenStage{size: 6, stonePointList: []Point{Point{x: 2, y: 1}}}
	expect := "000000001000000000000000000000000000"
	if s.toString() != expect {
		t.Errorf("{x:2, y:1} must be %q. string = %q", expect, s)
	}
}

func TestHasKyouen(t *testing.T) {
	points := []Point{
		Point{x: 1, y: 1},
		Point{x: 2, y: 2},
		Point{x: 3, y: 2},
		Point{x: 2, y: 3},
		Point{x: 3, y: 3},
	}
	actual := HasKyouen(points)
	expectPoints := []Point{
		Point{x: 2, y: 2},
		Point{x: 3, y: 2},
		Point{x: 2, y: 3},
		Point{x: 3, y: 3},
	}
	if !reflect.DeepEqual(actual.points, expectPoints) {
		t.Errorf("2,2 , 3,2 , 2,3 , 3,3 must be oval kyouen. actual = %v", actual)
	}
}

func TestIsKyouenWithNotKyouen(t *testing.T) {
	p1 := Point{x: 2, y: 2}
	p2 := Point{x: 3, y: 2}
	p3 := Point{x: 2, y: 3}
	p4 := Point{x: 3, y: 4}
	actual := IsKyouen(p1, p2, p3, p4)
	if actual != nil {
		t.Errorf("2,3 , 3,2 , 2,3 , 3,4 must be oval kyouen. actual = %v", actual)
	}
}

func TestIsKyouenWithOval(t *testing.T) {
	p1 := Point{x: 2, y: 2}
	p2 := Point{x: 3, y: 2}
	p3 := Point{x: 2, y: 3}
	p4 := Point{x: 3, y: 3}
	actual := *IsKyouen(p1, p2, p3, p4)
	if actual.lineKyouen == true {
		t.Errorf("2,3 , 3,2 , 2,3 , 3,3 must be oval kyouen. actual = %v", actual)
	}
	if actual.center != (FloatPoint{x: 2.5, y: 2.5}) {
		t.Errorf("2,3 , 3,2 , 2,3 , 3,3 must have center 2.5,2.5. actual = %v", actual)
	}
}

func TestIsKyouenWithLine(t *testing.T) {
	p1 := Point{x: 0, y: 2}
	p2 := Point{x: 2, y: 2}
	p3 := Point{x: 4, y: 2}
	p4 := Point{x: 5, y: 2}
	actual := *IsKyouen(p1, p2, p3, p4)
	if actual.lineKyouen == false {
		t.Errorf("0,2 , 2,2 , 4,2 , 5,2 must be line kyouen. actual = %v", actual)
	}
	if actual.line.a != 0 || actual.line.b*2+actual.line.c != 0 {
		t.Errorf("0,2 , 2,2 , 4,2 , 5,2 must be line y = 2. actual = %v", actual)
	}
}
