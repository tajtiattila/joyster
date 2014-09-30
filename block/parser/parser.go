package parser

import (
	"fmt"
	"github.com/tajtiattila/joyster/block"
)

/*
stmt :=
	'block' name blockspec
	'conn' portspec portspec
	'set' namedarglist
arglist := posarglist | namedarglist
posarglist :=
	value [' ' value]*
value :=
	digit* [
namedarglist :=
	name '=' value [' ' value]*
portspecs := portspec [' ' portspec]*
portspec :=
	name
	name.selector
blockspec :=
	plugvalue
	portspec
	newblockspec
	{ newblockspec [newblockspec]* }
newblockspec :=
	'[' blocktype [portspecs] [':' arglist] ']'
	'$' '[' blocktype [':' arglist] ']'
	'{' inputnames [blockspec]* '}'
plugvalue :=
	0
	1
	hatoff
	floatnumber
*/

type parser struct {
	*Context
	r     *sourcereader
	specs []blockspec
}

func newparser(p []byte) *parser {
	return &parser{
		Context: &Context{
			Profile:   new(block.Profile),
			config:    make(map[string]float64),
			blockline: make(map[block.Block]int),
			names: map[string]blockspec{
				"true":       constbool(true),
				"on":         constbool(true),
				"false":      constbool(false),
				"off":        constbool(false),
				"hat_off":    constint(block.HatCentre),
				"hat_centre": constint(block.HatCentre),
				"hat_center": constint(block.HatCentre),
				"hat_north":  constint(block.HatNorth),
				"hat_east":   constint(block.HatEast),
				"hat_south":  constint(block.HatSouth),
				"hat_west":   constint(block.HatWest),
				"centre":     constint(block.HatCentre),
				"center":     constint(block.HatCentre),
				"north":      constint(block.HatNorth),
				"east":       constint(block.HatEast),
				"south":      constint(block.HatSouth),
				"west":       constint(block.HatWest),
			},
		},
		r: &sourcereader{src: p},
	}
}

func (p *parser) parse() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = p.r.formaterror(r.(string))
		}
	}()
	p.parseimpl()
	for _, c := range p.conns {
		tblks, ok := p.names[c.name]
		if !ok {
			return srcerrf(p.r, "conn target %s missing", c.name)
		}
		p.dependency(tblks, c.blockspec)
	}
	for _, s := range p.specs {
		s.Deps(p.Context)
	}
	for {
		done, progress := true, false
		for k := range p.deps {
			var w []blockspec
			for _, dep := range p.deps[k] {
				done = false
				if wd, ok := p.deps[dep]; !ok || len(wd) == 0 {
					progress = true
					blk, err := dep.Prepare(p.Context)
					if err != nil {
						return err
					}
					fmt.Printf("depfix %s -> %p\n", dep.String(), blk)
					if err = blk.Validate(); err != nil {
						return errf("dependency failure on %s: %v", dep.String(), err)
					}
				} else {
					w = append(w, dep)
				}
			}
			p.deps[k] = w
		}
		if done {
			break
		}
		if !progress {
			return errf("circular dependency")
		}
	}
	for _, c := range p.conns {
		tblks, ok := p.names[c.name]
		if !ok {
			return srcerrf(p.r, "conn target %s missing", c.name)
		}
		tblk, err := tblks.Prepare(p.Context)
		if err != nil {
			return srcerr(p.r, err)
		}
		is := tblk.Input()
		if is == nil {
			return srcerrf(p.r, "conn target %s has no inputs", c.name)
		}
		sblk, err := c.blockspec.Prepare(p.Context)
		if err != nil {
			return srcerr(p.r, err)
		}
		os := sblk.Output()
		if os == nil {
			return srcerrf(p.r, "source of conn %s has no outputs", c.name)
		}
		port, err := os.Get("")
		if err != nil {
			return srcerr(p.r, err)
		}
		is.Set(c.sel, port)
	}
	for blk, lineno := range p.blockline {
		if err := blk.Validate(); err != nil {
			return errf("line %d: validation failure: %v", lineno, err)
		}
	}
	return
}

