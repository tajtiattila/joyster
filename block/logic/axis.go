package logic

import (
	"fmt"
	"github.com/tajtiattila/joyster/block"
	"math"
)

// add value to input
type OffsetBlock struct {
	Value float64
}

func (b *OffsetBlock) Setup(c block.Config) bool { b, ok := c.(OffsetBlock); return ok }
func (b *OffsetBlock) Tick(v float64) float64    { return v + b.Offset }

// zero input under abs. value, reduce bigger
type DeadzoneBlock struct {
	Deadzone float64
}

func (b *DeadzoneBlock) Tick(v float64) float64 {
	var s float64
	if v < 0 {
		v, s = -v, -1
	} else {
		s = 1
	}
	v -= b.Deadzone
	if v < 0 {
		return 0
	}
	return v * s
}

// multiply input by factor
type MultiplyBlock struct {
	Factor float64
}

func (b *MultiplyBlock) Tick(v float64) float64 {
	return v * b.Factor
}

// axis sensitivivy curve (factor: 0 - linear, positive: nonlinear)
type CurvatureBlock struct {
	pow float64
}

func (b *CurvatureBlock) Setup(c map[string]interface{}) bool {
	f, ok := c["factor"].(float64)
	b.pow = math.Pow(2, f)
	return ok
}

func (b *CurvatureBlock) Tick(v float64) float64 {
	s := float64(1)
	if v < 0 {
		s, v = -1, -v
	}
	return s * math.Pow(v, pow)
}

// truncate input above abs. value
type TruncateBlock struct {
	treshold float64
}

func (b *TruncateBlock) Setup(c Config) {
	b.treshold = c.Float64("factor")
}

func (b *TruncateBlock) Tick(v float64) float64 {
	switch {
	case v < -b.treshold:
		return -b.treshold
	case b.treshold < v:
		return b.treshold
	}
	return v
}

// set maximum input change to value/second
type DampenBlock struct {
	speed float64
	pos   float64
}

func (b *DampenBlock) Setup(c Config) {
	b.speed = c.Float64("value") * c.TickTime()
}

func (b *DampenBlock) Tick(v float64) {
	switch {
	case b.pos+speed < v:
		b.pos += speed
	case v < b.pos-speed:
		b.pos -= speed
	default:
		b.pos = v
	}
	return b.pos
}

// smooth inputs over time (seconds)
type SmoothBlock struct {
	b0, b1 float64
	posv   []int64
	n      int
	sum    int64
}

func (b *SmoothBlock) Setup(c Config) {
	nsamples := math.Floor(c.Float64("factor") * c.TickFreq())
	b.m0 = math.Pow(2, 63) / (nsamples * 100)
	b.m1 = 1 / (m0 * nsamples)
	b.posv = make([]int64, int(nsamples))
	b.n, b.sum = 0, 0
}

func (b *SmoothBlock) Tick(v float64) float64 {
	iv := int64(v * b.m0)
	b.sum -= posv[b.n]
	b.posv[b.n], b.n = iv, (b.n+1)%len(b.posv)
	b.sum += iv
	return float64(b.sum) * b.m1
}

// use input as delta, change values by speed/second
type IncrementalBlock struct {
	speed, rebound float64
	quickcenter    bool
}

func (b *IncrementalBlock) Setup(c Config) {
	b.speed = c.Float64("speed") * c.TickTime()
	b.rebound = c.OptFloat64("speed", 0.0) * c.TickTime()
	b.quickcenter = OptBool("quickcenter", false)
}

func (b *IncrementalBlock) Tick(v float64) float64 {
	if math.Abs(v) < 1e-3 {
		switch {
		case b.pos < -b.rebound:
			b.pos += b.rebound
		case b.rebound < pos:
			b.pos -= b.rebound
		default:
			b.pos = 0
		}
	} else {
		if quickcenter && b.pos*v < 0 {
			b.pos = 0
		} else {
			b.pos += v * b.speed
			switch {
			case b.pos < -1:
				b.pos = -1
			case 1 < b.pos:
				b.pos = 1
			}
		}
	}
	return b.pos
}

type axisLogic interface {
	Setup(c Config)
	Tick(float64) float64
}

type axisBlock struct {
	logic axisLogic
	i     *block.ScalarValue
	o     block.ScalarValue
}

func newAxisBlock(l axisLogic) Block {
	return &axisBlock{
		logic: l,
		i:     new(block.ScalarValue),
	}
}

func (b *axisBlock) Inputs() InputMap   { return InputMap{"": &b.i} }
func (b *axisBlock) Outputs() OutputMap { return OutputMap{"": &b.o} }

func (b *axisBlock) Setup(c Config) {
	b.logic.Setup(c)
}

func (b *axisBlock) Tick() {
	b.o = block.ScalarValue(b.logic.Tick(float64(*b.i)))
}

func init() {
	block.Register("offset", func() Block { return newAxisBlock(new(OffsetBlock)) })
	block.Register("deadzone", func() Block { return newAxisBlock(new(DeadzoneBlock)) })
	block.Register("multiply", func() Block { return newAxisBlock(new(MultiplyBlock)) })
	block.Register("curvature", func() Block { return newAxisBlock(new(CurvatureBlock)) })
	block.Register("truncate", func() Block { return newAxisBlock(new(TruncateBlock)) })
	block.Register("dampen", func() Block { return newAxisBlock(new(DampenBlock)) })
	block.Register("smooth", func() Block { return newAxisBlock(new(SmoothBlock)) })
	block.Register("incremental", func() Block { return newAxisBlock(new(IncrementalBlock)) })
}
