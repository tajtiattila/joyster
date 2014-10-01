package parser

type Profile struct {
	Blocks *Blk
}

type Blk struct {
	Name   string
	Type   Type
	Param  Param
	Inputs map[string]Source
}

func (b *Blk) SetInput(name string, src Source) error {
	if b.Inputs == nil {
		b.Inputs = make(map[string]Source)
	}
	if old, ok := b.Inputs[name]; ok && old != src {
		return errf("block '%s' has %s already set", b.Name, nice(inport, name))
	}
	b.Inputs[name] = src
	return nil
}

type Source interface{}

type BlkPortSource struct {
	Blk *Blk
	Sel string
}

type ValueSource struct {
	Value interface{}
}
