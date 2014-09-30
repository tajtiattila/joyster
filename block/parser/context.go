package parser

import (
	"fmt"
	"github.com/tajtiattila/joyster/block"
)

type Context struct {
	*block.Profile

	config map[string]float64
	names  map[string]blockspec
	conns  []connspec
	deps   map[blockspec][]blockspec

	blockline map[block.Block]int
}

func (c *Context) createBlock(typ string, p block.Param) (block.Block, error) {
	if f, ok := block.DefaultTypeMap[typ]; ok {
		blk, err := f(p)
		if err != nil {
			return nil, err
		}
		c.Blocks = append(c.Blocks, blk)
		return blk, nil
	}
	return nil, fmt.Errorf("unknown type '%s'", typ)
}

func (c *Context) addBlock(b block.Block) {
	if t, ok := b.(block.Ticker); ok {
		c.Tickers = append(c.Tickers, t)
	}
}

func (c *Context) dependency(spec, dependency blockspec) {
	if c.deps == nil {
		c.deps = make(map[blockspec][]blockspec)
	}
	for _, d := range c.deps[spec] {
		if d == dependency {
			return
		}
	}
	fmt.Printf("Dependency: %v → %v\n", spec, dependency)
	c.deps[spec] = append(c.deps[spec], dependency)
}

type blockspec interface {
	Deps(c *Context) error
	Prepare(c *Context) (block.Block, error)
	String() string
}

type dummyBlock struct {
	name string
	v    bool
}

func (b *dummyBlock) OutputNames() []string                      { return []string{""} }
func (b *dummyBlock) Output(sel string) (block.Port, error)      { return &b.v, nil }
func (b *dummyBlock) InputNames() []string                       { return []string{""} }
func (b *dummyBlock) SetInput(sel string, port block.Port) error { return nil }
