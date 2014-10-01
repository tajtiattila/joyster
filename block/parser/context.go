package parser

import ()

type Context struct {
	TypeMap
	Config      map[string]float64
	PortNames   map[string]SpecSource
	sinkNames   map[string]*Blk
	sourceNames map[string]*Blk
	vblk        []*Blk
	vlink       []Link
}

func (c *Context) dependency(cons, prod *Blk) error {
	// TODO
	return nil
}

type BlkSpec interface {
	String() string
	SrcLine() int
	OutputNames() []string
	InBlk() *Blk
	OutBlk() *Blk
}

// Type can tell what inputs and outputs its block has
type Type interface {
	InputNames() []string
	OutputNames() []string
	MustHaveInput() bool
}

// Namespace knows which named block types are available
type TypeMap interface {
	GetType(n string) (Type, error)
}

type Link struct {
	sink   SpecSink
	source SpecSource
}

func (l *Link) markdep(c *Context) error {
	consumer, err := l.sink.Blk(c)
	if err != nil {
		return err
	}
	producer, err := l.source.Blk(c)
	if err != nil {
		return err
	}
	if consumer != nil && producer != nil {
		return c.dependency(consumer, producer)
	}
	return nil
}

func (l *Link) setup(c *Context) error {
	src, err := l.source.Source(c)
	if err != nil {
		return err
	}
	return l.sink.SetTo(c, src)
}

type SpecSink interface {
	Blk(c *Context) (*Blk, error)
	SetTo(c *Context, src Source) error
}

type SpecSource interface {
	Blk(c *Context) (*Blk, error)
	Source(c *Context) (Source, error)
}
