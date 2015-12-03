package nanovgo

import (
	"math"
)

// DegToRad converts degree to radian.
func DegToRad(deg float32) float32 {
	return deg / 180.0 * PI
}

// RadToDeg converts radian to degree.
func RadToDeg(rad float32) float32 {
	return rad / PI * 180.0
}

func signF(a float32) float32 {
	if a > 0.0 {
		return 1.0
	}
	return -1.0
}

func clampF(a, min, max float32) float32 {
	if a < min {
		return min
	}
	if a > max {
		return max
	}
	return a
}

func clampI(a, min, max int) int {
	if a < min {
		return min
	}
	if a > max {
		return max
	}
	return a
}

func hue(h, m1, m2 float32) float32 {
	if h < 0.0 {
		h++
	} else if h > 1 {
		h--
	}
	if h < 1.0/6.0 {
		return m1 + (m2-m1)*h*6.0
	} else if h < 3.0/6.0 {
		return m2
	} else if h < 4.0/6.0 {
		return m1 + (m2-m1)*(2.0/3.0-h)*6.0
	}
	return m1
}

func minF(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func maxF(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func maxI(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxFs(v float32, values ...float32) float32 {
	max := v
	for _, value := range values {
		if max < value {
			max = value
		}
	}
	return max
}

func minFs(v float32, values ...float32) float32 {
	min := v
	for _, value := range values {
		if min > value {
			min = value
		}
	}
	return min
}

func cross(dx0, dy0, dx1, dy1 float32) float32 {
	return dx1*dy0 - dx0*dy1
}

func absF(a float32) float32 {
	if a > 0.0 {
		return a
	}
	return -a
}

func sqrtF(a float32) float32 {
	return float32(math.Sqrt(float64(a)))
}
func atan2F(a, b float32) float32 {
	return float32(math.Atan2(float64(a), float64(b)))
}

func acosF(a float32) float32 {
	return float32(math.Acos(float64(a)))
}

func tanF(a float32) float32 {
	return float32(math.Tan(float64(a)))
}

func sinCosF(a float32) (float32, float32) {
	s, c := math.Sincos(float64(a))
	return float32(s), float32(c)
}

func ceilF(a float32) int {
	return int(math.Ceil(float64(a)))
}

func normalize(x, y float32) (float32, float32, float32) {
	d := float32(math.Sqrt(float64(x*x + y*y)))
	if d > 1e-6 {
		id := 1.0 / d
		x *= id
		y *= id
	}
	return d, x, y
}

func intersectRects(ax, ay, aw, ah, bx, by, bw, bh float32) [4]float32 {
	minX := maxF(ax, bx)
	minY := maxF(ay, by)
	maxX := minF(ax+aw, bx+bw)
	maxY := minF(ay+ah, by+bh)
	return [4]float32{
		minX,
		minY,
		maxF(0.0, maxX-minX),
		maxF(0.0, maxY-minY),
	}
}

func ptEquals(x1, y1, x2, y2, tol float32) bool {
	dx := x2 - x1
	dy := y2 - y1
	return dx*dx+dy*dy < tol*tol
}

func distPtSeg(x, y, px, py, qx, qy float32) float32 {
	pqx := qx - px
	pqy := qy - py
	dx := x - px
	dy := y - py
	d := pqx*pqx + pqy*pqy
	t := clampF(pqx*dx+pqy*dy, 0.0, 1.1)
	if d > 0 {
		t /= d
	}
	dx = px + t*pqx - x
	dy = py + t*pqy - y
	return dx*dx + dy*dy
}

func triArea2(ax, ay, bx, by, cx, cy float32) float32 {
	abX := bx - ax
	abY := by - ay
	acX := cx - ax
	acY := cy - ay
	return acX*abY - abX*acY
}

func polyArea(points []nvgPoint, npts int) float32 {
	var area float32
	a := &points[0]
	for i := 2; i < npts; i++ {
		b := &points[i-1]
		c := &points[i]
		area += triArea2(a.x, a.y, b.x, b.y, c.x, c.y)
	}
	return area * 0.5
}

func polyReverse(points []nvgPoint, npts int) {
	i := 0
	j := npts - 1
	for i < j {
		points[i], points[j] = points[j], points[i]
		i++
		j--
	}
}

func curveDivs(r, arc, tol float32) int {
	da := math.Acos(float64(r/(r+tol))) * 2.0
	return maxI(2, int(math.Ceil(float64(arc)/da)))
}

func chooseBevel(bevel bool, p0, p1 *nvgPoint, w float32) (x0, y0, x1, y1 float32) {
	if bevel {
		x0 = p1.x + p0.dy*w
		y0 = p1.y - p0.dx*w
		x1 = p1.x + p1.dy*w
		y1 = p1.y - p1.dx*w
	} else {
		x0 = p1.x + p1.dmx*w
		y0 = p1.y + p1.dmy*w
		x1 = p1.x + p1.dmx*w
		y1 = p1.y + p1.dmy*w
	}
	return
}

func roundJoin(dst []nvgVertex, index int, p0, p1 *nvgPoint, lw, rw, lu, ru float32, nCap int, fringe float32) int {
	dlx0 := p0.dy
	dly0 := -p0.dx
	dlx1 := p1.dy
	dly1 := -p1.dx
	isInnerBevel := p1.flags&nvgPrINNERBEVEL != 0
	if p1.flags&nvgPtLEFT != 0 {
		lx0, ly0, lx1, ly1 := chooseBevel(isInnerBevel, p0, p1, lw)
		a0 := atan2F(-dly0, -dlx0)
		a1 := atan2F(-dly1, -dlx1)
		if a1 > a0 {
			a1 -= PI * 2
		}
		(&dst[index]).set(lx0, ly0, lu, 1)
		(&dst[index+1]).set(p1.x-dlx0*rw, p1.y-dly0*rw, ru, 1)
		index += 2
		n := clampI(ceilF(((a0-a1)/PI)*float32(nCap)), 2, nCap)
		for i := 0; i < n; i++ {
			u := float32(i) / float32(n-1)
			a := a0 + u*(a1-a0)
			s, c := sinCosF(a)
			rx := p1.x + c*rw
			ry := p1.y + s*rw
			(&dst[index]).set(p1.x, p1.y, 0.5, 1)
			(&dst[index+1]).set(rx, ry, ru, 1)
			index += 2
		}
		(&dst[index]).set(lx1, ly1, lu, 1)
		(&dst[index+1]).set(p1.x-dlx1*rw, p1.y-dly1*rw, ru, 1)
		index += 2
	} else {
		rx0, ry0, rx1, ry1 := chooseBevel(isInnerBevel, p0, p1, -rw)
		a0 := atan2F(dly0, dlx0)
		a1 := atan2F(dly1, dlx1)
		if a1 < a0 {
			a1 += PI * 2
		}
		(&dst[index]).set(p1.x+dlx0*rw, p1.y+dly0*rw, lu, 1)
		(&dst[index+1]).set(rx0, ry0, ru, 1)
		index += 2
		n := clampI(ceilF(((a1-a0)/PI)*float32(nCap)), 2, nCap)
		for i := 0; i < n; i++ {
			u := float32(i) / float32(n-1)
			a := a0 + u*(a1-a0)
			s, c := sinCosF(a)
			lx := p1.x + c*lw
			ly := p1.y + s*lw
			(&dst[index]).set(lx, ly, lu, 1)
			(&dst[index+1]).set(p1.x, p1.y, 0.5, 1)
			index += 2
		}
		(&dst[index]).set(p1.x+dlx1*rw, p1.y+dly1*rw, lu, 1)
		(&dst[index+1]).set(rx1, ry1, ru, 1)
		index += 2
	}
	return index
}

func bevelJoin(dst []nvgVertex, index int, p0, p1 *nvgPoint, lw, rw, lu, ru, fringe float32) int {
	dlx0 := p0.dy
	dly0 := -p0.dx
	dlx1 := p1.dy
	dly1 := -p1.dx
	isInnerBevel := p1.flags&nvgPrINNERBEVEL != 0
	isBevel := p1.flags&nvgPtBEVEL != 0
	if p1.flags&nvgPtLEFT != 0 {
		lx0, ly0, lx1, ly1 := chooseBevel(isInnerBevel, p0, p1, lw)

		(&dst[index]).set(lx0, ly0, lu, 1)
		(&dst[index+1]).set(p1.x-dlx0*rw, p1.y-dly0*rw, ru, 1)
		index += 2

		if isBevel {
			(&dst[index]).set(lx0, ly0, lu, 1)
			(&dst[index+1]).set(p1.x-dlx0*rw, p1.y-dly0*rw, ru, 1)

			(&dst[index+2]).set(lx1, ly1, lu, 1)
			(&dst[index+3]).set(p1.x-dlx1*rw, p1.y-dly1*rw, ru, 1)

			index += 4
		} else {
			rx0 := p1.x - p1.dmx*rw
			ry0 := p1.y - p1.dmy*rw

			(&dst[index]).set(p1.x, p1.y, 0.5, 1)
			(&dst[index+1]).set(p1.x-dlx0*rw, p1.y-dly0*rw, ru, 1)

			(&dst[index+2]).set(rx0, ry0, ru, 1)
			(&dst[index+3]).set(rx0, ry0, ru, 1)

			(&dst[index+4]).set(p1.x, p1.y, 0.5, 1)
			(&dst[index+5]).set(p1.x-dlx1*rw, p1.y-dly1*rw, ru, 1)

			index += 6
		}
		(&dst[index]).set(lx1, ly1, lu, 1)
		(&dst[index+1]).set(p1.x-dlx1*rw, p1.y-dly1*rw, ru, 1)
		index += 2
	} else {
		rx0, ry0, rx1, ry1 := chooseBevel(isInnerBevel, p0, p1, -rw)

		(&dst[index]).set(p1.x+dlx0*lw, p1.y+dly0*lw, lu, 1)
		(&dst[index+1]).set(rx0, ry0, ru, 1)
		index += 2

		if isBevel {
			(&dst[index]).set(p1.x+dlx0*lw, p1.y+dly0*lw, lu, 1)
			(&dst[index+1]).set(rx0, ry0, ru, 1)

			(&dst[index+2]).set(p1.x+dlx1*rw, p1.y+dly1*rw, lu, 1)
			(&dst[index+3]).set(rx1, ry1, ru, 1)

			index += 4
		} else {
			lx0 := p1.x + p1.dmx*rw
			ly0 := p1.y + p1.dmy*rw

			(&dst[index]).set(p1.x+dlx0*lw, p1.y+dly0*lw, lu, 1)
			(&dst[index+1]).set(p1.x, p1.y, 0.5, 1)

			(&dst[index+2]).set(lx0, ly0, lu, 1)
			(&dst[index+3]).set(lx0, ly0, lu, 1)

			(&dst[index+4]).set(p1.x+dlx1*lw, p1.y+dly1*lw, lu, 1)
			(&dst[index+5]).set(p1.x, p1.y, 0.5, 1)

			index += 6
		}
		(&dst[index]).set(p1.x+dlx1*lw, p1.y+dly1*lw, lu, 1)
		(&dst[index+1]).set(rx1, ry1, ru, 1)
		index += 2
	}
	return index
}

func buttCapStart(dst []nvgVertex, index int, p *nvgPoint, dx, dy, w, d, aa float32) int {
	px := p.x - dx*d
	py := p.y - dy*d
	dlx := dy
	dly := -dx
	(&dst[index]).set(px+dlx*w-dx*aa, py+dly*w-dy*aa, 0, 0)
	(&dst[index+1]).set(px-dlx*w-dx*aa, py-dly*w-dy*aa, 1, 0)
	(&dst[index+2]).set(px+dlx*w, py+dly*w, 0, 1)
	(&dst[index+3]).set(px-dlx*w, py-dly*w, 1, 1)
	return index + 4
}

func buttCapEnd(dst []nvgVertex, index int, p *nvgPoint, dx, dy, w, d, aa float32) int {
	px := p.x + dx*d
	py := p.y + dy*d
	dlx := dy
	dly := -dx
	(&dst[index]).set(px+dlx*w, py+dly*w, 0, 1)
	(&dst[index+1]).set(px-dlx*w, py-dly*w, 1, 1)
	(&dst[index+2]).set(px+dlx*w+dx*aa, py+dly*w+dy*aa, 0, 0)
	(&dst[index+3]).set(px-dlx*w+dx*aa, py-dly*w-dy*aa, 1, 0)
	return index + 4
}

func roundCapStart(dst []nvgVertex, index int, p *nvgPoint, dx, dy, w float32, nCap int, aa float32) int {
	px := p.x
	py := p.y
	dlx := dy
	dly := -dx
	for i := 0; i < nCap; i++ {
		a := float32(i) / float32(nCap-1) * PI
		s, c := sinCosF(a)
		ax := c * w
		ay := s * w
		(&dst[index]).set(px-dlx*ax-dx*ay, py-dly*ax-dy*ay, 0, 1)
		(&dst[index+1]).set(px, py, 0.5, 1)
		index += 2
	}
	(&dst[index]).set(px+dlx*w, py+dly*w, 0, 1)
	(&dst[index+1]).set(px-dlx*w, py-dly*w, 1, 1)
	return index + 2
}

func roundCapEnd(dst []nvgVertex, index int, p *nvgPoint, dx, dy, w float32, nCap int, aa float32) int {
	px := p.x
	py := p.y
	dlx := dy
	dly := -dx
	(&dst[index]).set(px+dlx*w, py+dly*w, 0, 1)
	(&dst[index+1]).set(px-dlx*w, py-dly*w, 1, 1)
	index += 2
	for i := 0; i < nCap; i++ {
		a := float32(i) / float32(nCap-1) * PI
		s, c := sinCosF(a)
		ax := c * w
		ay := s * w
		(&dst[index]).set(px, py, 0.5, 1)
		(&dst[index+1]).set(px-dlx*ax+dx*ay, py-dly*ax+dy*ay, 0, 1)
		index += 2
	}
	return index
}

func nearestPow2(num int) int {
	var n uint
	uNum := uint(num)
	if uNum > 0 {
		n = uNum - 1
	} else {
		n = 0
	}
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return int(num)
}

func quantize(a, d float32) float32 {
	return float32(int(a/d+0.5)) * d
}
