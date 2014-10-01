package parser

import (
	"fmt"
)

type factory struct {
	tname string
	typ   Type
	param Param
}

type factoryblkspec struct {
	lineno int
	factory
	inputs []SpecSource

	blk *Blk
}

func (b *factoryblkspec) String() string {
	return fmt.Sprintf("factory:%s@%d", b.tname, b.lineno)
}

func (b *factoryblkspec) InputNames() []string  { return b.typ.InputNames() }
func (b *factoryblkspec) OutputNames() []string { return b.typ.OutputNames() }
func (b *factoryblkspec) SrcLine() int          { return b.lineno }

func (b *factoryblkspec) InBlk() *Blk  { return b.blk }
func (b *factoryblkspec) OutBlk() *Blk { return b.blk }

type groupblkspec struct {
	lineno int
	sels   []string
	v      []grpchild
}

type grpchild struct {
	dollar bool
	factory
}

func (b *groupblkspec) String() string {
	return fmt.Sprintf("group@%d", b.lineno)
}

func (b *groupblkspec) SrcLine() int { return b.lineno }

type constport struct {
	v interface{}
}

func (*constport) SrcLine() int                 { return -1 }
func (p *constport) Blk(*Context) (*Blk, error) { return nil, nil }
func (p *constport) Source(*Context) (Source, error) {
	return &ValueSource{p.v}, nil
}

func (b *constport) String() string {
	return fmt.Sprintf("const:%T=%v", b.v, b.v)
}

type valueport struct {
	lineno int
	constport
}

func (b *valueport) String() string {
	return fmt.Sprintf("value@%d:%T=%v", b.lineno, b.v, b.v)
}

func (b *valueport) SrcLine() int { return b.lineno }

type concreteblksink struct {
	lno int
	blk *Blk
	sel string
}

func (k *concreteblksink) SrcLine() int               { return k.lno }
func (k *concreteblksink) Blk(*Context) (*Blk, error) { return k.blk, nil }
func (k *concreteblksink) SetTo(c *Context, s Source) error {
	return k.blk.SetInput(k.sel, s)
}

type namedsink struct {
	lno  int
	name string
	sel  string
}

func (k *namedsink) SrcLine() int { return k.lno }
func (k *namedsink) Blk(c *Context) (*Blk, error) {
	if blk, ok := c.sinkNames[k.name]; ok {
		return blk, nil
	}
	return nil, errf("block '%s' missing", k.name)
}

func (k *namedsink) SetTo(c *Context, s Source) error {
	blk, err := k.Blk(c)
	if err != nil {
		return err
	}
	return blk.SetInput(k.sel, s)
}

type connspec struct {
	name, sel string // target
	Source

	lineno int
}

type concreteblksource struct {
	lno int
	blk *Blk
	sel string
}

func (e *concreteblksource) SrcLine() int                 { return e.lno }
func (e *concreteblksource) Blk(c *Context) (*Blk, error) { return e.blk, nil }
func (e *concreteblksource) Source(c *Context) (Source, error) {
	return BlkPortSource{e.blk, e.sel}, nil
}

func has(v []string, s string) bool {
	for _, w := range v {
		if s == w {
			return true
		}
	}
	return false
}

type namedsource struct {
	lineno    int
	name, sel string
}

func (e *namedsource) String() string {
	var sel string
	if e.sel != "" {
		sel = "." + e.sel
	}
	return fmt.Sprintf("named@%d:%s%s", e.lineno, e.name, sel)
}

func (e *namedsource) SrcLine() int { return e.lineno }
func (e *namedsource) Blk(c *Context) (*Blk, error) {
	if _, ok := c.PortNames[e.name]; ok {
		return nil, nil
	}
	if blk, ok := c.sourceNames[e.name]; ok {
		return blk, nil
	}
	return nil, errf("block '%s' missing", e.name)
}

func (e *namedsource) Source(c *Context) (Source, error) {
	if p, ok := c.PortNames[e.name]; ok {
		return p.Source(c)
	}
	blk, err := e.Blk(c)
	if err != nil {
		return nil, err
	}
	return BlkPortSource{blk, e.sel}, nil
}
