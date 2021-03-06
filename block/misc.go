package block

import (
	"fmt"
	"math"
)

const debug = false

func RegisterScalarFunc(name string, fn func(Param) (func(float64) float64, error)) {
	RegisterParam(name, func(p Param) (Block, error) {
		f, err := fn(p)
		if err != nil {
			return nil, err
		}
		return &scalarfnblk{typ: name, f: f}, nil
	})
}

type scalarfnblk struct {
	typ string
	i   *float64
	o   float64
	f   func(float64) float64
}

func (b *scalarfnblk) Tick() {
	b.o = b.f(*b.i)
	if debug {
		if math.IsNaN(b.o) {
			panic(b.typ + " yielded NaN")
		}
	}
}
func (b *scalarfnblk) Input() InputMap   { return SingleInput(b.typ, &b.i) }
func (b *scalarfnblk) Output() OutputMap { return SingleOutput(b.typ, &b.o) }
func (b *scalarfnblk) Validate() error   { return CheckInputs(b.typ, &b.i) }

var unsetBool = new(bool)

func init() {
	RegisterType(&Proto{"toggle", false, func(Param) (Block, error) {
		return &toggle{
			i:     unsetBool,
			set:   unsetBool,
			reset: unsetBool,
		}, nil
	}})
}

type toggle struct {
	i, set, reset *bool
	o             bool

	il, sl, rl bool
}

func (b *toggle) Tick() {
	if *b.i == true && b.il == false {
		b.o = !b.o
	}
	if *b.set == true && b.sl == false {
		b.o = true
	}
	if *b.reset == true && b.rl == false {
		b.o = false
	}
	b.il, b.sl, b.rl = *b.i, *b.set, *b.reset
}

func (b *toggle) Input() InputMap {
	return MapInput("toggle", pt("", &b.i), pt("set", &b.set), pt("reset", &b.reset))
}
func (b *toggle) Output() OutputMap { return SingleOutput("toggle", &b.o) }
func (b *toggle) Validate() error {
	if err := CheckInputs("toggle", &b.i, &b.set, &b.reset); err != nil {
		return err
	}
	if b.i == unsetBool {
		if b.set == unsetBool || b.reset == unsetBool {
			return fmt.Errorf("'toggle' unnamed input is unassigned")
		}
	}
	return nil
}

func RegisterBoolFunc(name string, fn func(Param) (func(bool) bool, error)) {
	RegisterParam(name, func(p Param) (Block, error) {
		f, err := fn(p)
		if err != nil {
			return nil, err
		}
		return &boolfnblk{typ: name, f: f}, nil
	})
}

type boolfnblk struct {
	typ string
	i   *bool
	o   bool
	f   func(bool) bool
}

func (b *boolfnblk) Tick()             { b.o = b.f(*b.i) }
func (b *boolfnblk) Input() InputMap   { return SingleInput(b.typ, &b.i) }
func (b *boolfnblk) Output() OutputMap { return SingleOutput(b.typ, &b.o) }
func (b *boolfnblk) Validate() error   { return CheckInputs(b.typ, &b.i) }

type StickFunc func(xi, yi float64) (xo, yo float64)

func RegisterStickFunc(name string, ff func(p Param) (StickFunc, error)) {
	RegisterParam(name, func(p Param) (Block, error) {
		f, err := ff(p)
		if err != nil {
			return nil, err
		}
		b := &stickfuncblk{typ: name, f: f}
		return b, nil
	})
}

type stickfuncblk struct {
	typ    string
	xi, yi *float64
	xo, yo float64
	f      func(xi, yi float64) (xo, yo float64)
}

func (b *stickfuncblk) Input() InputMap   { return MapInput(b.typ, pt("x", &b.xi), pt("y", &b.yi)) }
func (b *stickfuncblk) Output() OutputMap { return MapOutput(b.typ, pt("x", &b.xo), pt("y", &b.yo)) }
func (b *stickfuncblk) Validate() error   { return CheckInputs(b.typ, &b.xi, &b.yi) }
func (b *stickfuncblk) Tick()             { b.xo, b.yo = b.f(*b.xi, *b.yi) }

type HatFunc func(xi, yi int) int

func RegisterHatFunc(name string, f HatFunc) {
	Register(name, func() Block { return &hatfuncblk{typ: name, f: f} })
}

type hatfuncblk struct {
	typ    string
	xi, yi *int
	o      int
	f      func(a, b int) int
}

func (b *hatfuncblk) Input() InputMap   { return MapInput(b.typ, pt("x", &b.xi), pt("y", &b.yi)) }
func (b *hatfuncblk) Output() OutputMap { return SingleOutput(b.typ, pt("", &b.o)) }
func (b *hatfuncblk) Validate() error   { return CheckInputs(b.typ, &b.xi, &b.yi) }
func (b *hatfuncblk) Tick()             { b.o = b.f(*b.xi, *b.yi) }

func init() {
	Register("stick", func() Block { return new(stickblk) })
}

type stickblk struct {
	x, y *float64
}

func (b *stickblk) Input() InputMap   { return MapInput("stick", pt("x", &b.x), pt("y", &b.y)) }
func (b *stickblk) Output() OutputMap { return MapOutput("stick", pt("x", b.x), pt("y", b.y)) }
func (b *stickblk) Validate() error   { return CheckInputs("stick", &b.x, &b.y) }
