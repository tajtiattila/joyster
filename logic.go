package main

import (
	"math"
)

type thumbStick struct {
	xv float64
	xs float64
	yv float64
	ys float64
}

func (t *thumbStick) set(xi, yi int16) {
	t.xv, t.xs = float64FromInt16(xi)
	t.yv, t.ys = float64FromInt16(yi)
}

func (t *thumbStick) circularize() {
	if t.xv*t.xv+t.yv*t.yv < 1e-3 {
		return
	}

	var u float64
	if t.xv > t.yv {
		u = t.yv / t.xv // yu/xu = yv/xv
	} else {
		u = t.xv / t.yv // xu/yu = xv/yv
	}
	m := math.Sqrt(1 + u*u)
	t.xv *= m
	t.yv *= m
}

func float64FromInt16(v int16) (abs float64, sign float64) {
	if v < 0 {
		return float64(-int(v)) / 0x8000, -1
	} else {
		return float64(v) / 0x7fff, 1
	}
	return 0, 0 // not reached
}

type viewaccumulatelogic struct {
	x, y float64
	s    float64
}

func (v *viewaccumulatelogic) update(c *Config, xi, yi float32) (xo, yo float32) {
	if c.HeadLook != nil {
		xv, yv := float64(xi), float64(yi)
		if tiny(xv) && tiny(yv) {
			v.centeraccel(c.HeadLook.acapertick, c.HeadLook.AutoCenterDist)
		} else {
			v.x += xv * c.HeadLook.movepertick
			v.y += yv * c.HeadLook.movepertick
			if v.x < -1 {
				v.x = -1
			}
			if 1 < v.x {
				v.x = 1
			}
			if v.y < -1 {
				v.y = -1
			}
			if 1 < v.y {
				v.y = 1
			}
		}
	}
	return float32(v.x), float32(v.y)
}

func (v *viewaccumulatelogic) jumpToOrigin(c *Config) (xo, yo float32) {
	if c.HeadLook != nil {
		v.centeraccel(c.HeadLook.jumppertick, 1e6)
	}
	return float32(v.x), float32(v.y)
}

func (v *viewaccumulatelogic) centeraccel(a, limit float64) {
	// d=a/2*t²
	d := math.Sqrt(v.x*v.x + v.y*v.y)
	switch {
	case d < 1e-6:
		v.x, v.y, v.s = 0, 0, 0
	case d < limit:
		t := math.Sqrt(2 * d / a)
		maxs := a * t
		v.s += a
		if v.s > maxs {
			v.s = maxs
		}
		m := 1 - v.s/d
		if m < 0 {
			m = 0
		}
		v.x *= m
		v.y *= m
	}
}

func tiny(v float64) bool {
	return math.Abs(v) < 1e-3
}

func viewauto(pv *float64, limit, move float64) {
	v := *pv
	var vabs, vsign float64
	switch {
	case v <= -limit:
		vabs, vsign = -v, -1
	case v >= limit:
		vabs, vsign = v, 1
	default:
		*pv = 0
		return
	}
	if vabs <= limit {
		if move < vabs {
			*pv = v - vsign*move
		} else {
			*pv = 0
		}
	}
}

func triggermap(v, thr uint16, pow float64) float32 {
	if v <= thr {
		return 0
	}
	vv := float64(v-thr) / float64(255-thr)
	return float32(math.Pow(vv, pow))
}

func axismap(c *AxisConfig, vabs float64, sign float64) float64 {
	if vabs <= c.Min {
		return 0
	}
	if c.Max <= vabs {
		return sign
	}
	vabs = (vabs - c.Min) / (c.Max - c.Min)
	return sign * math.Pow(vabs, c.Pow)
}