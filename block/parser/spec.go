package parser

import (
	"fmt"
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

func (b *factoryblkspec) String() string {
	return fmt.Sprintf("factory:%s@%d", b.typ, b.lineno)
}

func (b *factoryblkspec) sourceline() int { return b.lineno }

func (b *factoryblkspec) Deps(c *Context) error {
	for _, d := range b.inputs {
		c.dependency(b, d)
		d.Deps(c)
	}
	return nil
}

func (b *factoryblkspec) Prepare(c *Context) (block.Block, error) {
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

func (b *factoryblkspec) create(c *Context, grp bool) (block.Block, error) {
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
		if xblk.Input() == nil || yblk.Input() == nil {
			return nil, srcerrf(b, "$ block '%s' must support inputs", b.typ)
		}
		return &xyblk{xblk, yblk}, nil
	}
	return blk, nil
}

func (b *factoryblkspec) inputsetup(c *Context, blk block.Block) error {
	if len(b.inputs) != 0 {
		im := blk.Input()
		if im == nil {
			return srcerrf(b, "block type '%s' doesn't accept inputs", b.typ)
		}
		names := im.Names()
		if len(b.inputs) != len(names) {
			return srcerrf(b, "block type '%s' needs exactly %d inputs, not %d", b.typ, len(names), len(b.inputs))
		}
		for i, input := range b.inputs {
			cblk, err := input.Prepare(c)
			if err != nil {
				return srcerr(b, err)
			}
			om := cblk.Output()
			if om == nil {
				return srcerrf(b, "child has no output")
			}
			port, err := om.Get("")
			if err != nil {
				return srcerr(b, err)
			}
			err = im.Set(names[i], port)
			if err != nil {
				return srcerr(b, err)
			}
		}
	}
	return nil
}

func (b *factoryblkspec) newBlock(c *Context) (block.Block, error) {
	param := b.param.Prepare(c)
	blk, err := c.createBlock(b.typ, param)
	if err != nil {
		return nil, srcerr(b, err)
	}
	if param.err() != nil {
		return nil, srcerr(b, param.err())
	}
	c.blockline[blk] = b.sourceline()
	return blk, nil
}

type xyblk struct {
	x, y block.Block
}

func (b *xyblk) Input() block.InputMap { return &xyio{b} }

func (b *xyblk) Output() block.OutputMap { return &xyio{b} }
func (b *xyblk) Validate() error {
	if err := b.x.Validate(); err != nil {
		return err
	}
	return b.y.Validate()
}

type xyio struct {
	b *xyblk
}

func (xy *xyio) Names() []string { return []string{"x", "y"} }

func (xy *xyio) Get(sel string) (block.Port, error) {
	switch sel {
	case "x":
		return xy.b.x.Output().Get("")
	case "y":
		return xy.b.y.Output().Get("")
	}
	return nil, errf("'xy' has inputs 'x' and 'y', but not '%s'", sel)
}

func (xy *xyio) Set(sel string, port block.Port) error {
	var child block.Block
	switch sel {
	case "x":
		child = xy.b.x
	case "y":
		child = xy.b.y
	default:
		return errf("'xy' has inputs 'x' and 'y', but not '%s'", sel)
	}
	return child.Input().Set("", port)
}

type groupblkspec struct {
	lineno int
	sels   []string
	v      []*factoryblkspec
}

func (b *groupblkspec) String() string {
	return fmt.Sprintf("group@%d", b.lineno)
}

func (b *groupblkspec) sourceline() int { return b.lineno }

func (b *groupblkspec) Deps(c *Context) error {
	var last blockspec
	for _, d := range b.v {
		if last != nil {
			c.dependency(d, last)
		}
		last = d
	}
	c.dependency(b, last)
	return nil
}

func (b *groupblkspec) Prepare(c *Context) (block.Block, error) {
	var (
		first block.InputMap
		last  block.OutputMap
		blks  []block.Block
	)
	for _, cblks := range b.v {
		cblk, err := cblks.create(c, true)
		if err != nil {
			return nil, srcerr(b, err)
		}
		im, om := cblk.Input(), cblk.Output()
		if im == nil {
			return nil, srcerrf(cblks, "block type '%s' doesn't support inputs", cblks.typ)
		}
		if om == nil {
			return nil, srcerrf(cblks, "block type '%s' doesn't support outputs", cblks.typ)
		}
		sels := b.sels
		if len(sels) == 0 {
			if !has(im.Names(), "") {
				return nil, srcerrf(cblks, "block type '%s' has no unnamed input", cblks.typ)
			}
			if !has(om.Names(), "") {
				return nil, srcerrf(cblks, "block type '%s' has no unnamed output", cblks.typ)
			}
			sels = []string{""}
		} else {
			for _, n := range sels {
				if !has(im.Names(), n) {
					return nil, srcerrf(cblks, "block type '%s' has no input '%s'", cblks.typ, n)
				}
				if !has(om.Names(), n) {
					return nil, srcerrf(cblks, "block type '%s' has no output '%s'", cblks.typ, n)
				}
			}
		}
		c.addBlock(cblk)
		blks = append(blks, cblk)
		if first == nil {
			first = im
		} else {
			for _, n := range sels {
				p, err := last.Get(n)
				if err != nil {
					return nil, srcerr(b, err)
				}
				err = im.Set(n, p)
				if err != nil {
					return nil, srcerr(b, err)
				}
			}
		}
		last = om
	}
	return &groupblk{blks, first, last}, nil
}

type groupblk struct {
	blks  []block.Block
	first block.InputMap
	last  block.OutputMap
}

func (g *groupblk) Input() block.InputMap   { return g.first }
func (g *groupblk) Output() block.OutputMap { return g.last }
func (g *groupblk) Validate() error {
	for _, c := range g.blks {
		if err := c.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type constblkspec struct {
	p block.Port
}

func (b *constblkspec) String() string {
	return fmt.Sprintf("const:%T=%v", b.p, b.p)
}

func (b *constblkspec) Deps(c *Context) error                   { return nil }
func (b *constblkspec) Prepare(c *Context) (block.Block, error) { return b, nil }

func (b *constblkspec) Input() block.InputMap   { return nil }
func (b *constblkspec) Output() block.OutputMap { return block.SingleOutput("const", b.p) }
func (b *constblkspec) Validate() error         { return nil }

func (*constblkspec) sourceline() int { return -1 }

type valueblkspec struct {
	constblkspec
	lineno int
}

func (b *valueblkspec) String() string {
	return fmt.Sprintf("value@%d:%T=%v", b.lineno, b.p, b.p)
}

func (b *valueblkspec) sourceline() int { return b.lineno }

type namedblkspec struct {
	lineno    int
	name, sel string
}

func (b *namedblkspec) String() string {
	var sel string
	if b.sel != "" {
		sel = "." + b.sel
	}
	return fmt.Sprintf("named@%d:%s%s", b.lineno, b.name, sel)
}

func (b *namedblkspec) sourceline() int { return b.lineno }
func (b *namedblkspec) Deps(c *Context) error {
	d, ok := c.names[b.name]
	if !ok {
		return srcerrf(b, "input %s is missing", b.name)
	}
	c.dependency(b, d)
	return nil
}

func (b *namedblkspec) Prepare(c *Context) (block.Block, error) {
	blks, ok := c.names[b.name]
	if !ok {
		return nil, srcerrf(b, "input %s is missing", b.name)
	}
	blk, err := blks.Prepare(c)
	if err != nil {
		return nil, srcerr(b, err)
	}
	om := blk.Output()
	if om == nil {
		return nil, srcerrf(b, "input %s does not have output", b.name)
	}
	port, err := om.Get(b.sel)
	if err != nil {
		return nil, srcerr(b, err)
	}
	return &namedport{block.SingleOutput(b.name+"."+b.sel, port)}, nil
}

type namedport struct{ m block.OutputMap }

func (p *namedport) Input() block.InputMap   { return nil }
func (p *namedport) Output() block.OutputMap { return p.m }
func (p *namedport) Validate() error         { return nil }

type connspec struct {
	name, sel string
	blockspec

	lineno int
}

func has(v []string, s string) bool {
	for _, w := range v {
		if s == w {
			return true
		}
	}
	return false
}
