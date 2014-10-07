package block

import (
	"fmt"
)

func MapInput(t string, v ...MapDecl) InputMap {
	n, m := make([]string, 0, len(v)), make(map[string]interface{})
	for _, p := range v {
		switch p.V.(type) {
		case **bool, **float64, **int:
		default:
			panic("invalid type in MapInput")
		}
		n, m[p.N] = append(n, p.N), p.V
	}
	return &mapInput{t, n, m}
}

func MapOutput(t string, v ...MapDecl) OutputMap {
	n, m := make([]string, 0, len(v)), make(map[string]interface{})
	for _, p := range v {
		switch p.V.(type) {
		case *bool, *float64, *int:
		default:
			panic("invalid type in MapOutput")
		}
		n, m[p.N] = append(n, p.N), p.V
	}
	return &mapOutput{t, n, m}
}

type MapDecl struct {
	N string
	V interface{}
}

func pt(n string, v interface{}) MapDecl { return MapDecl{n, v} }

func SingleOutput(t string, p Port) OutputMap { return &singleOutput{t, p} }

func SingleInput(t string, i interface{}) InputMap { return &singleInput{t, i} }

type mapInput struct {
	title string

	v []string
	m map[string]interface{}
}

func (m *mapInput) Names() []string { return m.v }

func (m *mapInput) Value(sel string) interface{} {
	p, ok := m.m[sel]
	if !ok {
		return nil
	}
	return pval(p)
}

func (m *mapInput) Set(sel string, port Port) error {
	p, ok := m.m[sel]
	if !ok {
		return fmt.Errorf("'%s' has no input named '%s'", m.title, sel)
	}
	if err := Connect(p, port); err != nil {
		return fmt.Errorf("cant connect '%s' port '%s': %s", m.title, sel, err.Error())
	}
	return nil
}

func (m *mapInput) Type(sel string) PortType {
	p, ok := m.m[sel]
	if !ok {
		return Invalid
	}
	return PtrTypeOf(p)
}

type mapOutput struct {
	title string
	v     []string
	m     map[string]interface{}
}

func (m *mapOutput) Names() []string { return m.v }

func (m *mapOutput) Value(sel string) interface{} {
	p, ok := m.m[sel]
	if !ok {
		return nil
	}
	return pval(p)
}

func (m *mapOutput) Get(sel string) (Port, error) {
	if p, ok := m.m[sel]; ok {
		if err := CheckOutput(p); err != nil {
			return nil, fmt.Errorf("'%s' port '%s' error: %s", m.title, sel, err.Error())
		}
		return Port(p), nil
	}
	return nil, fmt.Errorf("'%s' has no output named '%s'", m.title, sel)
}

type singleOutput struct {
	title string
	p     Port
}

func (o *singleOutput) Names() []string { return []string{""} }
func (o *singleOutput) Value(sel string) interface{} {
	if sel == "" {
		return pval(o.p)
	}
	return nil
}

func (o *singleOutput) Get(sel string) (Port, error) {
	if sel != "" {
		return nil, fmt.Errorf("'%s' has no named outputs; '%s' requested", o.title, sel)
	}
	if err := CheckOutput(o.p); err != nil {
		return nil, fmt.Errorf("'%s' unnamed port error: %s", o.title, err.Error())
	}
	return o.p, nil
}

type singleInput struct {
	title string
	ii    interface{}
}

func (i *singleInput) Names() []string { return []string{""} }
func (i *singleInput) Value(sel string) interface{} {
	if sel == "" {
		return pval(i.ii)
	}
	return nil
}

func (i *singleInput) Set(sel string, port Port) error {
	if sel != "" {
		return fmt.Errorf("'%s' has no named inputs; '%s' requested", i.title, sel)
	}
	if err := Connect(i.ii, port); err != nil {
		return fmt.Errorf("cant connect '%s' unnamed port: %s", i.title, err.Error())
	}
	return nil
}

func (i *singleInput) Type(sel string) PortType {
	if sel != "" {
		return Invalid
	}
	return PtrTypeOf(i.ii)
}

func pval(port interface{}) interface{} {
	switch p := port.(type) {
	case *bool:
		return *p
	case *float64:
		return *p
	case *int:
		return *p
	case **bool:
		return **p
	case **float64:
		return **p
	case **int:
		return **p
	}
	return nil
}
