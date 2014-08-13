package main

type ButtonHandler interface {
	Handle(bool) bool // update with actual input, return output
}

type SimpleButton struct{}

func (*SimpleButton) Handle(i bool) bool {
	return i
}

type DelayedButton struct {
	needtick int

	ntick int
	state bool
}

func (b *DelayedButton) Handle(i bool) bool {
	if i != b.state {
		b.ntick++
		if b.needtick <= b.ntick {
			b.ntick = 0
			b.state = i
		}
	}
	return b.state
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

func (b *MultiButton) Handle(i bool) bool {
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
		return true
	}
	return false
}
