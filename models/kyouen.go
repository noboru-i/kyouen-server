package models

import (
	"math"
	"strings"
)

type KyouenStage struct {
	size           int
	stonePointList []Point
}

type KyouenData struct {
	points     []FloatPoint
	lineKyouen bool
	center     FloatPoint
	radius     float64
	line       Line
}

func NewKyouenStage(size int, stage string) *KyouenStage {
	points := []Point{}
	for i, s := range stage {
		if string(s) == "1" {
			x := i % size
			y := i / size
			p := Point{x: x, y: y}
			points = append(points, p)
		}
	}
	return &KyouenStage{size: size, stonePointList: points}
}

func NewKyouenDataWithLine(p1 FloatPoint, p2 FloatPoint, p3 FloatPoint, p4 FloatPoint, aLine Line) *KyouenData {
	return NewKyouenData(p1, p2, p3, p4, true, FloatPoint{}, 0.0, aLine)
}

func NewKyouenDataWithOval(p1 FloatPoint, p2 FloatPoint, p3 FloatPoint, p4 FloatPoint, aCenter FloatPoint, aRadius float64) *KyouenData {
	return NewKyouenData(p1, p2, p3, p4, false, aCenter, aRadius, Line{})
}

func NewKyouenData(p1 FloatPoint, p2 FloatPoint, p3 FloatPoint, p4 FloatPoint, aIsLine bool, aCenter FloatPoint, aRadius float64, aLine Line) *KyouenData {
	points := []FloatPoint{p1, p2, p3, p4}
	return &KyouenData{points, aIsLine, aCenter, aRadius, aLine}
}

func (k KyouenStage) toString() string {
	result := make([]string, k.size*k.size)
	for i := 0; i < k.size*k.size; i++ {
		result[i] = "0"
	}
	for _, point := range k.stonePointList {
		index := point.x + point.y*k.size
		result[index] = "1"
	}
	return strings.Join(result, "")
}

func IsKyouen(p1 FloatPoint, p2 FloatPoint, p3 FloatPoint, p4 FloatPoint) *KyouenData {
	// p1,p2の垂直二等分線を求める
	l12 := *GetMidperpendicular(p1, p2)
	// p2,p3の垂直二等分線を求める
	l23 := *GetMidperpendicular(p2, p3)

	// 交点を求める
	intersection123 := GetIntersection(l12, l23)
	if intersection123 == nil {
		// p1,p2,p3が直線上に存在する場合
		l34 := *GetMidperpendicular(p3, p4)
		// p2,p3,p4が直線上に存在する場合
		intersection234 := GetIntersection(l23, l34)
		if intersection234 == nil {
			return NewKyouenDataWithLine(p1, p2, p3, p4, *NewLine(p1, p2))
		}
	} else {
		dist1 := p1.Distance(*intersection123)
		dist2 := p4.Distance(*intersection123)
		if math.Abs(dist1-dist2) < 0.0000001 {
			return NewKyouenDataWithOval(p1, p2, p3, p4, *intersection123, dist1)
		}
	}
	return nil
}
