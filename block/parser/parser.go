package parser

import (
	"fmt"
	//"fmt"
)

/*
stmt :=
	'block' name blockspec
	'port' name block ['.' spec]
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

// copy of values in block
const (
	hatC = 0
	hatN = 1
	hatE = 2
	hatS = 4
	hatW = 8
)

type portdir int

const (
	port portdir = iota
	outport
	inport
)

var singlePort = []string{""}

func newparser(t TypeMap) *parser {
	return &parser{context: newcontext(t)}
}

func (p *parser) parse(src []byte) (err error) {
	p.r = &sourcereader{src: src}
	defer func() {
		if r := recover(); r != nil {
			err = p.r.formaterror(r)
		}
	}()
	p.parseimpl()

	/*
		for _, c := range p.Conns {
			_, ok := p.Names[c.name]
			if !ok {
				return srcerrf(c, "conn target %s missing", c.name)
			}
		}
	*/

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
			m, ok := p.parseparam().(NamedParam)
			if !ok {
				panic("'set' needs named parameters")
			}
			for n, v := range m {
				p.config[n] = v
			}
			p.r.endstatement()
		case p.r.eat("port"):
			name := p.r.name()
			p.r.skiplinespace()
			blk, spec := p.r.spec()
			p.portNames[name] = nsource(p.r.sourceline(), blk, spec)
		case p.r.eat("block"):
			name := p.r.name()
			p.topblockspec(name)
			p.r.endstatement()
		case p.r.eat("conn"):
			name, spec := p.r.spec()
			lineno := p.r.sourceline()
			p.vlink = append(p.vlink, Link{nsink(lineno, name, spec), p.parsesource()})
			p.r.endstatement()
		default:
			panic("unexpected")
		}
	}
}

func (p *parser) topblockspec(name string) {
	if _, ok := p.portNames[name]; ok {
		panic("duplicate name")
	}
	if _, ok := p.sinkNames[name]; ok {
		panic("duplicate name")
	}
	if _, ok := p.sourceNames[name]; ok {
		panic("duplicate name")
	}
	p.r.skiplinespace()
	switch p.r.ch() {
	case '{':
		p.parsegroup(name)
	case '[':
		p.parseblock(name)
	default:
		panic("invalid block specification")
	}
	return
}

func (p *parser) parsesource() (input specSource) {
	p.r.skiplinespace()
	switch {
	case p.r.ch() == '[':
		lno := p.r.sourceline()
		f, inputs := p.parsefactory(&blkconstraint{
			inpdisp:    inpdef_required,
			mustoutput: []string{""},
		})
		blk := p.newblk(lno, fmt.Sprintf("«%s:%d»", f.tname, lno), f.typ, f.param)
		for i, n := range f.typ.InputNames() {
			p.vlink = append(p.vlink, Link{&concreteblksink{lno, blk, n}, inputs[i]})
		}
		input = &concreteblksource{lno, blk, ""}
	case isnumstart(p.r.ch()):
		n := p.r.number()
		input = &valueport{p.r.sourceline(), constport{&n}}
	default:
		n, s := p.r.spec()
		input = nsource(p.r.sourceline(), n, s)
	}
	return
}

func (p *parser) parseblock(name string) {
	lno := p.r.sourceline()
	f, inputs := p.parsefactory(&blkconstraint{})
	blk := p.newblk(lno, name, f.typ, f.param)
	if len(inputs) == 0 {
		p.sinkNames[name] = blk
	} else {
		for i, n := range f.typ.InputNames() {
			p.vlink = append(p.vlink, Link{&concreteblksink{lno, blk, n}, inputs[i]})
		}
	}
	p.sourceNames[name] = blk
}

func (p *parser) parsegroup(name string) {
	if !p.r.eatch('{') {
		panic("invalid group block spec")
	}
	var names []string
	for {
		p.r.skipallspace()
		if isblkdefstart(p.r.ch()) {
			break
		}
		names = append(names, p.r.name())
	}
	// TODO
	idx := 0
	var last, first portMapper
	for {
		p.r.skipallspace()
		if !isblkdefstart(p.r.ch()) {
			if p.r.eatch('}') {
				break
			}
			panic("unclosed group")
		}
		dollar := p.r.eatch('$')
		lno := p.r.sourceline()
		var cur portMapper
		if dollar {
			f, _ := p.parsefactory(haveinout(""))
			m := make(map[string]*Blk)
			for _, sel := range names {
				blk := p.newblk(lno, fmt.Sprintf("%s#%d.%s", name, idx, sel), f.typ, f.param)
				m[sel] = blk
			}
			cur = &dollarPortMapper{p.r.sourceline(), m}
		} else {
			f, _ := p.parsefactory(haveinout(names...))
			blk := p.newblk(lno, fmt.Sprintf("%s#%d", name, idx), f.typ, f.param)
			cur = blk
		}
		if idx == 0 {
			first = cur
		} else {
			for _, sel := range names {
				eblk, esel, err := last.port(sel)
				if err != nil {
					panic(err) // impossible
				}
				kblk, ksel, err := cur.port(sel)
				if err != nil {
					panic(err) // impossible
				}
				p.vlink = append(p.vlink, Link{&concreteblksink{lno, kblk, ksel}, &concreteblksource{lno, eblk, esel}})
			}
		}
		last = cur
		idx++
	}
	if idx == 0 {
		panic("group has no elements")
	}
	p.sinkNames[name] = first
	p.sourceNames[name] = last
}

