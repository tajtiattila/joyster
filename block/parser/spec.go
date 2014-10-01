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
	inputs []BlkSpec
}

func (b *factoryblkspec) String() string {
	return fmt.Sprintf("factory:%s@%d", b.tname, b.lineno)
}

func (b *factoryblkspec) SrcLine() int { return b.lineno }

func (b *factoryblkspec) InputNames() []string  { return b.typ.InputNames() }
func (b *factoryblkspec) OutputNames() []string { return b.typ.InputNames() }

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

type constblkspec struct {
	p interface{}
}

func (*constblkspec) SrcLine() int { return -1 }

func (b *constblkspec) String() string {
	return fmt.Sprintf("const:%T=%v", b.p, b.p)
}

type valueblkspec struct {
	constblkspec
	lineno int
}

func (b *valueblkspec) String() string {
	return fmt.Sprintf("value@%d:%T=%v", b.lineno, b.p, b.p)
}

func (b *valueblkspec) SrcLine() int { return b.lineno }

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

func (b *namedblkspec) SrcLine() int { return b.lineno }

type connspec struct {
	name, sel string // target
	BlkSpec

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
