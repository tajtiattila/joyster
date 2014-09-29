package block

import (
	"fmt"
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
	typ    string
	i1, i2 *float64
	o      float64
	tick   func(a, b float64) float64
}

func (b *mathopblk) Tick() {
	b.o = b.tick(*b.i1, *b.i2)
}

func (b *mathopblk) InputNames() []string { return []string{"1", "2"} }
func (b *mathopblk) SetInput(sel string, port Port) error {
	if sel != "1" && sel != "2" {
		return fmt.Errorf("'%s' has inputs '1' and '2', but not '%s'", b.typ, sel)
	}
	i, ok := port.(*float64)
	if !ok {
		return fmt.Errorf("'%s' block needs scalar, not %s", b.typ, PortString(port))
	}
	if sel == "1" {
		b.i1 = i
	} else {
		b.i2 = i
	}
	return nil
}
func (b *mathopblk) OutputNames() []string { return []string{""} }
func (b *mathopblk) Output(sel string) (Port, error) {
	if sel != "" {
		return nil, fmt.Errorf("'%s' block has no named output '%s'", b.typ, sel)
	}
	return &b.o, nil
}
