package block

import (
	"fmt"
)

var varArgNames = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}

type varArgInput struct {
	title string

	a portArray
}

func VarArgInput(t string, iv interface{}) *varArgInput {
	a, err := makePortArray(iv)
	if err != nil {
		panic(err)
	}
	return &varArgInput{t, a}
}

func (*varArgInput) Names() []string { return varArgNames }

func (v *varArgInput) Value(sel string) interface{} {
	i, ok := selidx(sel)
	if !ok || v.a.Len() <= i {
		return nil
	}
	return pval(v.a.Port(i))
}

func (v *varArgInput) Set(sel string, port Port) error {
	i, ok := selidx(sel)
	if !ok {
		return fmt.Errorf("'%s' has no input named '%s'", v.title, sel)
	}
	if v.a.Len() <= i {
		v.a.SetLen(i + 1)
	}
	if err := Connect(v.a.Port(i), port); err != nil {
		return fmt.Errorf("cant connect '%s' port '%s': %s", v.title, sel, err.Error())
	}
	return nil
}

func (v *varArgInput) Type(sel string) PortType {
	if _, ok := selidx(sel); !ok {
		return Invalid
	}
	return v.a.PortType()
}

func VarArgCheck(t string, iv interface{}, min int) error {
	a, err := makePortArray(iv)
	if err != nil {
		return err
	}
	if a.Len() < min {
		return fmt.Errorf("'%s' has only %d input(s) but needs %d", t, a.Len(), min)
	}
	for i := 0; i < a.Len(); i++ {
		pi := a.Port(i)
		var isnil, isinvalid bool
		switch p := pi.(type) {
		case **bool:
			isnil = *p == nil
		case **float64:
			isnil = *p == nil
		case **int:
			isnil = *p == nil
		case nil:
			panic("impossible")
		default:
			isinvalid = true
		}
		if isnil {
			return fmt.Errorf("'%s' port '%d' is nil", t, i+1)
		}
		if isinvalid {
			return fmt.Errorf("'%s' port '%d' has invalid type %T", t, i+1, pi)
		}
	}
	return nil
}

func selidx(s string) (int, bool) {
	n := 0
	for _, ch := range s {
		digit := int(ch - '0')
		if digit < 0 || digit > 9 || (digit == 0 && n == 0) {
			return 0, false
		}
		n = n*10 + digit
	}
	if n == 0 {
		return 0, false
	}
	return n - 1, true
}

type portArray interface {
	Len() int
	SetLen(int)
	Port(int) interface{}
	PortType() PortType
}

func makePortArray(iv interface{}) (portArray, error) {
	switch p := iv.(type) {
	case *[]*bool:
		return (*boolPortArray)(p), nil
	case *[]*float64:
		return (*float64PortArray)(p), nil
	case *[]*int:
		return (*intPortArray)(p), nil
	}
	return nil, fmt.Errorf("invalid type %T for portArray", iv)
}

type boolPortArray []*bool

func (v *boolPortArray) PortType() PortType     { return Bool }
func (v *boolPortArray) Len() int               { return len(*v) }
func (v *boolPortArray) Port(i int) interface{} { return &((*v)[i]) }

func (v *boolPortArray) SetLen(n int) {
	if cap(*v) < n {
		*v = append(make([]*bool, 0, nextlen(cap(*v))), (*v)...)
	}
	*v = (*v)[:n]
}

type float64PortArray []*float64

func (v *float64PortArray) PortType() PortType     { return Float64 }
func (v *float64PortArray) Len() int               { return len(*v) }
func (v *float64PortArray) Port(i int) interface{} { return &((*v)[i]) }

func (v *float64PortArray) SetLen(n int) {
	if cap(*v) < n {
		*v = append(make([]*float64, 0, nextlen(cap(*v))), (*v)...)
	}
	*v = (*v)[:n]
}

type intPortArray []*int

func (v *intPortArray) PortType() PortType     { return Int }
func (v *intPortArray) Len() int               { return len(*v) }
func (v *intPortArray) Port(i int) interface{} { return &((*v)[i]) }

func (v *intPortArray) SetLen(n int) {
	if cap(*v) < n {
		*v = append(make([]*int, 0, nextlen(cap(*v))), (*v)...)
	}
	*v = (*v)[:n]
}

func nextlen(oldcap int) int {
	if oldcap < 16 {
		return 16
	}
	return 2 * oldcap
}
