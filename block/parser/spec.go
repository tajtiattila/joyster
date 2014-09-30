package parser

import (
	"github.com/tajtiattila/joyster/block"
)

type factoryblkspec struct {
	lineno int
	dollar bool
	typ    string
	inputs []blockspec
	param  paramspec

	blk block.Block
}

func (b *factoryblkspec) sourceline() int { return b.lineno }

func (b *factoryblkspec) Prepare(c *context) (block.Block, error) {
	if b.blk == nil {
		var err error
		if b.blk, err = b.create(c, false); err != nil {
			return nil, srcerr(b, err)
		}
		if err = b.inputsetup(c, b.blk); err != nil {
			return nil, err
		}
		c.addBlock(b.blk)
	}
	return b.blk, nil
}

func (b *factoryblkspec) create(c *context, grp bool) (block.Block, error) {
	blk, err := b.newBlock(c)
	if err != nil {
		return nil, err
	}
	if b.dollar {
		xblk := blk
		yblk, err := b.newBlock(c)
		if err != nil {
			return nil, err
		}
		if _, ok := xblk.(block.InputSetter); !ok {
			return nil, srcerrf(b, "$ block '%s' must support inputs", b.typ)
		}
		if _, ok := yblk.(block.InputSetter); !ok {
			return nil, srcerrf(b, "$ block '%s' must support inputs", b.typ)
		}
		return &xyblk{xblk, yblk}, nil
	}
	return blk, nil
}

func (b *factoryblkspec) inputsetup(c *context, blk block.Block) error {
	if len(b.inputs) != 0 {
		is, ok := blk.(block.InputSetter)
		if !ok {
			return srcerrf(b, "block type '%s' doesn't accept inputs", b.typ)
		}
		names := is.InputNames()
		if len(b.inputs) != len(names) {
			return srcerrf(b, "block type '%s' needs exactly %d inputs, not %d", b.typ, len(names), len(b.inputs))
		}
		for i, input := range b.inputs {
			cblk, err := input.Prepare(c)
			if err != nil {
				return srcerr(b, err)
			}
			port, err := cblk.Output("")
			if err != nil {
				return srcerr(b, err)
			}
			err = is.SetInput(names[i], port)
			if err != nil {
				return srcerr(b, err)
			}
		}
	}
	return nil
}

func (b *factoryblkspec) newBlock(c *context) (block.Block, error) {
	param := b.param.Prepare(c)
	blk, err := c.createBlock(b.typ, param)
	if err != nil {
		return nil, srcerr(b, err)
	}
	if param.err() != nil {
		return nil, srcerr(b, param.err())
	}
	return blk, nil
}

type xyblk struct {
	x, y block.Block
}

func (b *xyblk) OutputNames() []string { return []string{"x", "y"} }
func (b *xyblk) Output(sel string) (block.Port, error) {
	switch sel {
	case "x":
		return b.x.Output("")
	case "y":
		return b.y.Output("")
	}
	return nil, errf("'xy' has outputs 'x' and 'y', but not '%s'", sel)
}

func (b *xyblk) InputNames() []string { return []string{"x", "y"} }
func (b *xyblk) SetInput(sel string, port block.Port) error {
	var child block.Block
	switch sel {
	case "x":
		child = b.x
	case "y":
		child = b.y
	default:
		return errf("'xy' has inputs 'x' and 'y', but not '%s'", sel)
	}
	is, ok := child.(block.InputSetter)
	if !ok {
		return errf("'xy' child '%s' doesn't accept inputs", sel)
	}
	return is.SetInput("", port)
}

type groupblkspec struct {
	lineno int
	sels   []string
	v      []*factoryblkspec
}

func (b *groupblkspec) sourceline() int { return b.lineno }

func (b *groupblkspec) Prepare(c *context) (block.Block, error) {
	var first block.InputSetter
	var last block.Block
	for _, cblks := range b.v {
		cblk, err := cblks.create(c, true)
		if err != nil {
			return nil, srcerr(b, err)
		}
		is, ok := cblk.(block.InputSetter)
		if !ok {
			return nil, srcerrf(cblks, "block type '%s' doesn't support inputs", cblks.typ)
		}
		sels := b.sels
		if len(sels) == 0 {
			if !has(is.InputNames(), "") {
				return nil, srcerrf(cblks, "block type '%s' has no unnamed input", cblks.typ)
			}
			if !has(cblk.OutputNames(), "") {
				return nil, srcerrf(cblks, "block type '%s' has no unnamed output", cblks.typ)
			}
			sels = []string{""}
		} else {
			for _, n := range sels {
				if !has(is.InputNames(), n) {
					return nil, srcerrf(cblks, "block type '%s' has no input '%s'", cblks.typ, n)
				}
				if !has(cblk.OutputNames(), n) {
					return nil, srcerrf(cblks, "block type '%s' has no output '%s'", cblks.typ, n)
				}
			}
		}
		c.addBlock(cblk)
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

func (b *simplemultiblk) InputNames() []string { return b.names }

func (b *simplemultiblk) SetInput(sel string, input block.Port) error {
	for i, n := range b.names {
		if n == sel {
			return b.v[i].(block.InputSetter).SetInput("", input)
		}
	}
	return errf("simplemultiblk has no input '%s'", sel)
}

func (b *simplemultiblk) OutputNames() []string { return b.names }

func (b *simplemultiblk) Output(sel string) (block.Port, error) {
	for i, n := range b.names {
		if n == sel {
			return b.v[i].Output("")
		}
	}
	return nil, errf("simplemultiblk has no output '%s'", sel)
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

func (b *constblkspec) OutputNames() []string { return []string{""} }

func (b *constblkspec) Output(sel string) (block.Port, error) {
	if sel != "" {
		return nil, srcerrf(b, "const has no named port %s", sel)
	}
	return b.p, nil
}

func (*constblkspec) sourceline() int { return -1 }

type valueblkspec struct {
	constblkspec
	lineno int
}

func (b *valueblkspec) sourceline() int { return b.lineno }

type namedblkspec struct {
	lineno    int
	name, sel string
}

func (b *namedblkspec) sourceline() int { return b.lineno }
func (b *namedblkspec) Prepare(c *context) (block.Block, error) {
	blks, ok := c.specs[b.name]
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

func (p *namedport) sourceline() int       { return p.lineno }
func (p *namedport) OutputNames() []string { return []string{""} }
func (p *namedport) Output(sel string) (block.Port, error) {
	if sel != "" {
		return nil, srcerrf(p, "internal error: selector '%s' on namedblkspec is impossible", sel)
	}
	return p.blk.Output(p.sel)
}

type connspec struct {
	name, sel string
	blockspec

	lineno int
}
