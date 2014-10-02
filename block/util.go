package block

import (
	"fmt"
)

func MapInput(t string, m map[string]interface{}) InputMap {
	for _, p := range m {
		switch p.(type) {
		case **bool, **float64, **int:
		default:
			panic("invalid type in MapInput")
		}
	}
	return &mapInput{t, m}
}

func MapOutput(t string, m map[string]interface{}) OutputMap {
	for _, p := range m {
		switch p.(type) {
		case *bool, *float64, *int:
		default:
			panic("invalid type in MapOutput")
		}
	}
	return &mapOutput{t, m}
}

func SingleOutput(t string, p Port) OutputMap { return &singleOutput{t, p} }

func SingleInput(t string, i interface{}) InputMap { return &singleInput{t, i} }

type mapInput struct {
	title string
	m     map[string]interface{}
}

func (m *mapInput) Names() []string {
	v := make([]string, 0, len(m.m))
	for n := range m.m {
		v = append(v, n)
	}
	return v
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
	m     map[string]interface{}
}

func (m *mapOutput) Names() []string {
	v := make([]string, 0, len(m.m))
	for n := range m.m {
		v = append(v, n)
	}
	return v
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

func (i *singleInput) Set(sel string, port Port) error {
	if sel != "" {
		return fmt.Errorf("'%s' has no named outputs; '%s' requested", i.title, sel)
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
