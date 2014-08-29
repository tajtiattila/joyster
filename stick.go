package main

import (
	"fmt"
	"math"
)

func NewStickLogic(c *Config, args interface{}) (l StickLogic, err error) {
	var fc *FilterCall
	fc, err = NewFilterCall(stickFuncs, args)
	if err != nil {
		ax, err := NewAxisLogic(c, args)
		if err != nil {
			return nil, err
		}
		ay, err := NewAxisLogic(c, args)
		if err != nil {
			return nil, err
		}
		return StickFunc(func(p *StickPos) {
			p.X = ax(p.X)
			p.Y = ay(p.Y)
		}), nil
	}
	fac := fc.Sig.Factory.(StickLogicFactory)
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case *ErrArgDecode:
				l, err = nil, fmt.Errorf("argument decode error in function %s", fc.Sig.Name)
			default:
				l, err = nil, fmt.Errorf("%s", r)
			}
		}
	}()
	return fac(c, fc.Argv), nil
}

var stickFuncs = make(FilterMap)

func init() {
	FillFilterMap(stickFuncs,
		stickf(Sig("x", Sub("filter", axisFuncs)),
			"apply axis filter to x position",
			func(c *Config, p FilterArgs) StickLogic {
				var fc *FilterCall
				p.Args(&fc)
				fac := fc.Sig.Factory.(AxisFuncFactory)
				a := fac(c, fc.Argv)
				return StickFunc(func(p *StickPos) {
					p.X = a(p.X)
				})
			}),
		stickf(Sig("y", Sub("filter", axisFuncs)),
			"apply axis filter to y position",
			func(c *Config, p FilterArgs) StickLogic {
				var fc *FilterCall
				p.Args(&fc)
				fac := fc.Sig.Factory.(AxisFuncFactory)
				a := fac(c, fc.Argv)
				return StickFunc(func(p *StickPos) {
					p.Y = a(p.Y)
				})
			}),
		stickf(Sig("xy", Sub("filter", axisFuncs)),
			"apply axis filter to both x and y position",
			func(c *Config, p FilterArgs) StickLogic {
				var fc *FilterCall
				p.Args(&fc)
				fac := fc.Sig.Factory.(AxisFuncFactory)
				ax := fac(c, fc.Argv)
				ay := fac(c, fc.Argv)
				return StickFunc(func(p *StickPos) {
					p.X = ax(p.X)
					p.Y = ay(p.Y)
				})
			}),
		stickf(Sig("circle_to_square", OptF("factor", 1.0)),
			"convert the circular positions into positions on the square (0 < factor < 1)",
			func(c *Config, p FilterArgs) StickLogic {
				factor := p.F()
				if factor < 0 || 1 < factor {
					fmt.Println("circle_to_square factor should be between 0 and 1")
				}
				of, mf := 1-factor, factor
				return StickFunc(func(p *StickPos) {
					if p.X*p.X+p.Y*p.Y < 1e-3 {
						return
					}

					xa, ya := math.Abs(p.X), math.Abs(p.Y)

					var u float64
					if xa > ya {
						u = ya / xa // yu/xu = ya/xa
					} else {
						u = xa / ya // xu/yu = xa/ya
					}
					m := math.Sqrt(1+u*u) * mf
					p.X *= (of + m)
					p.Y *= (of + m)
				})
			}),
		stickf(Sig("circular_deadzone", F("radius")),
			"zero positions inside the circle, move outside closer to center",
			func(c *Config, p FilterArgs) StickLogic {
				value := p.F()
				value2 := value * value
				return StickFunc(func(p *StickPos) {
					mag2 := p.X*p.X + p.Y*p.Y
					if mag2 <= value2 {
						p.X, p.Y = 0, 0
					} else {
						mag := math.Sqrt(mag2)
						m := 1 - value/mag
						p.X *= m
						p.Y *= m
					}
				})
			}),
	)
}

func stickf(s *FilterSig, h string, f StickLogicFactory) *FilterSig {
	s.Help = h
	s.Factory = f
	return s
}

type StickLogic interface {
	Visit(*StickPos)
}

type StickLogicFactory func(*Config, FilterArgs) StickLogic

type StickPos struct {
	X, Y float64
}

func (p *StickPos) Init(x, y int16) {
	p.X = float64sFromInt16(x)
	p.Y = float64sFromInt16(y)
}

func (p *StickPos) Apply(l StickLogic) {
	l.Visit(p)
}

func (p *StickPos) Values() (x, y float32) {
	return float32(p.X), float32(p.Y)
}

type StickFunc func(p *StickPos)

func (f StickFunc) Visit(p *StickPos) {
	f(p)
}
