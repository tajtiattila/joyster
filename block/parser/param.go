package parser

import (
	"github.com/tajtiattila/joyster/block"
)

type paramspec interface {
	Prepare(c *context) param
}

type param interface {
	block.Param
	err() error
}

type posparamspec []float64

func (spec posparamspec) Prepare(c *context) param {
	return &positionalparam{
		v:       spec,
		globals: c.config,
	}
}

type namedparamspec map[string]float64

func (spec namedparamspec) Prepare(c *context) param {
	return &namedparam{
		m:       spec,
		globals: c.config,
	}
}

type positionalparam struct {
	v       []float64
	globals map[string]float64

	idx      int
	firsterr error
}

func (p *positionalparam) Arg(n string) float64 {
	if p.idx < len(p.v) {
		i := p.idx
		p.idx++
		return p.v[i]
	}
	if v, ok := p.globals[n]; ok {
		return v
	}
	if p.firsterr == nil {
		p.firsterr = errf("argument '%s' missing at position %d missing", n, p.idx)
	}
	return 0
}

func (p *positionalparam) OptArg(n string, def float64) float64 {
	if p.idx < len(p.v) {
		i := p.idx
		p.idx++
		return p.v[i]
	}
	if v, ok := p.globals[n]; ok {
		return v
	}
	return def
}

func (p *positionalparam) TickFreq() float64 {
	return 1 / p.TickTime()
}

func (p *positionalparam) TickTime() float64 {
	return p.OptArg("Update", block.DefaultTickTime)
}

func (p *positionalparam) err() error {
	if p.firsterr != nil {
		return p.firsterr
	}
	if p.idx < len(p.v) {
		return errf("too many arguments (needs at most %d, have %d)", p.idx, len(p.v))
	}
	return nil
}

type namedparam struct {
	m       map[string]float64
	globals map[string]float64

	used     map[string]bool
	firsterr error
}

func (p *namedparam) Arg(n string) float64 {
	v, ok := p.val(n, 0)
	if !ok && p.firsterr == nil {
		p.firsterr = errf("argument '%s' missing", n)
	}
	return v
}

func (p *namedparam) OptArg(n string, def float64) float64 {
	v, _ := p.val(n, def)
	return v
}

func (p *namedparam) TickFreq() float64 {
	return 1 / p.TickTime()
}

func (p *namedparam) TickTime() float64 {
	return p.OptArg("Update", block.DefaultTickTime)
}

func (p *namedparam) val(n string, def float64) (v float64, ok bool) {
	if v, ok = p.m[n]; ok {
		p.used[n] = true
		return
	}
	if v, ok = p.globals[n]; ok {
		return
	}
	return def, false
}

func (p *namedparam) err() error {
	if p.firsterr != nil {
		return p.firsterr
	}
	if len(p.m) != len(p.used) {
		for n := range p.m {
			if !p.used[n] {
				return errf("named parameter '%s' unknown", n)
			}
		}
	}
	return nil
}
