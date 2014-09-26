package parser

import (
	"errors"
	"fmt"
	"github.com/tajtiattila/joyster/block"
)

type factoryblkspec struct {
	lineno int
	xy     bool
	typ    string
	inputs []blockspec
	param  *block.Param

	blk block.Block
}

func (b *factoryblkspec) SourceLine() int { return b.lineno }

func (b *factoryblkspec) Prepare(c *context) (block.Block, error) {
	if b.blk == nil {
		var err error
		if b.blk, err = c.createBlock(b.typ); err != nil {
			return nil, srcerr(b, err)
		}
		if b.param != nil {
			sup, ok := b.blk.(block.Setupper)
			if !ok {
				return nil, srcerrf(b, "block type '%s' doesn't support parameters", b.typ)
			}
			if err = sup.Setup(b.param); err != nil {
				return nil, srcerr(b, err)
			}
		}
		if len(b.inputs) != 0 {
			is, ok := b.blk.(block.InputSetter)
			if !ok {
				return nil, srcerrf(b, "block type '%s' doesn't support inputs", b.typ)
			}
			names := is.InputNames()
			if len(names) != is.NMinInput() {
				if len(b.inputs) < is.NMinInput() {
					return nil, srcerrf(b, "block type '%s' needs at least %d inputs, not %d", b.typ, is.NMinInput(), len(b.inputs))
				} else if len(names) < len(b.inputs) {
					return nil, srcerrf(b, "block type '%s' needs at most %d inputs, not %d", b.typ, len(names), len(b.inputs))
				}
			} else {
				if len(b.inputs) != len(names) {
					return nil, srcerrf(b, "block type '%s' needs exactly %d inputs, not %d", b.typ, len(names), len(b.inputs))
				}
			}
			for i, input := range b.inputs {
				cblk, err := input.Prepare(c)
				if err != nil {
					return nil, srcerr(b, err)
				}
				port, err := cblk.Output("")
				if err != nil {
					return nil, srcerr(b, err)
				}
				err = is.SetInput(names[i], port)
				if err != nil {
					return nil, srcerr(b, err)
				}
			}
		}
		if t, ok := b.blk.(block.Ticker); ok {
			c.rdep = append(c.rdep, t)
		}
	}
	return b.blk, nil
}

type groupblkspec struct {
	lineno int
	sels   []string
	v      []*factoryblkspec
}

func (b *groupblkspec) SourceLine() int { return b.lineno }

func (b *groupblkspec) Prepare(c *context) (block.Block, error) {
	var first block.InputSetter
	var last block.Block
	var tickers []block.Ticker
	for _, cblks := range b.v {
		cblk, err := c.createBlock(cblks.typ)
		if err != nil {
			return nil, srcerr(b, err)
		}
		is, ok := cblk.(block.InputSetter)
		if !ok {
			return nil, srcerrf(cblks, "block type '%s' doesn't support inputs", cblks.typ)
		}
		if len(b.sels) != 0 {
			if has(is.InputNames(), "") {
				v := []block.Block{cblk}
				for len(v) < len(is.InputNames()) {
					xblk, err := c.createBlock(cblks.typ)
					if err != nil {
						return nil, srcerr(b, err)
					}
					v = append(v, xblk)
				}
				if _, ok := cblk.(block.Ticker); ok {
					mt := make([]block.Ticker, len(v))
					for i := range v {
						mt[i] = v[i].(block.Ticker)
					}
					cblk = &tickermultiblk{simplemultiblk{b.sels, v}, mt}
				} else {
					cblk = &simplemultiblk{b.sels, v}
				}
			} else {
				for _, n := range b.sels {
					if !has(is.InputNames(), n) {
						return nil, srcerrf(cblks, "block type '%s' has no input '%s'", cblks.typ, n)
					}
					if !has(cblk.OutputNames(), n) {
						return nil, srcerrf(cblks, "block type '%s' has no output '%s'", cblks.typ, n)
					}
				}
			}
		}
		if first == nil {
			first = is
		} else {
			for _, n := range last.OutputNames() {
				p, err := last.Output(n)
				if err != nil {
					return nil, srcerr(b, err)
				}
				err = is.SetInput(n, p)
				if err != nil {
					return nil, srcerr(b, err)
				}
			}
		}
		last = cblk
		if t, ok := cblk.(block.Ticker); ok {
			tickers = append(tickers, t)
		}
	}
	for len(tickers) > 0 {
		var t block.Ticker
		n := len(tickers) - 1
		t, tickers = tickers[n], tickers[:n]
		c.rdep = append(c.rdep, t)
	}
	return &groupblk{first, last}, nil
}

