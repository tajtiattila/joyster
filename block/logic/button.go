package logic

import (
	"fmt"
	"github.com/tajtiattila/joyster/block"
)

func init() {
	block.RegisterParam("multibutton", func(p block.Param) (block.Block, error) {
		return newMultiButton(p), nil
	})
	block.RegisterParam("doublebutton", func(p block.Param) (block.Block, error) {
		return newDoubleButton(p), nil
	})
}

func newDoubleButton(p block.Param) block.Block {
	b := new(doubleButton)
	b.taptick = uint(p.Arg("TapDelay") * p.TickFreq())
	b.pushlen = uint(p.Arg("KeepPushed") * p.TickFreq())
	return b
}

type doubleButton struct {
	taptick uint // max time between taps
	pushlen uint // length of output press

	state bool // last input state
	tapc  uint // counter: decreased over time, increased on tap
	ntap  int  // number of taps so far
	push  uint // counter: decreased over time, nonzero is pressed

	i      *bool
	o, dbl bool
}

func (b *doubleButton) Tick() {
	i := *b.i
	if i != b.state {
		b.state = i
		if i {
			b.tapc += b.taptick
			b.ntap++
			if b.ntap == 2 {
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

	if b.push != 0 {
		b.push--
		b.o, b.dbl = false, true
	} else {
		b.o, b.dbl = i, false
	}
}

func (b *doubleButton) Input() block.InputMap { return block.SingleInput("doublebutton", &b.i) }
func (b *doubleButton) Output() block.OutputMap {
	return block.MapOutput("doublebutton", pt("", &b.o), pt("double", &b.dbl))
}
func (b *doubleButton) Validate() error { return block.CheckInputs("multibutton", &b.i) }

// multiButton keeps holding one of its outputs depending on how many times it were tapped.
// a double tap does not result in the single tap output firing
func newMultiButton(p block.Param) block.Block {
	b := new(multiButton)
	var nmaxtaps int
	if p == block.ProtoParam {
		nmaxtaps = 16
	} else {
		nmaxtaps = int(p.Arg("NumTaps"))
	}
	b.taptick = uint(p.Arg("TapDelay") * p.TickFreq())
	b.pushlen = uint(p.Arg("KeepPushed") * p.TickFreq())
	b.v = make([]*tapMultiOut, nmaxtaps)
	for i := range b.v {
		b.v[i] = new(tapMultiOut)
	}
	return b
}

type multiButton struct {
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

func (b *multiButton) Input() block.InputMap { return block.SingleInput("multibutton", &b.i) }
func (b *multiButton) Output() block.OutputMap {
	d := make([]block.MapDecl, len(b.v))
	for i, t := range b.v {
		d[i] = pt(fmt.Sprint(i+1), &t.o)
	}
	return block.MapOutput("multibutton", d...)
}
func (b *multiButton) Validate() error { return block.CheckInputs("multibutton", &b.i) }

func (b *multiButton) Tick() {
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
				t.o = true
				t.hold = b.pushlen
			}
			b.ntap = 0
		}
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

func pt(n string, v interface{}) block.MapDecl { return block.MapDecl{n, v} }
