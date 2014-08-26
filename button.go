package main

type ButtonDevice interface {
	SetButton(idx uint, value bool)
}

type ButtonHandler interface {
	Handle(ButtonDevice, *ButtonConfig, bool) // update output based on input
}

type SimpleButton struct{}

func (*SimpleButton) Handle(d ButtonDevice, c *ButtonConfig, i bool) {
	d.SetButton(c.Output-1, i)
}

type DelayedButton struct {
	needtick int

	ntick int
	state bool
}

func (b *DelayedButton) Handle(d ButtonDevice, c *ButtonConfig, i bool) {
	if i != b.state {
		b.ntick++
		if b.needtick <= b.ntick {
			b.ntick = 0
			b.state = i
		}
	}
	d.SetButton(c.Output-1, b.state)
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

func (b *MultiButton) Handle(d ButtonDevice, c *ButtonConfig, i bool) {
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
	d.SetButton(c.Output-1, bn)
	d.SetButton(c.Double-1, bd)
}
