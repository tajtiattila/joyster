package block

import (
	"fmt"
	"github.com/tajtiattila/joyster/block/parser"
)

func newParserTypeMap(tm TypeMap) parser.TypeMap {
	ptm := make(parserTypeMap)
	for _, t := range tm {
		pt := &parserType{name: t.Name(), typ: t}
		im := t.Input()
		if im != nil {
			pt.inames = im.Names()
			for _, n := range im.Names() {
				pt.im = append(pt.im, parser.Port{n, parser.PortType(im.Type(n))})
			}
		}
		ptm[t.Name()] = pt
	}
	return ptm
}

type parserTypeMap map[string]parser.Type

func (pm parserTypeMap) GetType(n string) (parser.Type, error) {
	if t, ok := pm[n]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("unknown type '%s'", n)
}

type parserType struct {
	name   string
	typ    Type
	inames []string
	im     parser.PortMap
	om     map[uint64]parser.PortMap
}

func (t *parserType) Input() parser.PortMap { return t.im }

func (t *parserType) Output(forinput parser.PortMap) (om parser.PortMap, err error) {
	im := make(PortTypeMap)
	for _, p := range forinput {
		im[p.Name] = PortType(p.Type)
	}
	bom, err := t.typ.Accept(im)
	if err != nil {
		return nil, err
	}
	for n, t := range bom {
		om = append(om, parser.Port{n, parser.PortType(t)})
	}
	return
}

func (t *parserType) Param(p parser.Param, globals parser.NamedParam) error {
	pp := &parseParam{parser.NewParamReader(p, globals)}
	err := t.typ.Verify(pp)
	if err != nil {
		return err
	}
	return pp.Err()
}

type parseParam struct {
	r parser.ParamReader
}

func (p *parseParam) Arg(n string) float64               { return p.r.Arg(n) }
func (p *parseParam) OptArg(n string, d float64) float64 { return p.r.OptArg(n, d) }
func (p *parseParam) TickFreq() float64                  { return p.r.OptArg(defaultTickFreqName, DefaultTickFreq) }
func (p *parseParam) TickTime() float64                  { return 1 / p.TickFreq() }
func (p *parseParam) Err() error                         { return p.r.Err() }
