package block

import (
	"fmt"
)

func init() {
	Register("stick", func() Block { return new(stickblk) })
}

type stickblk struct {
	x, y *float64
}

func (b *stickblk) InputNames() []string { return []string{"x", "y"} }
func (b *stickblk) SetInput(sel string, port Port) error {
	var p **float64
	switch sel {
	case "x":
		p = &b.x
	case "y":
		p = &b.y
	default:
		return fmt.Errorf("stick block has no input named '%s'", sel)
	}
	var ok bool
	*p, ok = port.(*float64)
	if !ok {
		return fmt.Errorf("stick block needs scalar input")
	}
	return nil
}

func (b *stickblk) OutputNames() []string { return []string{"x", "y"} }
func (b *stickblk) Output(sel string) (Port, error) {
	switch sel {
	case "x":
		return b.x, nil
	case "y":
		return b.y, nil
	}
	return nil, fmt.Errorf("stick block has no output named '%s'", sel)
}

func (b *stickblk) Tick() {}
