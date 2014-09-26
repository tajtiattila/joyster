package main

import (
	"math"
)

type Node interface {
	Outputs() []Output
}

type Type int

const {
	Invalid Type = iota
	Bool
	Scalar
	Vector
}

// BlockClass represents block types the user can use
type BlockType struct {
	Type    string
	Inputs  map[string]Type
	Create  func(i []Input) Block
}

// specification for a user defined block
type BlockConfig struct {
	Name   string
	Type  string
	Inputs []string
}

type Profile struct {
	blocks map[string]BlockConfig
}

func (p *Profile) Create(c *data) (what, error) {
	ctx := &createContext{data:c, seen:make(map[string]Block)}
	for _, bc := range p.blocks {
		bt := DefaultBlockTypeMap.Lookup(bc.Type)
		if bt == nil {
			return nil, fmt.Errorf("unknown block type %s", bc.Type)
		}
		if bt.Output {
			p.createBlock(ctx, nil, bc)
		}
	}
}

func (p *Profile) createBlock(ctx *createContext, stack []string, bc *BlockConfig) Block {
	var err error
	for _, n := range stack {
		if n == bc.Name {
			panic(fmt.Errorf("recursion: %s", bc.Name))
		}
	}
	stack = append(stack, bc.Name)
	bt := DefaultBlockTypeMap.Lookup(bc.Type)
	iv := make([]Input, len(bc.Inputs))
	for i, src := range bc.Inputs {
		blkn, in := splitBlockInput(src)
		cblk, ok := ctx.seen[blkn]
		if !ok {
			bcc, ok := p.blocks[blkn]
			if !ok {
				panic(fmt.Errorf("Source %s unknown %s", blkn, bc.Name))
			}
			cblk = createBlock(ctx, stack, bcc)
		}
		iv[i] = cblk.Input(in)
		if iv[i] == nil {
			if len(in) > 2 {
				inv, ext := in[:len(in)-2], in[len(in)-2:]
				switch ext {
				case ".x", ".y":
					if ix = cblk.Input(inv); ix != 0 {
						if six, ok := ix.(VectorInput); ok {
							switch ext {
							case ".x":
								iv[i] = xscalar(six)
							case ".y":
								iv[i] = yscalar(six)
							}
						}
					}
				}
			}
		}
		if iv[i] == nil {
			panic(fmt.Errorf("Input %s unknown in %s (%s)", in, bc.Name, bt.Type))
		}
	}
	blk := bt.Create(iv)
	ctx.seen[bc.Name] = blk
	return blk
}

func splitBlockInput(blkin string) (blkn, in string) {
	for i, r := range blkin {
		if r == "." {
			return blkin[:i], blkin[i:]
		}
	}
	return blkin, ""
}

type createContext struct {
	data  *createData
	seen  map[string]Block
}

type Scalar float64

type Vector struct {
	X, Y Scalar
}

func (p *Profile) Connect(i, o string) error {
	return nil
}

type circular_deadzone struct {
	value Scalar
}

func (c *circular_deadzone) CreateNode() Node {
}

func (c *circular_deadzone) StartNode() []interface{} {
	value := c.value
	value2 := value * value
	o, i := make(chan Vector), make(chan Vector)
	go func() {
		for p := range i {
			mag2 := p.X*p.X + p.Y*p.Y
			if mag2 <= value2 {
				p.X, p.Y = 0, 0
			} else {
				mag := Scalar(math.Sqrt(float64(mag2)))
				m := 1 - value/mag
				p.X *= m
				p.Y *= m
			}
			o <- p
		}
	}()
	return cxn(o, i)
}

type enableAxis struct {}
func (*enableAxis) StartNode() []interface{} {
	o, i1, i2 := make(chan Scalar), make(chan Scalar), make(chan bool)
	go func() {
		var (
			vi, vo      Scalar
			enabled, ok bool
		)
		for {
			select {
			case vi, ok = <-i1:
				if !ok {
					return
				}
				if enabled {
					vo = vi
				}
			case enabled, ok = <-i2:
				if !ok {
					return
				}
				if !enabled {
					vo = 0
				}
			case o <- vo:
			}
		}
	}()
	return cxn(o, i1, i2)
}

func cxn(cv ...interface{}) []interface{} {
	r := make([]interface{}, len(cv))
	for i, c := range cv {
		if i == 0 {
			switch c := c.(type) {
			case chan bool:
				r[i] = <-chan bool(c)
			case <-chan bool:
				r[i] = c
			case chan Scalar:
				r[i] = <-chan Scalar(c)
			case <-chan Scalar:
				r[i] = c
			case chan Vector:
				r[i] = <-chan Vector(c)
			case <-chan Vector:
				r[i] = c
			}
		} else {
			switch c := c.(type) {
			case chan bool:
				r[i] = chan<- bool(c)
			case chan<- bool:
				r[i] = c
			case chan Scalar:
				r[i] = chan<- Scalar(c)
			case chan<- Scalar:
				r[i] = c
			case chan Vector:
				r[i] = chan<- Vector(c)
			case chan<- Vector:
				r[i] = c
			}
		}
	}
	return r
}

type VJoyNode struct {
}

type BoolInput interface {
	Bool() bool
}

type ScalarInput interface {
	Scalar() Scalar
}

func ScalarInputFunc func() Scalar

func (f ScalarInputFunc) Scalar() Scalar {
	return f()
}

func xscalar(i VectorInput) ScalarInput {
	return func() Scalar {
		x, _ := i.Vector()
		return x
	}
}

func yscalar(i VectorInput) ScalarInput {
	return func() Scalar {
		_, y := i.Vector()
		return y
	}
}

func axisFuncInput(i []Input) {

}

type VectorInput interface {
	Vector() (Scalar, Scalar)
}

func VectorInputFunc func() (Scalar, Scalar)

func (f VectorInputFunc) Vector() (Scalar, Scalar) {
	return f()
}

type BoolOutput interface {
	Bool(bool)
}

type ScalarOutput interface {
	Scalar(Scalar)
}

type VectorOutput interface {
	Vector(Scalar, Scalar)
}

type VectorLink struct {
	in, out chan Vector
}

func NewVectorLink() *VectorLink {
	l := &VectorLink{in: make(chan Vector), out: make(chan Vector)}
	go func() {
		var v Vector
		select {
		case v = <-l.in:
		case l.out <- v:
		}
	}()
	return l
}