func (p *parser) parsefactory(constr *blkconstraint) (f *factory, inputs []specSource) {
	p.r.skiplinespace()
	if !p.r.eatch('[') {
		panic("invalid factory block spec")
	}
	f = new(factory)
	f.tname = p.r.name()
	var err error
	f.typ, err = p.GetType(f.tname)
	if err != nil {
		panic(err)
	}
	if len(constr.mustoutput) != 0 {
		for _, n := range constr.mustoutput {
			if !has(f.typ.OutputNames(), n) {
				panic(errf("type '%s' does not have required %s",
					f.tname, nice(outport, n)))
			}
		}
	}
	if len(constr.mustonlyinput) != 0 {
		if len(constr.mustonlyinput) != len(f.typ.InputNames()) {
			panic(errf("type '%s' must have %d inputs, not %d",
				f.tname, len(constr.mustonlyinput), len(f.typ.InputNames())))
		}
		for _, n := range constr.mustonlyinput {
			if !has(f.typ.InputNames(), n) {
				panic(errf("type '%s' does not have required %s", f.tname, nice(inport, n)))
			}
		}
	}

	for {
		p.r.skiplinespace()
		if p.r.eatch(']') {
			break
		}
		if p.r.eatch(':') {
			f.param = p.parseparam()
			if !p.r.eatch(']') {
				panic("unclosed argument block")
			}
			break
		} else {
			var emptyparam PosParam
			f.param = emptyparam
		}
		if constr.inpdisp == inpdef_prohibited {
			panic(errf("input declaration for '%s' not allowed", f.tname))
		}
		inputs = append(inputs, p.parsesource())
	}

	switch constr.inpdisp {
	case inpdef_required:
		if len(inputs) != len(f.typ.InputNames()) {
			panic(errf("input count mismatch for '%s': needed %d, have %d",
				f.tname, len(f.typ.InputNames()), len(inputs)))
		}
	case inpdef_allowed:
		if len(inputs) != 0 && len(inputs) != len(f.typ.InputNames()) {
			panic(errf("input count mismatch for '%s': needs either zero or %d, have %d",
				f.tname, len(f.typ.InputNames()), len(inputs)))
		}
	}

	return
}

func (p *parser) parseparam() Param {
	p.r.skiplinespace()
	if isdigit(p.r.ch()) || p.r.ch() == '-' {
		var param PosParam
		for {
			param = append(param, p.r.number())
			if !isdigit(p.r.ch()) {
				break
			}
		}
		return param
	} else {
		param := make(NamedParam)
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
	return nil // not reached
}

func (p *parser) newblk(lno int, name string, typ Type, param Param) *Blk {
	blk := &Blk{Name: name, Type: typ, Param: param}
	if p.blklno == nil {
		p.blklno = make(map[*Blk]int)
	}
	p.blklno[blk] = lno
	p.vblk = append(p.vblk, blk)
	return blk
}

func constint(v int) *constport   { return &constport{v} }
func constbool(b bool) *constport { return &constport{b} }

func nice(d portdir, n string) string {
	var dir string
	switch d {
	case inport:
		dir = "input"
	case outport:
		dir = "output"
	default:
		dir = "port"
	}
	if n == "" {
		return "unnamed " + dir
	}
	return dir + " '" + n + "'"
}

func isblkdefstart(r rune) bool {
	return r == '[' || r == '$'
}

const (
	inpdef_allowed = iota
	inpdef_required
	inpdef_prohibited
)

type blkconstraint struct {
	inpdisp       int
	mustonlyinput []string // blk must have exactly these inputs
	mustoutput    []string // blk must have this output fields
}

var anyblock = &blkconstraint{}

func needoutput(names ...string) *blkconstraint { return &blkconstraint{mustoutput: names} }

func haveinout(names ...string) *blkconstraint {
	return &blkconstraint{
		inpdisp:       inpdef_prohibited,
		mustonlyinput: names,
		mustoutput:    names,
	}
}

func (c *blkconstraint) valueallowed() bool {
	if c.inpdisp == inpdef_required {
		return false
	}
	if len(c.mustoutput) > 1 || len(c.mustoutput) == 1 && c.mustoutput[0] != "" {
		return false
	}
	return true
}
