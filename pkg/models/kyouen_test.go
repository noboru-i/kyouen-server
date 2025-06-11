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

func TestNewRotatedKyouenStage(t *testing.T) {
	stage := "000000010000001100001100000000001000"
	s := *NewKyouenStage(6, stage)
	actual := NewRotatedKyouenStage(s)
	expect := "000000000010101100001100000000000000"
	if actual.ToString() != expect {
		t.Errorf("when %v rotate, it becomes %v. but %v.", stage, expect, actual.ToString())
	}
}

func TestNewMirroredKyouenStage(t *testing.T) {
	stage := "000000010000001100001100000000001000"
	s := *NewKyouenStage(6, stage)
	actual := NewMirroredKyouenStage(s)
	expect := "000000000010001100001100000000000100"
	if actual.ToString() != expect {
		t.Errorf("when %v mirror, it becomes %v. but %v.", stage, expect, actual.ToString())
	}
}

func TestToString(t *testing.T) {
	s := KyouenStage{size: 6, stonePointList: []Point{Point{x: 2, y: 1}}}
	expect := "000000001000000000000000000000000000"
	if s.ToString() != expect {
		t.Errorf("{x:2, y:1} must be %q. string = %q", expect, s)
	}
}

func TestStoneCount(t *testing.T) {
	stage := "001000000000100000000100000000000000"
	s := NewKyouenStage(6, stage)
	expect := 3
	if s.StoneCount() != expect {
		t.Errorf("%q has %v stones. current = %v", stage, expect, s.StoneCount())
	}
}

func TestHasKyouen(t *testing.T) {
	stage := "000000010000001100001100000000001000"
	s := NewKyouenStage(6, stage)
	actual := s.HasKyouen()
	expectPoints := []Point{
		Point{x: 2, y: 2},
		Point{x: 3, y: 2},
		Point{x: 2, y: 3},
		Point{x: 3, y: 3},
	}
	if !reflect.DeepEqual(actual.points, expectPoints) {
		t.Errorf("%s must be oval kyouen. actual = %v", stage, actual)
	}
}

func TestHasKyouenWithNoKyouen(t *testing.T) {
	stage := "000000010000000100001100000000001000"
	s := NewKyouenStage(6, stage)
	actual := s.HasKyouen()
	if actual != nil {
		t.Errorf("%s must not be kyouen. actual = %v", stage, actual)
	}
}

func TestIsKyouenWithNotKyouen(t *testing.T) {
	p1 := Point{x: 2, y: 2}
	p2 := Point{x: 3, y: 2}
	p3 := Point{x: 2, y: 3}
	p4 := Point{x: 3, y: 4}
	actual := isKyouen(p1, p2, p3, p4)
	if actual != nil {
		t.Errorf("2,3 , 3,2 , 2,3 , 3,4 must be oval kyouen. actual = %v", actual)
	}
}

func TestIsKyouenWithOval(t *testing.T) {
	p1 := Point{x: 2, y: 2}
	p2 := Point{x: 3, y: 2}
	p3 := Point{x: 2, y: 3}
	p4 := Point{x: 3, y: 3}
	actual := *isKyouen(p1, p2, p3, p4)
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
	actual := *isKyouen(p1, p2, p3, p4)
	if actual.lineKyouen == false {
		t.Errorf("0,2 , 2,2 , 4,2 , 5,2 must be line kyouen. actual = %v", actual)
	}
	if actual.line.a != 0 || actual.line.b*2+actual.line.c != 0 {
		t.Errorf("0,2 , 2,2 , 4,2 , 5,2 must be line y = 2. actual = %v", actual)
	}
}

func TestSomeKyouen(t *testing.T) {
	// all stages are kyouen
	stageList := []string{
		"000000010000001100001100000000001000",
		"000000000000000100010010001100000000",
		"000000001000010000000100010010001000",
		"001000001000000010010000010100000000",
		"000000001011010000000010001000000010",
		"000100000000101011010000000000000000",
		"000000001010000000010010000000001010",
		"001000000001010000010010000001000000",
		"000000001000010000000010000100001000",
		"000100000010010000000100000010010000",
	}

	for _, stageStr := range stageList {
		s := NewKyouenStage(6, stageStr)
		actual := s.HasKyouen()
		if actual == nil {
			t.Errorf("%s must be kyouen. actual = %v", stageStr, actual)
		}

		// if remove a point that contains kyouen, stage is not kyouen
		newPoints := []Point{}
		for _, point := range s.stonePointList {
			if actual.points[0] != point {
				newPoints = append(newPoints, point)
			}
		}
		s.stonePointList = newPoints
		newActual := s.HasKyouen()
		if newActual != nil {
			t.Errorf("%s must not be kyouen. actual = %v", s.ToString(), newActual)
		}
	}
}