func has(v []string, s string) bool {
	for _, w := range v {
		if s == w {
			return true
		}
	}
	return false
}

type simplemultiblk struct {
	names []string
	v     []block.Block
}

func (b *simplemultiblk) NMinInput() int       { return len(b.names) }
func (b *simplemultiblk) InputNames() []string { return b.names }

func (b *simplemultiblk) SetInput(sel string, input block.Port) error {
	for i, n := range b.names {
		if n == sel {
			return b.v[i].(block.InputSetter).SetInput("", input)
		}
	}
	return fmt.Errorf("simplemultiblk has no input '%s'", sel)
}

func (b *simplemultiblk) OutputNames() []string { return b.names }

func (b *simplemultiblk) Output(sel string) (block.Port, error) {
	for i, n := range b.names {
		if n == sel {
			return b.v[i].Output("")
		}
	}
	return nil, fmt.Errorf("simplemultiblk has no output '%s'", sel)
}

type tickermultiblk struct {
	simplemultiblk
	vt []block.Ticker
}

func (b *tickermultiblk) Tick() {
	for _, t := range b.vt {
		t.Tick()
	}
}

type groupblk struct {
	first block.InputSetter
	last  block.Block
}

func (g *groupblk) NMinInput() int       { return g.first.NMinInput() }
func (g *groupblk) InputNames() []string { return g.first.InputNames() }

func (g *groupblk) SetInput(sel string, input block.Port) error {
	return g.first.SetInput(sel, input)
}

func (g *groupblk) OutputNames() []string {
	return g.last.OutputNames()
}

func (g *groupblk) Output(sel string) (block.Port, error) {
	return g.last.Output(sel)
}

type constblkspec struct {
	p block.Port
}

func (b *constblkspec) Lineno() int { return -1 }

func (b *constblkspec) Prepare(c *context) (block.Block, error) {
	return b, nil
}

func (b *constblkspec) Setup(*block.Param) error { return srcerrf(b, "const does not support setup") }
func (b *constblkspec) OutputNames() []string    { return []string{""} }

func (b *constblkspec) Output(sel string) (block.Port, error) {
	if sel != "" {
		return nil, srcerrf(b, "const has no named port %s", sel)
	}
	return b.p, nil
}

func (*constblkspec) SourceLine() int { return -1 }

type valueblkspec struct {
	constblkspec
	lineno int
}

func (b *valueblkspec) SourceLine() int { return b.lineno }

type namedblkspec struct {
	lineno    int
	name, sel string
}

func (b *namedblkspec) SourceLine() int { return b.lineno }
func (b *namedblkspec) Prepare(c *context) (block.Block, error) {
	blks, ok := c.blocks[b.name]
	if !ok {
		return nil, srcerrf(b, "input %s is missing", b.name)
	}
	blk, err := blks.Prepare(c)
	if err != nil {
		return nil, srcerr(b, err)
	}
	return &namedport{b.lineno, blk, b.sel}, nil
}

type namedport struct {
	lineno int
	blk    block.Block
	sel    string
}

func (p *namedport) Setup(*block.Param) error { return srcerr(p, "named port does not have setup") }
func (p *namedport) SourceLine() int          { return p.lineno }
func (p *namedport) OutputNames() []string    { return []string{""} }
func (p *namedport) Output(sel string) (block.Port, error) {
	if sel != "" {
		return nil, srcerrf(p, "internal error: selector '%s' on namedblkspec is impossible", sel)
	}
	return p.blk.Output(p.sel)
}

type sourceliner interface {
	SourceLine() int
}

type connspec struct {
	name, sel string
	blockspec

	lineno int
}

type sourceerror struct {
	lineno int
	err    error
}

func (e *sourceerror) Error() string {
	return fmt.Sprintf("line %d: ", e.lineno) + e.err.Error()
}

func srcerr(s sourceliner, i interface{}) error {
	n := s.SourceLine()
	switch x := i.(type) {
	case *sourceerror:
		return x
	case error:
		return &sourceerror{n, x}
	}
	return &sourceerror{n, errors.New(fmt.Sprint(i))}
}

func srcerrf(s sourceliner, f string, args ...interface{}) error {
	return srcerr(s, fmt.Sprintf(f, args...))
}
