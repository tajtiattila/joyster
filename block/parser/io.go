package parser

import (
	"github.com/tajtiattila/joyster/block"
	"io"
	"io/ioutil"
)

func Read(r io.Reader) (*block.Profile, error) {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return Parse(src)
}

func Load(fn string) (*block.Profile, error) {
	src, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	return Parse(src)
}

func Parse(src []byte) (*block.Profile, error) {
	p := newparser(src)
	if err := p.parse(); err != nil {
		return nil, err
	}
	return p.Context.Profile, nil
}
