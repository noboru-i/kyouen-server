package models

// Line hold line data.
type Line struct {
	p1 FloatPoint
	p2 FloatPoint
	a  float64
	b  float64
	c  float64
}

// NewLine create line data from 2 points.
func NewLine(p1 FloatPoint, p2 FloatPoint) *Line {
	a := p1.y - p2.y
	b := p2.x - p1.x
	c := p1.x*p2.y - p2.x*p1.y

	return &Line{p1: p1, p2: p2, a: a, b: b, c: c}
}

// GetMidperpendicular returns line of midperpendicular by 2 points.
func GetMidperpendicular(p1 FloatPoint, p2 FloatPoint) *Line {
	midpoint := *GetMidpoint(p1, p2)
	dif := p1.Difference(p2)
	gradient := FloatPoint{x: dif.y, y: -1 * dif.x}

	return NewLine(midpoint, *midpoint.Add(gradient))
}

// GetIntersection returns point of intersection by 2 lines.
func GetIntersection(l1 Line, l2 Line) *FloatPoint {
	f1 := float64(l1.p2.x - l1.p1.x)
	g1 := float64(l1.p2.y - l1.p1.y)
	f2 := float64(l2.p2.x - l2.p1.x)
	g2 := float64(l2.p2.y - l2.p1.y)

	det := f2*g1 - f1*g2
	if det == 0.0 {
		return nil
	}

	dx := l2.p1.x - l1.p1.x
	dy := l2.p1.y - l1.p1.y
	t1 := (f2*dy - g2*dx) / det

	return &FloatPoint{x: l1.p1.x + f1*t1, y: l1.p1.y + g1*t1}
}
