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
	*context
	r *sourcereader
}

func newparser(p []byte) *parser {
	return &parser{
		context: &context{
			config: make(map[string]float64),
			blocks: map[string]blockspec{
				"true":      constbool(true),
				"on":        constbool(true),
				"false":     constbool(false),
				"off":       constbool(false),
				"hat_off":   constint(-1),
				"hat_north": constint(0),
				"hat_east":  constint(1),
				"hat_south": constint(2),
				"hat_west":  constint(3),
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
		tblks, ok := p.blocks[c.name]
		if !ok {
			return fmt.Errorf("line %d: conn target %s missing", p.r.lineno(), c.name)
		}
		tblk, err := tblks.Prepare(p.context)
		if err != nil {
			return fmt.Errorf("line %d: %s", p.r.lineno(), err.Error())
		}
		is, ok := tblk.(block.InputSetter)
		fmt.Printf("%d %#v %#v\n", c.lineno, tblk, is)
		if !ok {
			return fmt.Errorf("line %d: can't set input on '%s'", p.r.lineno(), c.name)
		}
		sblk, err := c.blockspec.Prepare(p.context)
		if err != nil {
			return fmt.Errorf("line %d: %s", p.r.lineno(), err.Error())
		}
		port, err := sblk.Output(c.sel)
		if err != nil {
			return fmt.Errorf("line %d: %s", p.r.lineno(), err.Error())
		}
		is.SetInput(c.sel, port)
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
			v := p.parseparam()
			if len(v.P) != 0 {
				panic("positional arguments in set")
			}
			for n, v := range v.N {
				p.config[n] = v
			}
			p.r.endstatement()
		case p.r.eat("block"):
			name := p.r.name()
			if _, ok := p.blocks[name]; ok {
				panic("duplicate name")
			}
			p.blocks[name] = p.parseblockspec(true)
			p.r.endstatement()
		case p.r.eat("conn"):
			name, spec := p.r.spec()
			lineno := p.r.lineno()
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
		grp := &groupblkspec{lineno: p.r.lineno(), sels: names}
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
		return grp
	case isnumstart(p.r.ch()):
		n := p.r.number()
		return &valueblkspec{constblkspec{&n}, p.r.lineno()}
	default:
		n, s := p.r.spec()
		return &namedblkspec{p.r.lineno(), n, s}
	}
}

func (p *parser) parsetypedblockspec(inputs bool) *factoryblkspec {
	dollar := p.r.eatch('$')
	if dollar {
		p.r.skiplinespace()
	}
	if !p.r.eatch('[') {
		panic("invalid typed block spec")
	}
	blk := &factoryblkspec{lineno: p.r.lineno(), dollar: dollar, typ: p.r.name()}
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
		}
		if !inputs {
			panic("input declaration not allowed")
		}
		blk.inputs = append(blk.inputs, p.parseblockspec(inputs))
	}
	return blk
}

func (p *parser) parseparam() *block.Param {
	v := new(block.Param)
	p.r.skiplinespace()
	if isdigit(p.r.ch()) || p.r.ch() == '-' {
		for {
			v.P = append(v.P, p.r.number())
			if !isdigit(p.r.ch()) {
				break
			}
		}
	} else {
		v.N = make(map[string]float64)
		for {
			name := p.r.name()
			if !p.r.eatch('=') {
				panic("'=' expected after name")
			}
			v.N[name] = p.r.number()
			p.r.skiplinespace()
			if !isnamestart(p.r.ch()) {
				break
			}
		}
	}
	return v
}

func constint(v int) *constblkspec   { return &constblkspec{&v} }
func constbool(b bool) *constblkspec { return &constblkspec{&b} }

func isblkdefstart(r rune) bool {
	return r == '[' || r == '$'
}
