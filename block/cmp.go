package block

func RegisterCmpFunc(name string, fn func(a, b float64) bool) {
	Register(name, func() Block {
		return &cmpopblk{typ: name, tick: fn}
	})
}

type cmpopblk struct {
	typ    string
	o      bool
	i1, i2 *float64
	tick   func(a, b float64) bool
}

func (b *cmpopblk) Tick() {
	b.o = b.tick(*b.i1, *b.i2)
}

func (b *cmpopblk) Input() InputMap {
	return MapInput(b.typ, map[string]interface{}{"1": &b.i1, "2": &b.i2})
}
func (b *cmpopblk) Output() OutputMap { return SingleOutput(b.typ, &b.o) }
func (b *cmpopblk) Validate() error   { return CheckInputs(b.typ, &b.i1, &b.i2) }

func init() {
	RegisterCmpFunc("eq", func(a, b float64) bool { return eq(a, b) })
	RegisterCmpFunc("ne", func(a, b float64) bool { return !eq(a, b) })
	RegisterCmpFunc("lt", func(a, b float64) bool { return a < b })
	RegisterCmpFunc("gt", func(a, b float64) bool { return a > b })
	RegisterCmpFunc("le", func(a, b float64) bool { return a <= b })
	RegisterCmpFunc("ge", func(a, b float64) bool { return a >= b })
}

func eq(a, b float64) bool {
	d := a - b
	const rounderr = 1e-4
	return -rounderr < d && d < rounderr
}
