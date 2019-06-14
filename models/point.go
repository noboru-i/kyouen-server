package models

import "math"

// Point have 2 int values.
type Point struct {
	x int
	y int
}

// FloatPoint have 2 float64 values.
type FloatPoint struct {
	x float64
	y float64
}

// NewFloatPoint returns FloatPoint by Point.
func NewFloatPoint(p Point) *FloatPoint {
	return &FloatPoint{x: float64(p.x), y: float64(p.y)}
}

// GetMidpoint returns midpoint by 2 points.
func GetMidpoint(p1 FloatPoint, p2 FloatPoint) *FloatPoint {
	x := (p1.x + p2.x) / 2
	y := (p1.y + p2.y) / 2
	return &FloatPoint{x: x, y: y}
}

// Add point.
func (p FloatPoint) Add(p2 FloatPoint) *FloatPoint {
	return &FloatPoint{p.x + p2.x, p.y + p2.y}
}

// Difference point.
func (p FloatPoint) Difference(p2 FloatPoint) *FloatPoint {
	return &FloatPoint{p.x - p2.x, p.y - p2.y}
}

// Distance point.
func (p FloatPoint) Distance(p2 FloatPoint) float64 {
	diff := p.Difference(p2)
	return math.Hypot(diff.x, diff.y)
}
