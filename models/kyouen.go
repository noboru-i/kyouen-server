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
	points     []Point
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

func NewKyouenDataWithLine(p1 Point, p2 Point, p3 Point, p4 Point, aLine Line) *KyouenData {
	return NewKyouenData(p1, p2, p3, p4, true, FloatPoint{}, 0.0, aLine)
}

func NewKyouenDataWithOval(p1 Point, p2 Point, p3 Point, p4 Point, aCenter FloatPoint, aRadius float64) *KyouenData {
	return NewKyouenData(p1, p2, p3, p4, false, aCenter, aRadius, Line{})
}

func NewKyouenData(p1 Point, p2 Point, p3 Point, p4 Point, aIsLine bool, aCenter FloatPoint, aRadius float64, aLine Line) *KyouenData {
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

func HasKyouen(points []Point) *KyouenData {
	size := len(points)
	for i := 0; i < size-3; i++ {
		p1 := points[i]
		for j := i + 1; j < size-2; j++ {
			p2 := points[j]
			for k := j + 1; k < size-1; k++ {
				p3 := points[k]
				for l := k + 1; l < size; l++ {
					p4 := points[l]
					result := IsKyouen(p1, p2, p3, p4)
					if result != nil {
						return result
					}
				}
			}
		}
	}
	return nil
}

func IsKyouen(p1 Point, p2 Point, p3 Point, p4 Point) *KyouenData {
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
