package parser

import (
	"bytes"
	"fmt"
	"unicode/utf8"
)

type sourcereader struct {
	src   []byte
	nline int
	pline int
	pos   int
}

func newsourcereader(p []byte) *sourcereader {
	return &sourcereader{src: p}
}

func (r *sourcereader) sourceline() int {
	return r.nline + 1
}

func (r *sourcereader) eof() bool {
	return r.pos == len(r.src)
}

func (r *sourcereader) ch() rune {
	run, _ := utf8.DecodeRune(r.src[r.pos:])
	return run
}

func (r *sourcereader) eatch(ch rune) bool {
	run, siz := utf8.DecodeRune(r.src[r.pos:])
	if run == ch {
		r.pos += siz
		return true
	}
	return false
}

func (r *sourcereader) haslen(siz int) bool {
	return r.pos+siz <= len(r.src)
}

func (r *sourcereader) endstatement() {
	ok := false
	for !r.eof() {
		run, siz := utf8.DecodeRune(r.src[r.pos:])
		switch run {
		case '#':
			r.skipline()
			ok = true
		case '\n':
			r.nline++
			r.pline = r.pos + 1
			fallthrough
		case ';':
			ok = true
			fallthrough
		case ' ', '\t':
			r.pos += siz
		default:
			if !ok {
				panic("unterminated statement")
			}
			return
		}
	}
}

func (r *sourcereader) skiplinespace() { r.skip(islinespace) }
func (r *sourcereader) skipallspace()  { r.skip(isspace) }

func (r *sourcereader) skip(f func(rune) bool) {
	for {
		run, siz := utf8.DecodeRune(r.src[r.pos:])
		if !f(run) {
			return
		}
		r.pos += siz
		if run == '\n' {
			r.nline++
			r.pline = r.pos
		}
	}
}

func (r *sourcereader) skipline() {
	i := bytes.IndexByte(r.src[r.pos:], '\n')
	if i != -1 {
		r.pos += i + 1
		r.nline++
		r.pline = r.pos
	} else {
		r.pos = len(r.src)
	}
}

func (r *sourcereader) eat(n string) bool {
	part := r.src[r.pos:]
	if len(part) >= len(n) && string(part[:len(n)]) == n {
		run, siz := utf8.DecodeRune(part[len(n):])
		if siz == 0 || isspace(run) {
			r.pos += len(n)
			if run != '\n' {
				r.pos += siz
			}
			return true
		}
	}
	return false
}

// parse number, eat space after value
func (r *sourcereader) number() float64 {
	sign := float64(1)
	if r.eatch('-') {
		sign = -1
	} else {
		r.eatch('+')
	}
	n, _ := r.digits()
	fracn, fracd := int64(0), int64(1)
	if r.eatch('.') {
		fracn, fracd = r.digits()
	}
	run, siz := utf8.DecodeRune(r.src[r.pos:])
	unit, skip := float64(1), true
	switch run {
	case 'k':
		unit = 1e3
	case 'm':
		unit = 1e-3
	case 'u', 'μ':
		unit = 1e-6
	case 'n':
		unit = 1e-9
	default:
		skip = false
	}
	if skip {
		r.pos += siz
	}
	r.skiplinespace()
	return sign * (float64(n) + float64(fracn)/float64(fracd)) * unit
}

func (r *sourcereader) digits() (n, d int64) {
	d = 1
	for {
		run, siz := utf8.DecodeRune(r.src[r.pos:])
		if !isdigit(run) {
			break
		}
		n = n*10 + int64(run-'0')
		r.pos += siz
	}
	return
}

func (r *sourcereader) name() string {
	part := r.src[r.pos:]
	run, siz := utf8.DecodeRune(part)
	if !isnamestart(run) {
		panic("not a name")
	}
	i := siz
	for {
		run, siz := utf8.DecodeRune(part[i:])
		if !isnamepart(run) {
			break
		}
		i += siz
	}
	r.pos += i
	return string(part[:i])
}

func (r *sourcereader) spec() (name, sel string) {
	name = r.name()
	if len(r.src) == 0 {
		return
	}
	if r.eatch('.') {
		part := r.src[r.pos:]
		run, siz := utf8.DecodeRune(part)
		if !isnamestart(run) && !isdigit(run) {
			panic("invalid selector")
		}
		var f func(rune) bool
		if isdigit(run) {
			f = isdigit
		} else {
			f = isnamepart
		}
		i := siz
		for {
			run, siz := utf8.DecodeRune(part[i:])
			if !f(run) {
				break
			}
			i += siz
		}
		r.pos += i
		sel = string(part[:i])
	}
	return
}

func (r *sourcereader) formaterror(msg interface{}) error {
	after := string(r.src[r.pline:r.pos])
	return fmt.Errorf("line %d: %s... %v", r.nline, after, msg)
}

type parseerr struct {
}

func isspace(ch rune) bool {
	return islinespace(ch) || ch == '\n'
}

func islinespace(ch rune) bool {
	return ch == ' ' || ch == '\r' || ch == '\t' || ch == '\xA0'
}

func isnumstart(ch rune) bool {
	return isdigit(ch) || ch == '+' || ch == '-' || ch == '.'
}

func isdigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isnamepart(ch rune) bool {
	return isnamestart(ch) || isdigit(ch)
}

func isnamestart(ch rune) bool {
	return ch == '_' || ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}
