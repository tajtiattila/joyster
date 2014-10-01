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
	lineno int
	err    error
}

func (e *sourceerror) Error() string {
	return fmt.Sprintf("line %d: ", e.lineno) + e.err.Error()
}

func srcerr(s srcliner, i interface{}) error {
	n := s.SrcLine()
	switch x := i.(type) {
	case *sourceerror:
		return x
	case error:
		return &sourceerror{n, x}
	}
	return &sourceerror{n, errors.New(fmt.Sprint(i))}
}

func srcerrf(s srcliner, f string, args ...interface{}) error {
	return srcerr(s, fmt.Sprintf(f, args...))
}
