package parser

import (
	"fmt"
)

type factory struct {
	tname string
	typ   Type
	param Param
}

type dollarPortMapper struct {
	lno int
	m   map[string]*Blk
}

func (m *dollarPortMapper) port(sel string) (*Blk, string, error) {
	if blk, ok := m.m[sel]; ok {
		return blk, "", nil
	}
	return nil, "", errf("$ %s does not exist", nice(port, sel))
}

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

func constint(v int) *constport   { return &constport{v} }
func constbool(b bool) *constport { return &constport{b} }

func (*constport) SrcLine() int                 { return -1 }
func (p *constport) Blk(*context) (*Blk, error) { return nil, nil }
func (p *constport) Source(*context) (Source, error) {
	s, err := Value(p.v)
	if err != nil {
		return nil, srcerr(p, err)
	}
	return s, nil
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

func (k *concreteblksink) SrcLine() int                     { return k.lno }
func (k *concreteblksink) Blk(*context) (*Blk, error)       { return k.blk, nil }
func (k *concreteblksink) SetTo(c *context, s Source) error { return k.blk.SetInput(k.sel, s) }

type concreteblksource struct {
	lno int
	blk *Blk
	sel string
}

func (e *concreteblksource) SrcLine() int                 { return e.lno }
func (e *concreteblksource) Blk(c *context) (*Blk, error) { return e.blk, nil }
func (e *concreteblksource) Source(c *context) (Source, error) {
	return &BlkPortSource{e.blk, e.sel}, nil
}

func has(v []string, s string) bool {
	for _, w := range v {
		if s == w {
			return true
		}
	}
	return false
}

type named struct {
	lno  int
	name string
	sel  string
}

func (n *named) SrcLine() int { return n.lno }

func (n *named) String() string {
	var sel string
	if n.sel != "" {
		sel = "." + n.sel
	}
	return fmt.Sprintf("named@%d:%s%s", n.lno, n.name, sel)
}

func (n *named) resolve(names portMap) (*Blk, string, error) {
	if pm, ok := names[n.name]; ok {
		blk, sel, err := pm.port(n.sel)
		if err != nil {
			return nil, "", err
		}
		return blk, sel, nil
	}
	return nil, "", srcerrf(n, "block '%s' missing", n.name)
}

func nsink(lno int, name string, sel string) *namedsink {
	return &namedsink{named{lno, name, sel}}
}

type namedsink struct {
	named
}

func (k *namedsink) Blk(c *context) (*Blk, error) {
	blk, _, err := k.resolve(c.sinkNames)
	return blk, err
}

func (k *namedsink) SetTo(c *context, s Source) error {
	blk, sel, err := k.resolve(c.sinkNames)
	if err != nil {
		return err
	}
	return blk.SetInput(sel, s)
}

func nsource(lno int, name string, sel string) *namedsource {
	return &namedsource{named{lno, name, sel}}
}

type namedsource struct {
	named
}

func (e *namedsource) Blk(c *context) (*Blk, error) {
	if _, ok := c.portNames[e.name]; ok {
		return nil, nil
	}
	blk, _, err := e.resolve(c.sourceNames)
	return blk, err
}

func (e *namedsource) Source(c *context) (Source, error) {
	if p, ok := c.portNames[e.name]; ok {
		return p.Source(c)
	}
	blk, sel, err := e.resolve(c.sourceNames)
	if err != nil {
		return nil, err
	}
	return &BlkPortSource{blk, sel}, nil
}

type outputconstraint struct {
	reason string
	sels   []string
}
