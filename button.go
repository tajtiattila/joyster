package main

import (
	"github.com/tajtiattila/vjoy"
)

type ButtonHandler interface {
	Handle(*vjoy.Device, *ButtonConfig, bool) // update output based on input
}

type SimpleButton struct{}

func (*SimpleButton) Handle(d *vjoy.Device, c *ButtonConfig, i bool) {
	b := d.Button(c.Output - 1)
	b.Set(i)
}

type DelayedButton struct {
	needtick int

	ntick int
	state bool
}

func (b *DelayedButton) Handle(d *vjoy.Device, c *ButtonConfig, i bool) {
	if i != b.state {
		b.ntick++
		if b.needtick <= b.ntick {
			b.ntick = 0
			b.state = i
		}
	}
	db := d.Button(c.Output - 1)
	db.Set(b.state)
}

type MultiButton struct {
	taptick uint // max time between taps
	needtap uint // taps needed for output
	pushlen uint // length of output press

	state bool // last input state
	tapc  uint // counter: decreased over time, increased on tap
	ntap  uint // number of taps so far
	push  uint // counter: decreased over time, nonzero is pressed
}

func (b *MultiButton) Handle(d *vjoy.Device, c *ButtonConfig, i bool) {
	bn := d.Button(c.Output - 1)
	bd := d.Button(c.Double - 1)
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

	if b.push != 0 {
		b.push--
		bn.Set(false)
		bd.Set(true)
	} else {
		bn.Set(i)
		bd.Set(false)
	}
}
