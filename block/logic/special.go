package logic

import (
	"github.com/tajtiattila/joyster/block"
	"math"
)

func init() {
	block.RegisterParam("headlook", func(p block.Param) (block.Block, error) {
		return newHeadlook(p), nil
	})
	block.RegisterParam("pedals", func(p block.Param) (block.Block, error) {
		return newPedals(p), nil
	})
}

type viewaccumulatelogic struct {
	movepertick    float64
	jumppertick    float64
	acapertick     float64
	autocenterdist float64

	enable *bool
	xi, yi *float64
	x, y   float64
	s      float64
}

func newHeadlook(p block.Param) *viewaccumulatelogic {
	dt := p.TickTime()
	l := new(viewaccumulatelogic)
	l.acapertick = p.Arg("AutoCenterAccel") * dt
	l.autocenterdist = p.Arg("AutoCenterDist")
	l.movepertick = p.Arg("MovePerSec") * dt
	l.jumppertick = p.Arg("JumpToCenterAccel") * dt
	if l.acapertick <= 0.0 {
		l.acapertick = 1
	}
	if l.jumppertick <= 0.0 {
		l.jumppertick = 1
	}
	return l
}

func (l *viewaccumulatelogic) Input() block.InputMap {
	return block.MapInput("headlook",
		pt("enable", &l.enable),
		pt("x", &l.xi),
		pt("y", &l.yi),
	)
}

func (l *viewaccumulatelogic) Output() block.OutputMap {
	return block.MapOutput("headlook",
		pt("x", &l.x),
		pt("y", &l.y),
	)
}

func (l *viewaccumulatelogic) Validate() error {
	return block.CheckInputs("headlook", &l.enable, &l.xi, &l.yi)
}

func (l *viewaccumulatelogic) Tick() {
	if *l.enable {
		xv, yv := float64(*l.xi), float64(*l.yi)
		if tiny(xv) && tiny(yv) {
			l.centeraccel(l.acapertick, l.autocenterdist)
		} else {
			l.x += xv * l.movepertick
			l.y += yv * l.movepertick
			if l.x < -1 {
				l.x = -1
			}
			if 1 < l.x {
				l.x = 1
			}
			if l.y < -1 {
				l.y = -1
			}
			if 1 < l.y {
				l.y = 1
			}
			l.s = 0
		}
	} else {
		l.centeraccel(l.jumppertick, 1e6)
	}
}

func (l *viewaccumulatelogic) centeraccel(a, limit float64) {
	// d=a/2*t²
	d := math.Sqrt(l.x*l.x + l.y*l.y)
	switch {
	case d < 1e-6:
		l.x, l.y, l.s = 0, 0, 0
	case d < limit:
		t := math.Sqrt(2 * d / a)
		maxs := a * t
		l.s += a
		if l.s > maxs {
			l.s = maxs
		}
		m := 1 - l.s/d
		if m < 0 {
			m = 0
		}
		l.x *= m
		l.y *= m
	}
}

type pedals struct {
	axisThreshold  float64
	breakThreshold float64
	exp            float64
	m              float64

	left, right *float64

	brk bool
	pos float64
}

func newPedals(p block.Param) block.Block {
	at := p.Arg("AxisThreshold")
	bt := p.Arg("BreakThreshold")
	exp := p.Arg("Exp")
	m := 1 / (1 - at)
	return &pedals{
		axisThreshold:  at,
		breakThreshold: bt,
		exp:            exp,
		m:              m,
	}
}

func (t *pedals) Input() block.InputMap {
	return block.MapInput("pedals",
		pt("left", &t.left),
		pt("right", &t.right),
	)
}

func (t *pedals) Output() block.OutputMap {
	return block.MapOutput("pedals",
		pt("", &t.pos),
		pt("break", &t.brk),
	)
}

func (t *pedals) Validate() error { return block.CheckInputs("pedals", &t.left, &t.right) }

func (t *pedals) Tick() {
	lv, rv := *t.left, *t.right
	lx, ly := lv > t.breakThreshold, rv > t.breakThreshold
	if lx == ly {
		t.brk = lx
		t.pos = 0
	} else {
		t.brk = false
		t.pos = t.triggervalue(rv) - t.triggervalue(lv)
	}
}

func (t *pedals) triggervalue(v float64) float64 {
	vv := v - t.axisThreshold
	if vv <= 0 {
		return 0
	}
	return math.Pow(vv*t.m, t.exp)
}

func tiny(v float64) bool {
	const rounderr = 1e-4
	return -rounderr < v && v < rounderr
}
