package parser

import (
	"errors"
	"fmt"
)

type srcliner interface {
	SrcLine() int
}

func errf(f string, v ...interface{}) error {
	return fmt.Errorf(f, v...)
}

type sourceerror struct {
	name   string
	lineno int
	err    error
}

func (e *sourceerror) Error() string {
	name := e.name
	if name != "" {
		name = "{input}"
	}
	return fmt.Sprintf("%s:%d: ", name, e.lineno) + e.err.Error()
}

func setsrc(obj interface{}, name string) error {
	if e, ok := obj.(*sourceerror); ok {
		return &sourceerror{name, e.lineno, e.err}
	}
	return &sourceerror{name, -1, errors.New(fmt.Sprint(obj))}
}

func srcerr(obj interface{}, i interface{}) error {
	switch s := obj.(type) {
	case srcliner:
		n := s.SrcLine()
		switch x := i.(type) {
		case *sourceerror:
			return x
		case error:
			return &sourceerror{"", n, x}
		}
		return &sourceerror{"", n, errors.New(fmt.Sprint(i))}
	case error:
		return s
	}
	return errors.New(fmt.Sprint(i))
}

func srcerrf(s interface{}, f string, args ...interface{}) error {
	return srcerr(s, fmt.Sprintf(f, args...))
}
