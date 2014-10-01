package parser

func newcontext(t TypeMap) *context {
	return &context{
		TypeMap:     t,
		config:      make(map[string]float64),
		sinkNames:   make(portMap),
		sourceNames: make(portMap),
		portNames: map[string]specSource{
			"true":       constbool(true),
			"on":         constbool(true),
			"false":      constbool(false),
			"off":        constbool(false),
			"hat_off":    constint(hatC),
			"hat_centre": constint(hatC),
			"hat_center": constint(hatC),
			"hat_north":  constint(hatN),
			"hat_east":   constint(hatE),
			"hat_south":  constint(hatS),
			"hat_west":   constint(hatW),
			"centre":     constint(hatC),
			"center":     constint(hatC),
			"north":      constint(hatN),
			"east":       constint(hatE),
			"south":      constint(hatS),
			"west":       constint(hatW),
		},
	}
}

type context struct {
	TypeMap
	config      map[string]float64
	portNames   map[string]specSource
	sinkNames   portMap
	sourceNames portMap
	blklno      map[*Blk]int
	vblk        []*Blk
	vlink       []Link
}

type portMap map[string]portMapper
type portMapper interface {
	port(isel string) (blk *Blk, osel string, err error)
}

type Link struct {
	sink   specSink
	source specSource
}

func (l Link) markdep(c *context, f func(*Blk, *Blk)) error {
	consumer, err := l.sink.Blk(c)
	if err != nil {
		return err
	}
	producer, err := l.source.Blk(c)
	if err != nil {
		return err
	}
	if consumer != nil && producer != nil {
		f(consumer, producer)
	}
	return nil
}

func (l Link) setup(c *context) error {
	src, err := l.source.Source(c)
	if err != nil {
		return err
	}
	return l.sink.SetTo(c, src)
}

type specSink interface {
	Blk(c *context) (*Blk, error)
	SetTo(c *context, src Source) error
}

type specSource interface {
	Blk(c *context) (*Blk, error)
	Source(c *context) (Source, error)
}
