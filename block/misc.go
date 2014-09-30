package block

import (
	"fmt"
)

func RegisterScalarFunc(name string, fn func(Param) (func(float64) float64, error)) {
	RegisterParam(name, func(p Param) (Block, error) {
		f, err := fn(p)
		if err != nil {
			return nil, err
		}
		return &scalarfnblk{
			typ: name,
			f:   f,
		}, nil
	})
}

type scalarfnblk struct {
	typ string
	i   *float64
	o   float64
	f   func(float64) float64
}

func (b *scalarfnblk) Tick() { b.o = b.f(*b.i) }

func (b *scalarfnblk) InputNames() []string { return []string{""} }
func (b *scalarfnblk) SetInput(sel string, port Port) error {
	if sel != "" {
		return fmt.Errorf("'%s' has no input named '%s'", b.typ, sel)
	}
	var ok bool
	b.i, ok = port.(*float64)
	if !ok {
		return fmt.Errorf("'%s' needs scalar input, not %s", b.typ, PortString(port))
	}
	return nil
}

func (b *scalarfnblk) OutputNames() []string { return []string{""} }
func (b *scalarfnblk) Output(sel string) (Port, error) {
	if sel != "" {
		return nil, fmt.Errorf("'%s' has no output named '%s'", b.typ, sel)
	}
	return &b.o, nil
}

func RegisterBoolFunc(name string, fn func(Param) (func(bool) bool, error)) {
	RegisterParam(name, func(p Param) (Block, error) {
		f, err := fn(p)
		if err != nil {
			return nil, err
		}
		return &boolfnblk{
			typ: name,
			f:   f,
		}, nil
	})
}

type boolfnblk struct {
	typ string
	i   *bool
	o   bool
	f   func(bool) bool
}

func (b *boolfnblk) Tick() { b.o = b.f(*b.i) }

func (b *boolfnblk) InputNames() []string { return []string{""} }
func (b *boolfnblk) SetInput(sel string, port Port) error {
	if sel != "" {
		return fmt.Errorf("'%s' has no input named '%s'", b.typ, sel)
	}
	var ok bool
	b.i, ok = port.(*bool)
	if !ok {
		return fmt.Errorf("'%s' needs bool input, not %s", b.typ, PortString(port))
	}
	return nil
}

func (b *boolfnblk) OutputNames() []string { return []string{""} }
func (b *boolfnblk) Output(sel string) (Port, error) {
	if sel != "" {
		return nil, fmt.Errorf("'%s' has no output named '%s'", b.typ, sel)
	}
	return &b.o, nil
}

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
		return fmt.Errorf("stick block needs scalar input, not %s", PortString(port))
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
