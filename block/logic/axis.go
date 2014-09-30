package logic

import (
	"github.com/tajtiattila/joyster/block"
	"math"
)

func init() {
	// add value to input
	block.RegisterScalarFunc("offset", func(p block.Param) (func(float64) float64, error) {
		ofs := p.Arg("value")
		return func(v float64) float64 {
			return v + ofs
		}, nil
	})

	// zero input under abs. value, reduce bigger
	block.RegisterScalarFunc("deadzone", func(p block.Param) (func(float64) float64, error) {
		dz := p.Arg("treshold")
		return func(v float64) float64 {
			var s float64
			if v < 0 {
				v, s = -v, -1
			} else {
				s = 1
			}
			v -= dz
			if v < 0 {
				return 0
			}
			return v * s
		}, nil
	})

	// multiply input by factor
	block.RegisterScalarFunc("multiply", func(p block.Param) (func(float64) float64, error) {
		f := p.Arg("factor")
		return func(v float64) float64 {
			return v * f
		}, nil
	})

	// axis sensitivivy curve (factor: 0 - linear, positive: nonlinear)
	block.RegisterScalarFunc("curvature", func(p block.Param) (func(float64) float64, error) {
		pow := math.Pow(2, p.Arg("factor"))
		return func(v float64) float64 {
			s := float64(1)
			if v < 0 {
				s, v = -1, -v
			}
			return s * math.Pow(v, pow)
		}, nil
	})

	// truncate input above abs. value
	block.RegisterScalarFunc("truncate", func(p block.Param) (func(float64) float64, error) {
		t := p.Arg("value")
		return func(v float64) float64 {
			switch {
			case v < -t:
				return -t
			case t < v:
				return t
			}
			return v
		}, nil
	})

	// set maximum input change to value/second
	block.RegisterScalarFunc("dampen", func(p block.Param) (func(float64) float64, error) {
		value := p.Arg("value")
		if value < 1e-6 {
			return func(v float64) float64 {
				return v
			}, nil
		}
		speed := p.TickTime() / value
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
		}, nil
	})

	// smooth inputs over time (seconds)
	block.RegisterScalarFunc("smooth", func(p block.Param) (func(float64) float64, error) {
		nsamples := math.Floor(p.Arg("time") * p.TickFreq())
		if nsamples < 2 {
			return func(v float64) float64 {
				return v
			}, nil
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
		}, nil
	})

	// use input as delta, change values by speed/second
	block.RegisterScalarFunc("incremental", func(p block.Param) (func(float64) float64, error) {
		speed := p.Arg("speed")
		rebound := p.OptArg("rebound", 0)
		quickcenter := 0 != p.OptArg("quickcenter", 0)

		speed *= p.TickTime()
		rebound *= p.TickTime()

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
		}, nil
	})
}
