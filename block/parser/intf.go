package parser

import (
	"io"
	"io/ioutil"
)

// Profile is a set of Blks ordered.
// Block is sorted such that Blks appear before their dependencies.
type Profile struct {
	Config NamedParam // parameters from set statements
	Blocks []*Blk     // slice of all Blks found in source
}

func ReadProfile(r io.Reader, tm TypeMap) (*Profile, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return read(data, tm)
}

func LoadProfile(fn string, tm TypeMap) (*Profile, error) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	return read(data, tm)
}

func read(data []byte, tm TypeMap) (*Profile, error) {
	p := newparser(tm)
	if err := p.parse(data); err != nil {
		return nil, err
	}
	if err := sort(p.context); err != nil {
		return nil, err
	}
	return &Profile{p.context.config, p.context.vblk}, nil
}

// Blk is the working unit in a Profile.
type Blk struct {
	Name   string
	Type   Type
	Param  Param
	Inputs map[string]Source
}

func (b *Blk) SetInput(name string, src Source) error {
	if b.Inputs == nil {
		b.Inputs = make(map[string]Source)
	}
	if old, ok := b.Inputs[name]; ok && old != src {
		return errf("block '%s' has %s already set", b.Name, nice(inport, name))
	}
	b.Inputs[name] = src
	return nil
}

func (b *Blk) port(sel string) (*Blk, string, error) { return b, sel, nil }

// Source represents input ports for Blk. It can be either a BlkPortSource that refers
// to an output port of another Blk, or a fixed value represented by ValueSource.
type Source interface{}

// BlkPortSource is a reference to an output of another Block in the same Profile.
type BlkPortSource struct {
	Blk *Blk
	Sel string
}

// ValueSource is a constant source. Value can be bool, int or float64.
type ValueSource struct {
	Value interface{}
}

// Type represents the type Blks, and what inputs and outputs its block has.
// Blocks must normally have all their inputs connected, except if MustHaveInput reports false.
type Type interface {
	InputNames() []string
	OutputNames() []string
	MustHaveInput() bool
}

// Namespace knows the types available for a Profile.
type TypeMap interface {
	GetType(n string) (Type, error)
}
