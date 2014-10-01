package parser

import ()

type Context struct {
	TypeMap
	Config map[string]float64
	Names  map[string]BlkSpec
	Refs   []PortRef
	conns  []connspec
}

type PortRef interface {
	String() string
	Ref() (BlkSpec, string, error)
}

type Spec interface {
	String() string
	SrcLine() int
}

type BlkSpec interface {
	Spec
}

type FactoryBlkSpec interface {
	BlkSpec
	InputNames() []string
	OutputNames() []string
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
