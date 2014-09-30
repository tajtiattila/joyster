package block

func RegisterScalarFunc(name string, fn func(Param) (func(float64) float64, error)) {
	RegisterParam(name, func(p Param) (Block, error) {
		f, err := fn(p)
		if err != nil {
			return nil, err
		}
		return &scalarfnblk{typ: name, f: f}, nil
	})
}

type scalarfnblk struct {
	typ string
	i   *float64
	o   float64
	f   func(float64) float64
}

func (b *scalarfnblk) Tick()             { b.o = b.f(*b.i) }
func (b *scalarfnblk) Input() InputMap   { return SingleInput(b.typ, &b.i) }
func (b *scalarfnblk) Output() OutputMap { return SingleOutput(b.typ, &b.o) }
func (b *scalarfnblk) Validate() error   { return CheckInputs(b.typ, &b.i) }

func init() {
	RegisterBoolFunc("toggle", func(Param) (func(bool) bool, error) {
		var val, last bool
		return func(v bool) bool {
			if v != last {
				last = v
				if v {
					val = !val
				}
			}
			return val
		}, nil
	})
}

func RegisterBoolFunc(name string, fn func(Param) (func(bool) bool, error)) {
	RegisterParam(name, func(p Param) (Block, error) {
		f, err := fn(p)
		if err != nil {
			return nil, err
		}
		return &boolfnblk{typ: name, f: f}, nil
	})
}

type boolfnblk struct {
	typ string
	i   *bool
	o   bool
	f   func(bool) bool
}

func (b *boolfnblk) Tick()             { b.o = b.f(*b.i) }
func (b *boolfnblk) Input() InputMap   { return SingleInput(b.typ, &b.i) }
func (b *boolfnblk) Output() OutputMap { return SingleOutput(b.typ, &b.o) }
func (b *boolfnblk) Validate() error   { return CheckInputs(b.typ, &b.i) }

type StickFunc func(xi, yi float64) (xo, yo float64)

func RegisterStickFunc(name string, ff func(p Param) (StickFunc, error)) {
	RegisterParam(name, func(p Param) (Block, error) {
		f, err := ff(p)
		if err != nil {
			return nil, err
		}
		return &stickfuncblk{typ: name, f: f}, nil
	})
}

type stickfuncblk struct {
	typ    string
	xi, yi *float64
	xo, yo float64
	f      func(xi, yi float64) (xo, yo float64)
}

func (b *stickfuncblk) Input() InputMap {
	return MapInput(b.typ, map[string]interface{}{"x": &b.xi, "y": &b.yi})
}

func (b *stickfuncblk) Output() OutputMap {
	return MapOutput(b.typ, map[string]interface{}{"x": &b.xo, "y": &b.yo})
}

func (b *stickfuncblk) Validate() error { return CheckInputs(b.typ, &b.xi, &b.yi) }
func (b *stickfuncblk) Tick()           { b.xo, b.yo = b.f(*b.xi, *b.yi) }

func init() {
	Register("stick", func() Block { return new(stickblk) })
}

type stickblk struct {
	x, y *float64
}

func (b *stickblk) Input() InputMap {
	return MapInput("stick", map[string]interface{}{"x": &b.x, "y": &b.y})
}

func (b *stickblk) Output() OutputMap {
	return MapOutput("stick", map[string]interface{}{"x": b.x, "y": b.y})
}

func (b *stickblk) Validate() error { return CheckInputs("stick", &b.x, &b.y) }
