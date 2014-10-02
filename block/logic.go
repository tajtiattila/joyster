package block

import (
	"fmt"
)

func RegisterLogicFunc(name string, fn func(a, b bool) bool) {
	Register(name, func() Block {
		return &logicopblk{typ: name, tick: fn}
	})
}

func init() {
	Register("not", func() Block { return new(notblk) })
	RegisterLogicFunc("and", func(a, b bool) bool { return a && b })
	RegisterLogicFunc("or", func(a, b bool) bool { return a || b })
	RegisterLogicFunc("xor", func(a, b bool) bool { return a != b })
	Register("if", func() Block { return new(ifblk) })
}

type notblk struct {
	o bool
	i *bool
}

func (b *notblk) Tick()             { b.o = !*b.i }
func (b *notblk) Input() InputMap   { return SingleInput("not", &b.i) }
func (b *notblk) Output() OutputMap { return SingleOutput("not", &b.o) }
func (b *notblk) Validate() error   { return CheckInputs("not", &b.i) }

type logicopblk struct {
	typ    string
	o      bool
	i1, i2 *bool
	tick   func(a, b bool) bool
}

func (b *logicopblk) Tick() {
	b.o = b.tick(*b.i1, *b.i2)
}

func (b *logicopblk) Input() InputMap {
	return MapInput(b.typ, map[string]interface{}{"1": &b.i1, "2": &b.i2})
}
func (b *logicopblk) Output() OutputMap { return SingleOutput(b.typ, &b.o) }
func (b *logicopblk) Validate() error   { return CheckInputs("not", &b.i1, &b.i2) }

type ifblk struct {
	cond             *bool
	valthen, valelse Port

	out  Port
	tick func()
}

func (b *ifblk) Output() OutputMap { return SingleOutput("if", b.out) }
func (b *ifblk) Input() InputMap   { return &ifinput{b} }
func (b *ifblk) Validate() error {
	return CheckInputs("if", &b.cond, portpt(b.valthen), portpt(b.valelse))
}
func (b *ifblk) Tick() { b.tick() }

type ifinput struct {
	b *ifblk
}

func (inp *ifinput) Names() []string { return []string{"cond", "then", "else"} }
func (inp *ifinput) Set(sel string, port Port) error {
	b := inp.b
	switch sel {
	case "cond":
		var ok bool
		b.cond, ok = port.(*bool)
		if !ok {
			return fmt.Errorf("if block needs bool condition, not %s", PortString(port))
		}
		if b.cond == nil {
			return fmt.Errorf("if block 'cond' input is nil")
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
			return fmt.Errorf("then and else must have the same type, has %s and %s", PortString(b.valthen), PortString(b.valelse))
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
		default:
			return fmt.Errorf("'if' internal error")
		}
	}
	return nil
}

func (inp *ifinput) Type(sel string) PortType {
	switch sel {
	case "cond":
		return Bool
	case "then":
		return Any
	case "else":
		return Any
	}
	panic(fmt.Sprint("if block has no input named '%s'", sel))
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

func portpt(port Port) interface{} {
	switch x := port.(type) {
	case *bool:
		return &x
	case *float64:
		return &x
	case *int:
		return &x
	}
	panic("portpt invalid")
}
