package block

import (
	"fmt"
	"runtime"
	"time"

	"github.com/tajtiattila/joyster/block/parser"
)

type Profile struct {
	D       time.Duration
	Blocks  []Block
	Tickers []Ticker
	Names   map[Block]string
}

func Parse(src string) (*Profile, error) {
	return ParseProfile(src, DefaultTypeMap)
}

func ParseProfile(src string, tm TypeMap) (*Profile, error) {
	p, err := parser.Parse(src, newParserTypeMap(tm))
	if err != nil {
		return nil, err
	}
	return instantiate(p, tm)
}

func Load(fn string) (*Profile, error) {
	return LoadProfile(fn, DefaultTypeMap)
}

func LoadProfile(fn string, tm TypeMap) (*Profile, error) {
	p, err := parser.LoadProfile(fn, newParserTypeMap(tm))
	if err != nil {
		return nil, err
	}
	return instantiate(p, tm)
}

func (p *Profile) Tick() {
	for _, t := range p.Tickers {
		t.Tick()
	}
}

func (p *Profile) Close() error {
	var firsterr error
	for _, blk := range p.Blocks {
		if c, ok := blk.(Closer); ok {
			err := c.Close()
			if firsterr == nil {
				firsterr = err
			}
		}
	}
	p.Blocks = nil
	p.Tickers = nil
	return firsterr
}

func instantiate(pprof *parser.Profile, tm TypeMap) (p *Profile, err error) {
	p = new(Profile)
	psave := p
	v, ok := pprof.Config[defaultTickFreqName]
	if !ok {
		v = DefaultTickFreq
	}
	p.D = time.Duration(float64(time.Second) / v)
	defer func() {
		if err != nil {
			psave.Close()
		}
	}()
	p.Names = make(map[Block]string)
	mblk := make(map[*parser.Blk]Block)
	for _, pb := range pprof.Blocks {
		ptyp, ok := pb.Type.(*parserType)
		if !ok {
			return nil, fmt.Errorf("unexpected type for block '%s'", pb.Name)
		}
		param := &parseParam{parser.NewParamReader(pb.Param, pprof.Config)}
		blk, err := ptyp.typ.New(param)
		if err != nil {
			return nil, err
		}
		if param.Err() != nil {
			return nil, fmt.Errorf("block '%s' setup error: %v", param.Err())
		}
		mblk[pb] = blk
		for name, port := range pb.Inputs {
			var p Port
			switch i := port.(type) {
			case *parser.BlkPortSource:
				iblk, ok := mblk[i.Blk]
				if !ok {
					return nil, fmt.Errorf("input block '%s' missing", i.Blk.Name)
				}
				if p, err = iblk.Output().Get(i.Sel); err != nil {
					return nil, fmt.Errorf("input port '%s' of block '%s' missing", i.Sel, i.Blk.Name)
				}
			case *parser.ValueSource:
				p = valuePort(i)
			default:
				return nil, fmt.Errorf("unexpected input '%s' for block '%s'", name, pb.Name)
			}
			if err = blk.Input().Set(name, p); err != nil {
				return nil, fmt.Errorf("can't set input '%s' on block '%s': %v", name, pb.Name, err)
			}
		}
		if err = blk.Validate(); err != nil {
			return nil, fmt.Errorf("loaded block '%s' invalid: %v", pb.Name, err)
		}
		p.Blocks = append(p.Blocks, blk)
		p.Names[blk] = pb.Name
		if t, ok := blk.(Ticker); ok {
			p.Tickers = append(p.Tickers, t)
		}
	}
	runtime.GC()
	// do a test tick to see if everything is in order
	if err := testtick(p); err != nil {
		return nil, err
	}
	return p, nil
}

func testtick(p *Profile) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred in first tick: %v", r)
		}
	}()
	p.Tick()
	return
}

func valuePort(vs *parser.ValueSource) Port {
	switch v := vs.Value.(type) {
	case bool:
		return &v
	case float64:
		return &v
	case int:
		return &v
	}
	return nil
}
