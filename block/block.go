package block

import ()

// a block is a set of outputs
type Block interface {
	OutputNames() []string
	Output(sel string) (Port, error)
}

// one of:
//  *float64 // axis
//  *bool // button, flag
//  *int // hat
type Port interface{}

func PortString(p Port) string {
	if p == nil {
		return "uninitialized"
	}
	switch p.(type) {
	case *bool:
		return "bool"
	case *float64:
		return "axis"
	case *int:
		return "hat"
	}
	return "invalid"
}

type Setupper interface {
	Setup(*Param) error
}

// ticker is an interface for blocks that needs update once upon each update tick.
type Ticker interface {
	Tick()
}

// InputSetter is implementet by blocks with input(s) can set its inputs independently
type InputSetter interface {
	NMinInput() int       // minimal number of inputs, must be smaller or equal to len(InputNames())
	InputNames() []string // names of all inputs. Mandatory first
	SetInput(sel string, port Port) error
}

type Param struct {
	P []float64
	N map[string]float64
}

type TypeMap map[string]func() Block

var DefaultTypeMap = make(TypeMap)

func Register(name string, fn func() Block) {
	if _, ok := DefaultTypeMap[name]; ok {
		panic("duplicate name: " + name)
	}
	DefaultTypeMap[name] = fn
}
