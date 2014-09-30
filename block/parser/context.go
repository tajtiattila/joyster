package parser

import (
	"fmt"
	"github.com/tajtiattila/joyster/block"
)

type context struct {
	config  map[string]float64
	typemap map[string](func() block.Block)
	specs   map[string]blockspec
	conns   []connspec
	tickers []block.Ticker
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
	return nil, fmt.Errorf("unknown type '%s'", typ)
	//return &dummyBlock{typ, false}, nil
}

func (c *context) addBlock(b block.Block) {
	if t, ok := b.(block.Ticker); ok {
		c.tickers = append(c.tickers, t)
	}
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
