package block

import (
	"fmt"
)

func init() {
	Register("not", func() Block { return new(notblk) })
	Register("and", func() Block { return new(andblk) })
	Register("or", func() Block { return new(orblk) })
	Register("xor", func() Block { return new(xorblk) })
	Register("if", func() Block { return new(ifblk) })
}

type boolblk struct {
	o bool
}

func (b *boolblk) OutputNames() []string { return []string{""} }
func (b *boolblk) Output(sel string) (Port, error) {
	if sel != "" {
		return nil, fmt.Errorf("logic block has no named output '%s'", sel)
	}
	return &b.o, nil
}

type notblk struct {
	boolblk
	i *bool
}

func (b *notblk) Tick() { b.o = !*b.i }

func (b *notblk) InputNames() []string { return []string{""} }
func (b *notblk) SetInput(sel string, port Port) error {
	if sel != "" {
		return fmt.Errorf("logic block has no named input '%s'", sel)
	}
	i, ok := port.(*bool)
	if !ok {
		return fmt.Errorf("logic block needs bool, not %s", PortString(port))
	}
	b.i = i
	return nil
}

type mboolblk struct {
	boolblk
	i1, i2 *bool
}

func (b *mboolblk) InputNames() []string { return []string{"1", "2"} }
func (b *mboolblk) SetInput(sel string, port Port) error {
	if sel != "1" && sel != "2" {
		return fmt.Errorf("logic block has inputs '1' and '2', not '%s'", sel)
	}
	i, ok := port.(*bool)
	if !ok {
		return fmt.Errorf("logic block needs bool, not %s", PortString(port))
	}
	if sel == "1" {
		b.i1 = i
	} else {
		b.i2 = i
	}
	return nil
}

type andblk struct{ mboolblk }

func (b *andblk) Tick() { b.o = *b.i1 && *b.i2 }

type orblk struct{ mboolblk }

func (b *orblk) Tick() { b.o = *b.i1 || *b.i2 }

type xorblk struct{ mboolblk }

func (b *xorblk) Tick() { b.o = *b.i1 != *b.i2 }

type ifblk struct {
	cond             *bool
	valthen, valelse Port

	out  Port
	tick func()
}

func (b *ifblk) OutputNames() []string { return []string{""} }
func (b *ifblk) Output(sel string) (Port, error) {
	if sel != "" {
		return nil, fmt.Errorf("if block has no named output '%s'", sel)
	}
	if b.out == nil {
		return nil, fmt.Errorf("unitialized if block")
	}
	return &b.out, nil
}

func (b *ifblk) Tick() {
	b.tick()
}

func (b *ifblk) InputNames() []string { return []string{"cond", "then", "else"} }
func (b *ifblk) SetInput(sel string, port Port) error {
	switch sel {
	case "cond":
		var ok bool
		b.cond, ok = port.(*bool)
		if !ok {
			return fmt.Errorf("if block needs bool condition, not %s", PortString(port))
		}
	case "then":
		b.valthen = port
	case "else":
		b.valelse = port
	default:
		return fmt.Errorf("if block has no input named '%s'", sel)
	}
	if b.valthen != nil && b.valelse != nil {
		if !matchport(b.valthen, b.valelse) {
			return fmt.Errorf("then and else must have the same type", sel)
		}
		switch th := b.valthen.(type) {
		case *bool:
			el := b.valelse.(*bool)
			o, ok := b.out.(*bool)
			if !ok || o == nil {
				o = new(bool)
				b.out = o
			}
			b.tick = func() {
				if *b.cond {
					*o = *th
				} else {
					*o = *el
				}
			}
		case *float64:
			el := b.valelse.(*float64)
			o, ok := b.out.(*float64)
			if !ok || o == nil {
				o = new(float64)
				b.out = o
			}
			b.tick = func() {
				if *b.cond {
					*o = *th
				} else {
					*o = *el
				}
			}
		case *int:
			el := b.valelse.(*int)
			o, ok := b.out.(*int)
			if !ok || o == nil {
				o = new(int)
				b.out = o
			}
			b.tick = func() {
				if *b.cond {
					*o = *th
				} else {
					*o = *el
				}
			}
		}
	}
	return nil
}

func matchport(a, b Port) bool {
	var match bool
	switch a.(type) {
	case *float64:
		_, match = b.(*float64)
	case *bool:
		_, match = b.(*bool)
	case *int:
		_, match = b.(*int)
	}
	return match
}
