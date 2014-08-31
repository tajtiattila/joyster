package main

import (
	"io"
)

// CleanJSON cleans json comments in place of v, and returns the new length
func CleanJSON(v []byte) int {
	j := &minifier{state: startState}
	return j.Bytes(v)
}

type JSONMinifiedReader struct {
	r io.Reader
	c *minifier
}

func NewJSONMinifiedReader(r io.Reader) *JSONMinifiedReader {
	return &JSONMinifiedReader{r, &minifier{state: startState}}
}

func (jr *JSONMinifiedReader) Read(v []byte) (n int, err error) {
	nread := 0
	for n < len(v) && err == nil {
		nread, err = jr.r.Read(v[n:])
		nread = jr.c.Bytes(v[n : n+nread])
		n += nread
	}
	return
}

type minifier struct {
	state      func(*minifier, byte)
	buf        []byte
	rpos, wpos int
}

func (m *minifier) getch() byte {
	p := m.rpos
	m.rpos++
	return m.buf[p]
}

func (m *minifier) putch(ch byte) {
	p := m.wpos
	m.wpos++
	m.buf[p] = ch
}

// bytes cleans json in place of v, and returns the new length
func (m *minifier) Bytes(v []byte) int {
	m.buf, m.rpos, m.wpos = v, 0, 0
	for m.rpos < len(m.buf) {
		m.state(m, m.getch())
	}
	return m.wpos
}

func startState(m *minifier, ch byte) {
	switch ch {
	case '/':
		m.state = slashState
		return
	case '"':
		m.state = quoteState
	}
	m.putch(ch)
	return
}

func quoteState(m *minifier, ch byte) {
	m.putch(ch)
	switch ch {
	case '"':
		m.state = startState
	case '\\':
		m.state = quoteBackslashState
	}
}

func quoteBackslashState(m *minifier, ch byte) {
	m.putch(ch)
	if ch == 'u' {
		m.state = quoteUnicode0State
	} else {
		m.state = quoteState
	}
}

func quoteUnicode0State(m *minifier, ch byte) { m.putch(ch); m.state = quoteUnicode1State }
func quoteUnicode1State(m *minifier, ch byte) { m.putch(ch); m.state = quoteUnicode2State }
func quoteUnicode2State(m *minifier, ch byte) { m.putch(ch); m.state = quoteUnicode3State }
func quoteUnicode3State(m *minifier, ch byte) { m.putch(ch); m.state = quoteState }

func slashState(m *minifier, ch byte) {
	switch ch {
	case '/':
		m.state = eolCommentState
	case '*':
		m.state = cstyleCommentState
	default:
		m.putch('/')
		if isSpace(ch) {
			m.putch(ch)
		}
		m.state = startState
	}
}

func eolCommentState(m *minifier, ch byte) {
	if ch == '\n' {
		m.state = startState
	}
}

func cstyleCommentState(m *minifier, ch byte) {
	if ch == '*' {
		m.state = cstyleCommentEndState
	}
}

func cstyleCommentEndState(m *minifier, ch byte) {
	switch ch {
	case '/':
		m.state = startState
	case '*':
		// no op
	default:
		m.state = cstyleCommentState
	}
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}
