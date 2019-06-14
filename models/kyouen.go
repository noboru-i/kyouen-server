package models

import (
	"math"
	"strings"
)

// KyouenStage hold a stage of kyouen.
type KyouenStage struct {
	size           int
	stonePointList []Point
}

// KyouenData hold kyouen result.
type KyouenData struct {
	points     []Point
	lineKyouen bool
	center     FloatPoint
	radius     float64
	line       Line
}

// NewKyouenStage create stage by string.
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

// NewKyouenDataWithLine create kyouen result of line.
func NewKyouenDataWithLine(p1 Point, p2 Point, p3 Point, p4 Point, aLine Line) *KyouenData {
	return newKyouenData(p1, p2, p3, p4, true, FloatPoint{}, 0.0, aLine)
}

// NewKyouenDataWithOval create kyouen result of oval.
func NewKyouenDataWithOval(p1 Point, p2 Point, p3 Point, p4 Point, aCenter FloatPoint, aRadius float64) *KyouenData {
	return newKyouenData(p1, p2, p3, p4, false, aCenter, aRadius, Line{})
}

func newKyouenData(p1 Point, p2 Point, p3 Point, p4 Point, aIsLine bool, aCenter FloatPoint, aRadius float64, aLine Line) *KyouenData {
	points := []Point{p1, p2, p3, p4}
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

// HasKyouen is checking stage has kyouen.
func (k KyouenStage) HasKyouen() *KyouenData {
	size := len(k.stonePointList)
	for i := 0; i < size-3; i++ {
		p1 := k.stonePointList[i]
		for j := i + 1; j < size-2; j++ {
			p2 := k.stonePointList[j]
			for l := j + 1; l < size-1; l++ {
				p3 := k.stonePointList[l]
				for m := l + 1; m < size; m++ {
					p4 := k.stonePointList[m]
					result := isKyouen(p1, p2, p3, p4)
					if result != nil {
						return result
					}
				}
			}
		}
	}
	return nil
}

func isKyouen(p1 Point, p2 Point, p3 Point, p4 Point) *KyouenData {
	fp1 := *NewFloatPoint(p1)
	fp2 := *NewFloatPoint(p2)
	fp3 := *NewFloatPoint(p3)
	fp4 := *NewFloatPoint(p4)

	// p1,p2の垂直二等分線を求める
	l12 := *GetMidperpendicular(fp1, fp2)
	// p2,p3の垂直二等分線を求める
	l23 := *GetMidperpendicular(fp2, fp3)

	// 交点を求める
	intersection123 := GetIntersection(l12, l23)
	if intersection123 == nil {
		// p1,p2,p3が直線上に存在する場合
		l34 := *GetMidperpendicular(fp3, fp4)
		// p2,p3,p4が直線上に存在する場合
		intersection234 := GetIntersection(l23, l34)
		if intersection234 == nil {
			return NewKyouenDataWithLine(p1, p2, p3, p4, *NewLine(fp1, fp2))
		}
	} else {
		dist1 := fp1.Distance(*intersection123)
		dist2 := fp4.Distance(*intersection123)
		if math.Abs(dist1-dist2) < 0.0000001 {
			return NewKyouenDataWithOval(p1, p2, p3, p4, *intersection123, dist1)
		}
	}
	return nil
}
