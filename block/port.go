package block

import (
	"errors"
	"fmt"
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

type InputMap interface {
	Names() []string
	Set(sel string, port Port) error
}

type OutputMap interface {
	Names() []string
	Get(sel string) (Port, error)
}

// Port is one of:
//  *float64 - axis
//  *bool - button, flag
//  *int - hat
type Port interface{}

func CheckInput(i interface{}) error {
	if i == nil {
		return errors.New("port is nil")
	}
	switch x := i.(type) {
	case **bool:
		if x != nil {
			return nil
		}
	case **float64:
		if x != nil {
			return nil
		}
	case **int:
		if x != nil {
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
		return fmt.Errorf("types %s and %s incompatible", PortString(my), PortString(port))
	}
	return nil
}
