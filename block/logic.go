package block

import (
	"fmt"
)

func RegisterLogicFunc(name string, fn func(a, b bool) bool) {
	Register(name, func() Block {
		return &logicopblk{boolblk: boolblk{typ: name}, tick: fn}
	})
}

func init() {
	Register("not", func() Block { return &notblk{boolblk: boolblk{typ: "not"}} })
	RegisterLogicFunc("and", func(a, b bool) bool { return a && b })
	RegisterLogicFunc("or", func(a, b bool) bool { return a || b })
	RegisterLogicFunc("xor", func(a, b bool) bool { return a != b })
	Register("if", func() Block { return new(ifblk) })
}

type boolblk struct {
	typ string
	o   bool
}

type notblk struct {
	boolblk
	i *bool
}

func (b *notblk) Tick() { b.o = !*b.i }

func (b *boolblk) OutputNames() []string { return []string{""} }
func (b *boolblk) Output(sel string) (Port, error) {
	if sel != "" {
		return nil, fmt.Errorf("'%s' block has no named output '%s'", b.typ, sel)
	}
	return &b.o, nil
}

func (b *notblk) InputNames() []string { return []string{""} }
func (b *notblk) SetInput(sel string, port Port) error {
	if sel != "" {
		return fmt.Errorf("'not' block has no named input '%s'", sel)
	}
	i, ok := port.(*bool)
	if !ok {
		return fmt.Errorf("'not' block needs bool, not %s", PortString(port))
	}
	b.i = i
	return nil
}

type logicopblk struct {
	boolblk
	i1, i2 *bool
	tick   func(a, b bool) bool
}

func (b *logicopblk) Tick() {
	b.o = b.tick(*b.i1, *b.i2)
}

func (b *logicopblk) InputNames() []string { return []string{"1", "2"} }
func (b *logicopblk) SetInput(sel string, port Port) error {
	if sel != "1" && sel != "2" {
		return fmt.Errorf("'%s' has inputs '1' and '2', but not '%s'", b.typ, sel)
	}
	i, ok := port.(*bool)
	if !ok {
		return fmt.Errorf("'%s' block needs bool, not %s", b.typ, PortString(port))
	}
	if sel == "1" {
		b.i1 = i
	} else {
		b.i2 = i
	}
	return nil
}

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
