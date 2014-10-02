package logic

import (
	"fmt"
	"github.com/tajtiattila/joyster/block"
)

func init() {
	block.RegisterParam("multibutton", func(p block.Param) (block.Block, error) {
		return newMultiButton(p), nil
	})
}

// multiButton keeps holding one of its outputs depending on how many times it were tapped.
// a double tap does not result in the single tap output firing
func newMultiButton(p block.Param) block.Block {
	b := new(multiButton)
	nmaxtaps := int(p.Arg("NumTaps"))
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

	state  bool // last input state
	tapc   uint // counter: decreased over time, increased on tap
	ntap   int  // number of taps so far
	ntapon int

	i *bool
	o bool
	v []*tapMultiOut
}

type tapMultiOut struct {
	o    bool
	hold uint
}

func (b *multiButton) Input() block.InputMap { return block.SingleInput("multibutton", &b.i) }
func (b *multiButton) Output() block.OutputMap {
	m := map[string]interface{}{"": &b.o}
	for i, t := range b.v {
		m[fmt.Sprint(i+1)] = &t.o
	}
	return block.MapOutput("multibutton", m)
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
				if !t.o {
					b.ntapon++
					t.o = true
				}
				t.hold = b.pushlen
			}
		}
		b.ntap = 0
	}

	for _, t := range b.v {
		if t.hold != 0 {
			t.hold--
			if t.hold == 0 {
				if t.o {
					b.ntapon--
					t.o = false
				}
			}
		}
	}
	b.o = b.ntapon == 0 && *b.i
}
