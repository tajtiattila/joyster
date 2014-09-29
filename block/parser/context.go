package parser

import (
	"github.com/tajtiattila/joyster/block"
)

type context struct {
	config  map[string]float64
	typemap map[string](func() block.Block)
	blocks  map[string]blockspec
	conns   []connspec
	deps    map[block.Block][]block.Block
}

func (c *context) createBlock(typ string, p *block.Param) (block.Block, error) {
	if f, ok := block.DefaultTypeMap[typ]; ok {
		return f(p)
	}
	/*
		f, ok := c.typemap[typ]
		if !ok {
			return nil, fmt.Errorf("block type '%s' unknown", typ)
		}
		b.xblk = f()
	*/
	return &dummyBlock{typ, false}, nil
}

func (c *context) depends(blk, dependency block.Block) {
	if c.deps == nil {
		c.deps = make(map[block.Block][]block.Block)
	}
	c.deps[blk] = append(c.deps[blk], dependency)
}

type blockspec interface {
	Prepare(c *context) (block.Block, error)
}

type dummyBlock struct {
	name string
	v    bool
}

func (b *dummyBlock) OutputNames() []string                      { return []string{""} }
func (b *dummyBlock) Output(sel string) (block.Port, error)      { return &b.v, nil }
func (b *dummyBlock) InputNames() []string                       { return []string{""} }
func (b *dummyBlock) SetInput(sel string, port block.Port) error { return nil }
