package logic

import (
	"github.com/tajtiattila/joyster/block"
)

func init() {
	block.RegisterParam("combo", func(p block.Param) (block.Block, error) {
		return newCombohat(p), nil
	})
}

func newCombohat(p block.Param) block.Block {
	return &combohat{
		taptick: uint(p.Arg("TapDelay") * p.TickFreq()),
		pushlen: uint(p.Arg("KeepPushed") * p.TickFreq()),
		phase:   phasewaitrelease,
	}
}

type combohat struct {
	taptick uint
	pushlen uint

	i *int
	o [5]combohatout

	phase combophase
	sel   int
	timer uint
}

func (h *combohat) Tick() {
	h.phase = h.phase(h)
	for i := range h.o {
		o := &h.o[i]
		o.t--
		if o.t == 0 {
			o.v = block.HatCentre
		}
	}
}

func (h *combohat) Validate() error       { return block.CheckInput(&h.i) }
func (h *combohat) Input() block.InputMap { return block.SingleInput("combohat", &h.i) }

func (h *combohat) Output() block.OutputMap {
	return block.MapOutput("combohat",
		pt("", &h.o[0].v),
		pt("n", &h.o[1].v),
		pt("s", &h.o[2].v),
		pt("e", &h.o[3].v),
		pt("w", &h.o[4].v),
	)
}

type combohatout struct {
	v int
	t uint
}

type combophase func(h *combohat) combophase

// phasewaitrelease is activated after a valid or incorrect combo (multiple directions)
// It waits until the hat is released, then returns phasestart.
func phasewaitrelease(h *combohat) combophase {
	if *h.i == 0 {
		return phasestart
	}
	return phasewaitrelease
}

// phasestarts is the initial state when no user input is in effect.
// It waits for a hat button press, and enters phasepushing or phasewaitrelease.
func phasestart(h *combohat) combophase {
	val := *h.i
	if val != 0 {
		if hatidxmap[val&15] == 0 {
			return phasewaitrelease
		}
		h.sel, h.timer = val, h.taptick
		return phasepushing
	}
	return phasestart
}

// phasepushing is active during the first hat activation.
func phasepushing(h *combohat) combophase {
	h.timer--
	if h.timer == 0 {
		h.o[0].v, h.o[0].t = h.sel, h.pushlen
		return phasewaitrelease
	}
	val := *h.i
	if val != h.sel {
		if val == 0 {
			return phasewaitnext
		}
		return phasewaitrelease
	}
	return phasepushing
}

// phasewaitnext waits for the second hat press
func phasewaitnext(h *combohat) combophase {
	h.timer--
	if h.timer == 0 {
		h.o[0].v, h.o[0].t = h.sel, h.pushlen
		return phasewaitrelease
	}
	val := *h.i
	if val != 0 {
		if hatidxmap[val&15] != 0 {
			o := &h.o[hatidxmap[h.sel&15]]
			o.v, o.t = val, h.pushlen
		}
		return phasewaitrelease
	}
	return phasewaitnext
}

var hatidxmap [16]int

func init() {
	hatidxmap[block.HatNorth] = 1
	hatidxmap[block.HatSouth] = 2
	hatidxmap[block.HatWest] = 3
	hatidxmap[block.HatEast] = 4
}
