package main

import (
	"fmt"
	"math"
)

func NewAxisLogic(c *Config, args interface{}) (f AxisFunc, err error) {
	var fc *FilterCall
	fc, err = NewFilterCall(axisFuncs, args)
	if err != nil {
		return nil, err
	}
	fac := fc.Sig.Factory.(AxisFuncFactory)
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case *ErrArgDecode:
				f, err = nil, fmt.Errorf("argument decode error (%s) in function %s", r.t, fc.Sig.Name)
			default:
				f, err = nil, fmt.Errorf("%s", r)
			}
		}
	}()
	return fac(c, fc.Argv), nil
}

type AxisFunc func(float64) float64

type AxisFuncFactory func(*Config, FilterArgs) AxisFunc

var axisFuncs = make(FilterMap)

func init() {
	FillFilterMap(axisFuncs,
		axisf(Sig("offset", F("value")),
			"add value to input",
			func(c *Config, p FilterArgs) AxisFunc {
				ofs := p.F()
				return func(v float64) float64 {
					return v + ofs
				}
			}),
		axisf(Sig("deadzone", F("value")),
			"zero input under abs. value, reduce bigger",
			func(c *Config, p FilterArgs) AxisFunc {
				dz := p.F()
				return func(v float64) float64 {
					var s float64
					if v < 0 {
						v, s = -v, -1
					} else {
						s = 1
					}
					if v < dz {
						return 0
					}
					return v * s
				}
			}),
		axisf(Sig("multiplier", F("factor")),
			"multiply input by factor",
			func(c *Config, p FilterArgs) AxisFunc {
				m := p.F()
				return func(v float64) float64 {
					return v * m
				}
			}),
		axisf(Sig("curvature", F("factor")),
			"axis sensitivivy curve (factor: 0 - linear, positive: nonlinear)",
			func(c *Config, p FilterArgs) AxisFunc {
				pow := math.Pow(2, p.F())
				return func(v float64) float64 {
					s := float64(1)
					if v < 0 {
						s, v = -1, -v
					}
					return s * math.Pow(v, pow)
				}
			}),
		axisf(Sig("truncate", OptF("value", 1)),
			"truncate input above abs. value",
			func(c *Config, p FilterArgs) AxisFunc {
				t := p.F()
				return func(v float64) float64 {
					switch {
					case v < -t:
						return -t
					case t < v:
						return t
					}
					return v
				}
			}),
		axisf(Sig("dampen", F("value")),
			"set maximum input change to value/second",
			func(c *Config, p FilterArgs) AxisFunc {
				value := p.F()
				if value < 1e-6 {
					return func(v float64) float64 {
						return v
					}
				}
				dt := float64(c.UpdateMicros) / 1e6
				speed := dt / value

				var pos float64

				return func(v float64) float64 {
					switch {
					case pos+speed < v:
						pos += speed
					case v < pos-speed:
						pos -= speed
					default:
						pos = v
					}
					return pos
				}
			}),
		axisf(Sig("smooth", F("time")),
			"smooth inputs over time (seconds)",
			func(c *Config, p FilterArgs) AxisFunc {
				tickpersec := 1e6 / float64(c.UpdateMicros)
				nsamples := math.Floor(p.F() * tickpersec)
				if nsamples < 2 {
					return func(v float64) float64 {
						return v
					}
				}

				m0 := math.Pow(2, 63) / (nsamples * 100)
				m1 := 1 / (m0 * nsamples)

				posv := make([]int64, int(nsamples))
				n := 0
				var sum int64

				return func(v float64) float64 {
					iv := int64(v * m0)
					sum -= posv[n]
					posv[n], n = iv, (n+1)%len(posv)
					sum += iv
					return float64(sum) * m1
				}
			}),
		axisf(Sig("incremental", F("speed"), OptF("rebound", 0.0), OptB("quickcenter", false)),
			"use input as delta, change values by speed/second",
			func(c *Config, p FilterArgs) AxisFunc {
				var (
					speed, rebound float64
					quickcenter    bool
				)
				p.Args(&speed, &rebound, &quickcenter)
				dt := float64(c.UpdateMicros) / 1e6
				speed *= dt
				rebound *= dt

				var pos float64

				return func(v float64) float64 {
					if math.Abs(v) < 1e-3 {
						switch {
						case pos < -rebound:
							pos += rebound
						case rebound < pos:
							pos -= rebound
						default:
							pos = 0
						}
					} else {
						if quickcenter && pos*v < 0 {
							pos = 0
						} else {
							pos += v * speed
							switch {
							case pos < -1:
								pos = -1
							case 1 < pos:
								pos = 1
							}
						}
					}
					return pos
				}
			}),
		axisf(Sig("input"),
			"return input unchanged",
			func(c *Config, p FilterArgs) AxisFunc {
				p.Args()
				return func(v float64) float64 {
					return v
				}
			}),
		axisf(Sig("chain", Subv("subfilter", axisFuncs)),
			"apply subfilters one after the other",
			func(c *Config, p FilterArgs) AxisFunc {
				var fcv []*FilterCall
				p.Args(&fcv)
				subv := make([]AxisFunc, len(fcv))
				for i, fc := range fcv {
					fac := fc.Sig.Factory.(AxisFuncFactory)
					subv[i] = fac(c, fc.Argv)
				}
				return func(v float64) float64 {
					for _, sub := range subv {
						v = sub(v)
					}
					return v
				}
			}),
		axisf(Sig("absmin", Subv("subfilter", axisFuncs)),
			"feed input through subfilters, set output to smallest",
			func(c *Config, p FilterArgs) AxisFunc {
				var fcv []*FilterCall
				p.Args(&fcv)
				if len(fcv) < 2 {
					panic("absmin needs at least 2 subfilters")
				}
				subv := make([]AxisFunc, len(fcv))
				for i, fc := range fcv {
					fac := fc.Sig.Factory.(AxisFuncFactory)
					subv[i] = fac(c, fc.Argv)
				}
				return func(v float64) float64 {
					vv := subv[0](v)
					for _, sub := range subv[1:] {
						vx := sub(v)
						if math.Abs(vx) < math.Abs(vv) {
							vv = vx
						}
					}
					return vv
				}
			}),
	)
}

func axisf(s *FilterSig, h string, f AxisFuncFactory) *FilterSig {
	s.Help = h
	s.Factory = f
	return s
}
