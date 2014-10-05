package parser

type Param interface {
	reader(globals NamedParam) ParamReader
}

type PosParam []float64

func (p PosParam) reader(globals NamedParam) ParamReader { return newpospr(p, globals) }

type NamedParam map[string]float64

func (p NamedParam) reader(globals NamedParam) ParamReader { return newnamedpr(p, globals) }

type ParamReader interface {
	Arg(n string) float64
	OptArg(n string, def float64) float64
	Err() error
}

func NewParamReader(p Param, globals NamedParam) ParamReader {
	if p != nil {
		return p.reader(globals)
	}
	return &emptypr{globals, nil}
}

type pospr struct {
	v       []float64
	globals map[string]float64

	idx      int
	firsterr error
}

func newpospr(p PosParam, globals NamedParam) ParamReader { return &pospr{v: p, globals: globals} }

func (p *pospr) Arg(n string) float64 {
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

func (p *pospr) OptArg(n string, def float64) float64 {
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

func (p *pospr) Err() error {
	if p.firsterr != nil {
		return p.firsterr
	}
	if p.idx < len(p.v) {
		return errf("too many arguments (needs at most %d, have %d)", p.idx, len(p.v))
	}
	return nil
}

type namedpr struct {
	m       map[string]float64
	globals map[string]float64

	used     map[string]bool
	firsterr error
}

func newnamedpr(p, globals NamedParam) ParamReader {
	return &namedpr{m: p, globals: globals, used: make(map[string]bool)}
}

func (p *namedpr) Arg(n string) float64 {
	v, ok := p.val(n, 0)
	if !ok && p.firsterr == nil {
		p.firsterr = errf("argument '%s' missing", n)
	}
	return v
}

func (p *namedpr) OptArg(n string, def float64) float64 {
	v, _ := p.val(n, def)
	return v
}

func (p *namedpr) val(n string, def float64) (v float64, ok bool) {
	if v, ok = p.m[n]; ok {
		p.used[n] = true
		return
	}
	if v, ok = p.globals[n]; ok {
		return
	}
	return def, false
}

func (p *namedpr) Err() error {
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

type emptypr struct {
	globals map[string]float64

	firsterr error
}

func (p *emptypr) Arg(n string) float64 {
	v, ok := p.val(n, 0)
	if !ok && p.firsterr == nil {
		p.firsterr = errf("argument '%s' missing", n)
	}
	return v
}

func (p *emptypr) OptArg(n string, def float64) float64 {
	v, _ := p.val(n, def)
	return v
}

func (p *emptypr) Err() error {
	if p.firsterr != nil {
		return p.firsterr
	}
	return nil
}

func (p *emptypr) val(n string, def float64) (v float64, ok bool) {
	if v, ok = p.globals[n]; ok {
		return
	}
	return def, false
}
