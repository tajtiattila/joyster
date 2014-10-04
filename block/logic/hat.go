package logic

import (
	"github.com/tajtiattila/joyster/block"
)

func init() {
	// hatelem takes a hat and produces button outputs for them
	block.Register("hatelem", func() block.Block { return new(hatelem) })

	// makehat makes a hat out of distinct flags/buttons
	block.Register("makehat", func() block.Block { return new(makehat) })

	block.RegisterHatFunc("hatadd", func(a, b int) int { return a | b })
	block.RegisterHatFunc("hatsub", func(a, b int) int { return a & ^b })
	block.RegisterHatFunc("hatxor", func(a, b int) int { return a ^ b })
}

type hatelem struct {
	i          *int
	n, s, e, w bool
}

func (h *hatelem) Tick() {
	val := *h.i
	h.n = (val & block.HatNorth) != 0
	h.s = (val & block.HatSouth) != 0
	h.w = (val & block.HatWest) != 0
	h.e = (val & block.HatEast) != 0
}

func (h *hatelem) Validate() error       { return block.CheckInput(&h.i) }
func (h *hatelem) Input() block.InputMap { return block.SingleInput("hatelem", &h.i) }

func (h *hatelem) Output() block.OutputMap {
	return block.MapOutput("hatelem",
		pt("n", &h.n),
		pt("s", &h.s),
		pt("e", &h.e),
		pt("w", &h.w),
	)
}

type makehat struct {
	n, s, e, w *bool
	o          int
}

func (h *makehat) Tick() {
	val := 0
	if *h.n {
		val |= block.HatNorth
	}
	if *h.s {
		val |= block.HatSouth
	}
	if *h.w {
		val |= block.HatWest
	}
	if *h.e {
		val |= block.HatEast
	}
	h.o = val
}

func (h *makehat) Validate() error         { return block.CheckInputs("makehat", &h.n, &h.s, &h.w, &h.e) }
func (h *makehat) Output() block.OutputMap { return block.SingleOutput("makehat", &h.o) }
func (h *makehat) Input() block.InputMap {
	return block.MapInput("makehat",
		pt("n", &h.n),
		pt("s", &h.s),
		pt("w", &h.w),
		pt("e", &h.e),
	)
}
