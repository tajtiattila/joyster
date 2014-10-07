package block

import (
	"fmt"
)

// a block is a set of inputs and outputs
// producer blocks may have nil inputs, and sink blocks may have nil outputs
type Block interface {
	Input() InputMap
	Output() OutputMap

	// Validate reports whether the Block is valid after setup.
	// It must check if all its non-optional inputs were set correctly.
	// Optional inputs should be set to a usable value typically
	// during Block creation.
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
}

type TypeMap map[string]Type

var DefaultTypeMap = make(TypeMap)

func Register(name string, fn func() Block) {
	RegisterType(&Proto{name, true, func(Param) (Block, error) {
		return fn(), nil
	}})
}

func RegisterParam(name string, fn func(Param) (Block, error)) {
	RegisterType(&Proto{name, true, fn})
}

func RegisterType(t Type) {
	if _, ok := DefaultTypeMap[t.Name()]; ok {
		panic("duplicate name: " + t.Name())
	}
	DefaultTypeMap[t.Name()] = t
}

// Proto is a simple block type that implements input and output port
// reporting using a prototype block.
type Proto struct {
	TypeName  string
	NeedInput bool
	Create    func(Param) (Block, error)
}

func (t *Proto) Name() string               { return t.TypeName }
func (t *Proto) New(p Param) (Block, error) { return t.Create(p) }

func (t *Proto) Verify(p Param) error {
	blk, err := t.Create(p)
	if c, ok := blk.(Closer); ok {
		c.Close()
	}
	return err
}

func (t *Proto) Input() TypeInputMap {
	blk, err := t.Create(ProtoParam)
	if err != nil {
		panic(fmt.Sprintf("Proto '%s' does not accept ProtoParam", t.TypeName))
	}
	if c, ok := blk.(Closer); ok {
		defer c.Close()
	}
	return blk.Input()
}

func (t *Proto) Accept(i PortTypeMap) (PortTypeMap, error) {
	blk, err := t.Create(ProtoParam)
	if err != nil {
		panic(fmt.Sprintf("Proto '%s' does not accept ProtoParam", t.TypeName))
	}
	if c, ok := blk.(Closer); ok {
		defer c.Close()
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
			return nil, fmt.Errorf("Proto '%s' has no input '%s'", t.TypeName, name)
		}
		if typ == Invalid || typ == Any {
			return nil, fmt.Errorf("Proto '%s' can't test invalid/untyped input '%s'", t.TypeName, name)
		}
		if err = blki.Set(name, ZeroValue(typ)); err != nil {
			return nil, fmt.Errorf("Proto '%s' does not accept zero input for '%s': %v", t.TypeName, name, err)
		}
	}
	blko := blk.Output()
	if blko != nil {
		om := make(PortTypeMap)
		for _, name := range blko.Names() {
			o, err := blko.Get(name)
			if err != nil {
				panic(fmt.Sprintf("Proto '%s' does not have named param '%s': %v", t.TypeName, name, err))
			}
			if o != Invalid {
				om[name] = TypeOf(o)
			}
		}
		return om, nil
	}
	return nil, nil
}
