package logic

import (
	"fmt"
	"github.com/tajtiattila/joyster/block"
)

type buttonShifter struct {
	i, shift        *bool
	normal, shifted bool
}

func (b *buttonShifter) Setup(c block.Config)   {}
func (b *buttonShifter) Inputs() block.InputMap { return block.InputMap{"": &b.i, "shift": &b.shift} }
func (b *buttonShifter) Outputs() block.OutputMap {
	return block.OutputMap{"normal": &b.normal, "shifted": &b.shifted}
}
func (b *buttonShifter) Tick() {
	if *b.shift {
		b.normal = false
		b.shifted = *b.i
	} else {
		b.normal = *b.i
		b.shifted = false
	}
}

type contMultiButton struct {
	taptick uint // max time between taps
	pushlen uint // length of output press
	needtap uint // taps needed for output

	state bool // last input state
	tapc  uint // counter: decreased over time, increased on tap
	ntap  uint // number of taps so far
	push  uint // counter: decreased over time, nonzero is pressed

	i      *bool
	o, dbl bool
}

func (b *contMultiButton) Inputs() block.InputMap { return block.InputMap{"": &b.i} }
func (b *contMultiButton) Outputs() block.OutputMap {
	return block.OutputMap{"": &b.o, "double": &b.dbl}
}

func (b *contMultiButton) Setup(c block.Config) {
	b.taptick = uint(c.Float64("tapdelay") * c.TickFreq())
	b.pushlen = uint(c.Float64("keeppushed") * c.TickFreq())
	b.needtap = uint(c.OptInt("needtap", 2))
}

func (b *contMultiButton) Tick() {
	i := *b.i
	if i != b.state {
		b.state = i
		if i {
			b.tapc += b.taptick
			b.ntap++
			if b.ntap == b.needtap {
				b.push = b.pushlen
				b.tapc = 0
				b.ntap = 0
			}
		}
	}

	if b.tapc != 0 {
		b.tapc--
	} else {
		b.ntap = 0
	}

	var bn, bd bool
	if b.push != 0 {
		b.push--
		bd = true
	} else {
		bn = i
	}
	b.o, b.dbl = bn, bd
}

// tapMultiButton keeps holding one of its outputs depending on how many times it were tapped.
// a double tap does not result in the single tap output firing
type tapMultiButton struct {
	taptick uint // max time between taps
	pushlen uint // length of output press

	state bool // last input state
	tapc  uint // counter: decreased over time, increased on tap
	ntap  int  // number of taps so far

	i *bool
	v []*tapMultiOut
}

type tapMultiOut struct {
	o    bool
	hold uint
}

func (b *tapMultiButton) Setup(c block.Config) {
	b.taptick = uint(c.Float64("tapdelay") * c.TickFreq())
	b.pushlen = uint(c.Float64("keeppushed") * c.TickFreq())
	nmaxtaps := c.OptInt("numtaps", 2)
	b.v = make([]*tapMultiOut, nmaxtaps)
	for i := range b.v {
		b.v[i] = new(tapMultiOut)
	}
}

func (b *tapMultiButton) Inputs() block.InputMap { return block.InputMap{"": &b.i} }
func (b *tapMultiButton) Outputs() block.OutputMap {
	m := make(block.OutputMap)
	for i, t := range b.v {
		m[fmt.Sprint(i+1)] = &t.o
	}
	return m
}

func (b *tapMultiButton) Tick() {
	if *b.i != b.state {
		b.state = *b.i
		if b.state {
			b.tapc = b.taptick
			b.ntap++
		}
	}

	if b.tapc != 0 {
		b.tapc--
		if b.tapc == 0 {
			idx := b.ntap - 1
			if idx < len(b.v) {
				t := b.v[idx]
				t.o, t.hold = true, b.pushlen
			}
		}
		b.ntap = 0
	}

	for _, t := range b.v {
		if t.hold != 0 {
			t.hold--
			if t.hold == 0 {
				t.o = false
			}
		}
	}
}

func init() {
	block.Register("shift", func() block.Block { return new(buttonShifter) })
	block.Register("multi", func() block.Block { return new(contMultiButton) })
	block.Register("tapmulti", func() block.Block { return new(tapMultiButton) })
}