func (p *parser) parseimpl() {
	for !p.r.eof() {
		p.r.skipallspace()
		switch {
		case p.r.eatch('#'):
			p.r.skipline()
		case p.r.eat("set"):
			p.r.skiplinespace()
			m, ok := p.parseparam().(namedparamspec)
			if !ok {
				panic("positional arguments in set")
			}
			for n, v := range m {
				p.config[n] = v
			}
			p.r.endstatement()
		case p.r.eat("block"):
			name := p.r.name()
			if _, ok := p.names[name]; ok {
				panic("duplicate name")
			}
			p.names[name] = p.parseblockspec(true)
			p.r.endstatement()
		case p.r.eat("conn"):
			name, spec := p.r.spec()
			lineno := p.r.sourceline()
			p.conns = append(p.conns, connspec{name, spec, p.parseblockspec(true), lineno})
			p.r.endstatement()
		default:
			panic("unexpected")
		}
	}
}

func (p *parser) parseblockspec(inputs bool) blockspec {
	p.r.skiplinespace()
	switch {
	case isblkdefstart(p.r.ch()):
		return p.parsetypedblockspec(inputs)
	case p.r.eatch('{'):
		var names []string
		for {
			p.r.skipallspace()
			if isblkdefstart(p.r.ch()) {
				break
			}
			names = append(names, p.r.name())
		}
		grp := &groupblkspec{lineno: p.r.sourceline(), sels: names}
		for {
			p.r.skipallspace()
			if !isblkdefstart(p.r.ch()) {
				if p.r.eatch('}') {
					break
				}
				panic("unclosed group")
			}
			grp.v = append(grp.v, p.parsetypedblockspec(false))
		}
		return p.addspec(grp)
	case isnumstart(p.r.ch()):
		n := p.r.number()
		return p.addspec(&valueblkspec{constblkspec{&n}, p.r.sourceline()})
	default:
		n, s := p.r.spec()
		return p.addspec(&namedblkspec{p.r.sourceline(), n, s})
	}
}

func (p *parser) parsetypedblockspec(allowinput bool) *factoryblkspec {
	dollar := p.r.eatch('$')
	if dollar {
		p.r.skiplinespace()
	}
	if !p.r.eatch('[') {
		panic("invalid typed block spec")
	}
	blk := &factoryblkspec{lineno: p.r.sourceline(), dollar: dollar, typ: p.r.name()}
	for {
		p.r.skiplinespace()
		if p.r.eatch(']') {
			break
		}
		if p.r.eatch(':') {
			blk.param = p.parseparam()
			if !p.r.eatch(']') {
				panic("unclosed argument block")
			}
			break
		} else {
			var emptyparam posparamspec
			blk.param = emptyparam
		}
		if !allowinput {
			panic("input declaration not allowed")
		}
		input := p.parseblockspec(allowinput)
		blk.inputs = append(blk.inputs, input)
	}
	p.addspec(blk)
	return blk
}

func (p *parser) parseparam() paramspec {
	p.r.skiplinespace()
	if isdigit(p.r.ch()) || p.r.ch() == '-' {
		var param posparamspec
		for {
			param = append(param, p.r.number())
			if !isdigit(p.r.ch()) {
				break
			}
		}
		return param
	} else {
		param := make(namedparamspec)
		for {
			name := p.r.name()
			if !p.r.eatch('=') {
				panic("'=' expected after name")
			}
			param[name] = p.r.number()
			p.r.skiplinespace()
			if !isnamestart(p.r.ch()) {
				break
			}
		}
		return param
	}
	return nil
}

func (p *parser) addspec(s blockspec) blockspec {
	p.specs = append(p.specs, s)
	return s
}

func constint(v int) *constblkspec   { return &constblkspec{&v} }
func constbool(b bool) *constblkspec { return &constblkspec{&b} }

func isblkdefstart(r rune) bool {
	return r == '[' || r == '$'
}
