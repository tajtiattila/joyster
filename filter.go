package main

import (
	"fmt"
)

// FilterSig represents a filter declaration
type FilterSig struct {
	Name string          // the name of the filter
	Argv []FilterArgSpec // fixed and optional arguments

	Help    string
	Factory interface{} // factory function specific to the call site
}

func (sig *FilterSig) ParseArgs(input interface{}) (args []interface{}, err error) {
	r := make([]interface{}, len(sig.Argv))

	type idxsig struct {
		idx int
		sig *FilterArgSpec
	}
	names := make(map[string]idxsig)
	poskeys := make(map[string]bool)
	for i, spec := range sig.Argv {
		names[spec.Name] = idxsig{i, &spec}
		if spec.Default != nil {
			r[i] = spec.Default
		} else {
			poskeys[spec.Name] = true
		}
	}

	switch input := input.(type) {
	case []interface{}:
		if len(input) < len(poskeys) {
			return nil, &ParseArgError{sig, input, nil, "too few arguments"}
		}
		if len(input) > len(sig.Argv) {
			return nil, &ParseArgError{sig, input, nil, "too many arguments"}
		}
		for i := range input {
			r[i], err = sig.Argv[i].Type.Parse(input[i])
			if err != nil {
				return nil, err
			}
		}
	case map[string]interface{}:
		for key, arg := range input {
			ispec, ok := names[key]
			if !ok {
				return nil, &ParseArgError{sig, input, key, "unknown argument index"}
			}
			r[ispec.idx], err = ispec.sig.Type.Parse(arg)
			if err != nil {
				return nil, err
			}
			delete(poskeys, ispec.sig.Name)
		}
	default:
		return nil, &ParseArgError{sig, input, nil, "unrecognised argument format"}
	}
	return r, nil
}

// FilterArgSpec represents an argument in a filter declaration
type FilterArgSpec struct {
	Name    string
	Type    FilterArgType
	Default interface{}
}

type FilterArgType interface {
	Help(string, interface{}) string
	Parse(interface{}) (interface{}, error)
}

type boolFilterArgType struct{}

func (*boolFilterArgType) Help(n string, def interface{}) string {
	m := fmt.Sprintf("%v«bool»", n)
	if def != nil {
		m += fmt.Sprintf("=%v", def)
	}
	return m
}

func (*boolFilterArgType) Parse(i interface{}) (interface{}, error) {
	if v, ok := i.(bool); ok {
		return v, nil
	}
	return nil, &ErrTypeMismatch{"boolean", i}
}

type numberFilterArgType struct{}

func (*numberFilterArgType) Help(n string, def interface{}) string {
	m := fmt.Sprintf("%v«float»", n)
	if def != nil {
		m += fmt.Sprintf("=%v", def)
	}
	return m
}

func (*numberFilterArgType) Parse(i interface{}) (interface{}, error) {
	if v, ok := i.(float64); ok {
		return v, nil
	}
	return nil, &ErrTypeMismatch{"boolean", i}
}

type subFilterArgType struct {
	m FilterMap
}

func (*subFilterArgType) Help(n string, def interface{}) string {
	return "[\"" + n + "\"]"
}

func (t *subFilterArgType) Parse(i interface{}) (interface{}, error) {
	println(2)
	return NewFilterCall(t.m, i)
}

type subvFilterArgType struct {
	m FilterMap
}

func (*subvFilterArgType) Help(n string, def interface{}) string {
	return "[[\"" + n + "\"], ...]"
}

func (t *subvFilterArgType) Parse(input interface{}) (oo interface{}, err error) {
	if input, ok := input.([]interface{}); ok {
		o := make([]*FilterCall, len(input))
		for i, ie := range input {
			if o[i], err = NewFilterCall(t.m, ie); err != nil {
				fmt.Println(err)
				return
			}
		}
		return o, nil
	}
	return nil, &ErrTypeMismatch{"filter array declaration", input}
}

type FilterMap map[string]*FilterSig

func FillFilterMap(m FilterMap, v ...*FilterSig) {
	for _, s := range v {
		if _, ok := m[s.Name]; ok {
			panic("duplicate filter " + s.Name)
		}
		m[s.Name] = s
	}
}

func (m FilterMap) Print() {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for _, k := range keys {
		sig := m[k]
		m := fmt.Sprintf("[\"%s\"", sig.Name)
		for _, a := range sig.Argv {
			m += ", " + a.Type.Help(a.Name, a.Default)
		}
		m += "]"
		if sig.Help != "" {
			m += "\n    - " + sig.Help
		}
		fmt.Println("  " + m)
	}
}

type FilterCall struct {
	Sig  *FilterSig
	Argv FilterArgs
}

func NewFilterCall(m FilterMap, input interface{}) (*FilterCall, error) {
	var ni, args interface{}
	switch input := input.(type) {
	case []interface{}:
		if len(input) < 1 {
			return nil, &FuncNameError{"Missing function name", input}
		}

		ni, args = input[0], input[1:]
	case map[string]interface{}:
		var ok bool
		if ni, ok = input["filter"]; !ok {
			return nil, &FuncNameError{"Missing function name", input}
		}
		args = input
	default:
		return nil, &ErrTypeMismatch{"filter declaration", input}
	}

	name, ok := ni.(string)
	if !ok {
		return nil, &FuncNameError{"function name must be a string", input}
	}

	sig, ok := m[name]
	if !ok {
		return nil, &FuncNameError{fmt.Sprintf("unknown function '%s'", name), input}
	}

	pa, err := sig.ParseArgs(args)
	if err != nil {
		return nil, err
	}
	return &FilterCall{sig, pa}, nil
}

type FilterArgs []interface{}

func (a FilterArgs) Args(out ...interface{}) {
	if len(a) != len(out) {
		panic(&ErrArgDecode{"type"})
	}
	for i, arg := range a {
		var ok bool
		switch o := out[i].(type) {
		case *bool:
			*o, ok = arg.(bool)
		case *float64:
			*o, ok = arg.(float64)
		case **FilterCall:
			*o, ok = arg.(*FilterCall)
		case *[]*FilterCall:
			*o, ok = arg.([]*FilterCall)
		}
		if !ok {
			panic(&ErrArgDecode{fmt.Sprint("target %T incompatible with argument %#v", arg, out[i])})
		}
	}
}

func (a FilterArgs) F() float64 {
	var f float64
	a.Args(&f)
	return f
}

// ErrArgDecode represents an internal error: the factory expected
// different arguments as specified by the signature
type ErrArgDecode struct {
	t string
}

func (e *ErrArgDecode) Error() string { return "Argument decode error: " + e.t }

// ErrTypeMismatch represents an incorrect value passed during a call
type ErrTypeMismatch struct {
	Expect string      // expected type name as string
	Value  interface{} // received type
}

func (e *ErrTypeMismatch) Error() string {
	return fmt.Sprintf("Unexpected %s value: %#v", e.Expect, e.Value)
}

type ParseArgError struct {
	Sig   *FilterSig
	Input interface{}
	Key   interface{}
	What  string
}

func (e *ParseArgError) Error() string {
	msg := "error while parsing for filter \"" + e.Sig.Name + "\""
	if e.Key != nil {
		msg += fmt.Sprint(" (key: %#v)", e.Key)
	}
	msg += ": " + e.What
	msg += fmt.Sprintf(" (got %#v)", e.Input)
	return msg
}

type FuncNameError struct {
	Msg   string
	Input interface{}
}

func (e *FuncNameError) Error() string {
	return fmt.Sprintf("%s (got %#v)", e.Msg, e.Input)
}
