package block

import (
	"errors"
	"fmt"
	"github.com/tajtiattila/joyster/block/parser"
)

// hat values are bitmasks
const (
	HatCentre = 0

	HatNorth = 1
	HatEast  = 2
	HatSouth = 4
	HatWest  = 8

	HatMask = 15
	HatMax  = 16
)

type PortType parser.PortType

const (
	Invalid = PortType(parser.Invalid)
	Bool    = PortType(parser.Bool)
	Float64 = PortType(parser.Scalar)
	Int     = PortType(parser.Hat)
	Any     = PortType(parser.Any)
)

func TypeOf(port Port) PortType {
	switch port.(type) {
	case *bool:
		return Bool
	case *float64:
		return Float64
	case *int:
		return Int
	}
	return Invalid
}

func PtrTypeOf(pport interface{}) PortType {
	switch pport.(type) {
	case **bool:
		return Bool
	case **float64:
		return Float64
	case **int:
		return Int
	}
	return Invalid
}

// ZeroValue creates a zero value for a port type. Only concrete
// types are supported, for Invalid and Any it panics.
func ZeroValue(t PortType) Port {
	switch t {
	case Bool:
		return boolPortVal
	case Float64:
		return scalarPortVal
	case Int:
		return hatPortVal
	}
	panic("invalid input for ZeroValue")
}

type PortTypeMap map[string]PortType

type IO interface {
	Names() []string
	Value(sel string) interface{}
}

// TypeInputMap specifies order of inputs and their types
// for a Block Type. It is an error to call Type() with sel not in Names().
type TypeInputMap interface {
	IO
	Type(sel string) PortType
}

type InputMap interface {
	TypeInputMap
	Set(sel string, port Port) error
}

type OutputMap interface {
	IO
	Get(sel string) (Port, error)
}

// Port is one of:
//  *float64 - axis
//  *bool - button, flag
//  *int - hat
type Port interface{}

func CheckInputs(typ string, v ...interface{}) error {
	for i, input := range v {
		if err := CheckInput(input); err != nil {
			return fmt.Errorf("'%s' input %d: %s", typ, i+1, err.Error())
		}
	}
	return nil
}

func CheckInput(i interface{}) error {
	if i == nil {
		return errors.New("port is nil")
	}
	switch x := i.(type) {
	case **bool:
		if *x != nil {
			return nil
		}
	case **float64:
		if *x != nil {
			return nil
		}
	case **int:
		if *x != nil {
			return nil
		}
	default:
		return errors.New("port type invalid")
	}
	return errors.New("port uninitialized")
}

func CheckOutput(p Port) error {
	if p == nil {
		return errors.New("port is nil")
	}
	switch x := p.(type) {
	case *bool:
		if x != nil {
			return nil
		}
	case *float64:
		if x != nil {
			return nil
		}
	case *int:
		if x != nil {
			return nil
		}
	default:
		return errors.New("port type invalid")
	}
	return errors.New("port uninitialized")
}

func PortString(p Port) string {
	if p == nil {
		return "uninitialized"
	}
	ret := "nullptr"
	switch x := p.(type) {
	case *bool:
		if x != nil {
			return "bool"
		}
	case *float64:
		if x != nil {
			return "axis"
		}
	case *int:
		if x != nil {
			return "hat"
		}
	default:
		ret = "invalid"
	}
	return ret
}

func Connect(i interface{}, port Port) error {
	var (
		my Port
		ok bool
	)
	switch x := i.(type) {
	case **bool:
		my = *x
		*x, ok = port.(*bool)
	case **float64:
		my = *x
		*x, ok = port.(*float64)
	case **int:
		my = *x
		*x, ok = port.(*int)
	default:
		return errors.New("target type invalid")
	}
	if !ok {
		return fmt.Errorf("type %s and %s incompatible", PortString(my), PortString(port))
	}
	return nil
}

var (
	boolPortVal   = new(bool)
	scalarPortVal = new(float64)
	hatPortVal    = new(int)
)
