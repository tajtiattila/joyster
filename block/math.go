package block

import (
	"math"
)

func RegisterMathBlock(name string, fn func(a, b float64) float64) {
	Register(name, func() Block {
		return &mathopblk{typ: name, tick: fn}
	})
}

func init() {
	RegisterMathBlock("add", func(a, b float64) float64 { return a + b })
	RegisterMathBlock("sub", func(a, b float64) float64 { return a - b })
	RegisterMathBlock("mul", func(a, b float64) float64 { return a * b })
	RegisterMathBlock("div", func(a, b float64) float64 { return a / b })
	RegisterMathBlock("mod", func(a, b float64) float64 { return math.Mod(a, b) })
	RegisterMathBlock("pow", func(a, b float64) float64 { return math.Pow(a, b) })
	RegisterMathBlock("min", func(a, b float64) float64 { return math.Min(a, b) })
	RegisterMathBlock("max", func(a, b float64) float64 { return math.Max(a, b) })
	RegisterMathBlock("absmin", func(a, b float64) float64 {
		if math.Abs(a) < math.Abs(b) {
			return a
		}
		return b
	})
	RegisterMathBlock("absmax", func(a, b float64) float64 {
		if math.Abs(a) > math.Abs(b) {
			return a
		}
		return b
	})
}

type mathopblk struct {
	typ  string
	vi   []*float64
	o    float64
	tick func(a, b float64) float64
}

func (b *mathopblk) Tick() {
	b.o = b.tick(*b.vi[0], *b.vi[1])
	for _, p := range b.vi[2:] {
		b.o = b.tick(b.o, *p)
	}
}

func (b *mathopblk) Input() InputMap   { return VarArgInput(b.typ, &b.vi) }
func (b *mathopblk) Output() OutputMap { return SingleOutput(b.typ, &b.o) }
func (b *mathopblk) Validate() error   { return VarArgCheck(b.typ, &b.vi, 2) }
