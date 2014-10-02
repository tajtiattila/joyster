package block

import (
	"fmt"
)

// a block is a set of inputs and outputs
// producer blocks may have nil inputs, and sink blocks may have nil outputs
type Block interface {
	Input() InputMap
	Output() OutputMap
	Validate() error
}

// ticker is an interface for blocks that needs update once upon each update tick.
type Ticker interface {
	Tick()
}

type Closer interface {
	Close() error
}

type Type interface {
	Name() string
	New(Param) (Block, error)

	Verify(Param) error // verify parameters
	Input() TypeInputMap
	Accept(in PortTypeMap) (PortTypeMap, error) // tests wether type accepts in, and what it would return in this case

	MustHaveInput() bool
}

type TypeMap map[string]Type

var DefaultTypeMap = make(TypeMap)

func Register(name string, fn func() Block) {
	RegisterType(&simpleType{name, func(Param) (Block, error) {
		return fn(), nil
	}})
}

func RegisterParam(name string, fn func(Param) (Block, error)) {
	RegisterType(&simpleType{name, fn})
}

func RegisterType(t Type) {
	if _, ok := DefaultTypeMap[t.Name()]; ok {
		panic("duplicate name: " + t.Name())
	}
	DefaultTypeMap[t.Name()] = t
}

type simpleType struct {
	name string
	f    func(Param) (Block, error)
}

func (t *simpleType) Name() string               { return t.name }
func (t *simpleType) New(p Param) (Block, error) { return t.f(p) }
func (t *simpleType) MustHaveInput() bool        { return false }

func (t *simpleType) Verify(p Param) error {
	_, err := t.f(p)
	return err
}

func (t *simpleType) Input() TypeInputMap {
	blk, err := t.f(ProtoParam)
	if err != nil {
		panic(fmt.Sprintf("simpleType '%s' does not accept ProtoParam", t.name))
	}
	return blk.Input()
}

func (t *simpleType) Accept(i PortTypeMap) (PortTypeMap, error) {
	blk, err := t.f(ProtoParam)
	if err != nil {
		panic(fmt.Sprintf("simpleType '%s' does not accept ProtoParam", t.name))
	}
	blki := blk.Input()
	for name, typ := range i {
		has := false
		for _, n := range blki.Names() {
			if n == name {
				has = true
				break
			}
		}
		if !has {
			return nil, fmt.Errorf("simpleType '%s' has no input '%s'", t.name, name)
		}
		if err = blki.Set(name, ZeroValue(typ)); err != nil {
			return nil, fmt.Errorf("simpleType '%s' does not accept zero input for '%s': %v", t.name, name, err)
		}
	}
	blko := blk.Output()
	om := make(PortTypeMap)
	for _, name := range blko.Names() {
		o, err := blko.Get(name)
		if err != nil {
			panic(fmt.Sprintf("simpleType '%s' does not have named param '%s': %v", t.name, name, err))
		}
		if o != Invalid {
			om[name] = TypeOf(o)
		}
	}
	return om, nil
}
