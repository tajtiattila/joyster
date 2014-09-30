package logic

import (
	"fmt"
	"github.com/tajtiattila/joyster/block"
	"math"
)

func init() {
	// zero positions inside the circle, move outside closer to center
	block.RegisterStickFunc("circulardeadzone", func(p block.Param) (block.StickFunc, error) {
		value := p.Arg("threshold")
		value2 := value * value
		return func(xi, yi float64) (xo, yo float64) {
			mag2 := xi*xi + yi*yi
			if mag2 > value2 {
				mag := math.Sqrt(mag2)
				m := 1 - value/mag
				xo = xi * m
				yo = yi * m
			}
			return
		}, nil
	})

	// the circular positions into positions on the square (0 < factor < 1)
	block.RegisterStickFunc("circlesquare", func(p block.Param) (block.StickFunc, error) {
		factor := p.OptArg("factor", 1)
		if factor < 0 || 1 < factor {
			return nil, fmt.Errorf("circlesquare factor should be between 0 and 1, not %f", factor)
		}
		of, mf := 1-factor, factor
		return func(xi, yi float64) (xo, yo float64) {
			if xi*xi+yi*yi < 1e-3 {
				return xi, yi
			}
			xa, ya := math.Abs(xi), math.Abs(yi)
			var u float64
			if xa > ya {
				u = ya / xa // yu/xu = ya/xa
			} else {
				u = xa / ya // xu/yu = xa/ya
			}
			m := math.Sqrt(1+u*u) * mf
			xo = xi * (of + m)
			yo = yi * (of + m)
			return
		}, nil
	})

}
