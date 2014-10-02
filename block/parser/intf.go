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

func Parse(src string, tm TypeMap) (*Profile, error) {
	data := []byte(src)
	return read(data, tm)
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

type PortType int

const (
	Invalid PortType = iota
	Bool
	Scalar
	Hat
	Any
)

func PortStr(p PortType) string {
	switch p {
	case Bool:
		return "bool"
	case Scalar:
		return "scalar"
	case Hat:
		return "hat"
	case Any:
		return "any"
	}
	return "invalid"
}

func Match(a, b PortType) bool {
	if a == Invalid || b == Invalid {
		return false
	}
	if a == Any || b == Any {
		return true
	}
	return a == b
}

type PortMap []Port

func (m PortMap) Port(name string) PortType {
	for _, p := range m {
		if p.Name == name {
			return p.Type
		}
	}
	return Invalid
}

func (m PortMap) Names() []string {
	n := make([]string, len(m))
	for i, p := range m {
		n[i] = p.Name
	}
	return n
}

type Port struct {
	Name string
	Type PortType
}

// Source represents input ports for Blk. It can be either a BlkPortSource that refers
// to an output port of another Blk, or a fixed value represented by ValueSource.
type Source interface {
	Type() PortType
}

// BlkPortSource is a reference to an output of another Block in the same Profile.
type BlkPortSource struct {
	Blk *Blk
	Sel string
}

func (s *BlkPortSource) Type() PortType {
	var pm PortMap
	for _, p := range s.Blk.Type.Input() {
		var pt PortType
		if input := s.Blk.Inputs[p.Name]; input != nil {
			pt = input.Type()
		}
		pm = append(pm, Port{p.Name, pt})
	}
	pm = s.Blk.Type.Output(pm)
	return pm.Port(s.Sel)
}

// ValueSource is a constant source. Value can be bool, int or float64.
type ValueSource struct {
	Value interface{}
}

func Value(i interface{}) (Source, error) {
	switch i.(type) {
	case bool, int, float64:
		return &ValueSource{i}, nil
	}
	return nil, errf("invalid value: %#v", i)
}

func (s *ValueSource) Type() PortType {
	switch s.Value.(type) {
	case bool:
		return Bool
	case int:
		return Hat
	case float64:
		return Scalar
	}
	return Invalid
}

// Type represents the type Blks, and what inputs and outputs its block has.
// Blocks must normally have all their inputs connected, except if MustHaveInput of their Type
// returns false.
type Type interface {
	Input() PortMap                          // inputs needed for Blks of this Type
	Output(input PortMap) PortMap            // outputs provided by Blks of this Type for given input
	MustHaveInput() bool                     // tells wether all inputs must be set for this type
	Param(p Param, globals NamedParam) error // validate input parameters
}

// Namespace knows the types available for a Profile.
type TypeMap interface {
	GetType(n string) (Type, error)
}
