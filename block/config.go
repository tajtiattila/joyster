package block

const DefaultTickTime = 1e-3 // 1 millisecond

type Param interface {
	Arg(string) float64
	OptArg(string, float64) float64

	TickFreq() float64 // ticks per seconds: 1e6/float64(c.UpdateMicros)
	TickTime() float64 // time in seconds elapsed duting one tick: float64(c.UpdateMicros) / 1e6
}

// ProtoParam represents an empty parameter map used
// during prototype creation in type checks. Arg() values will are eauql to 0.5,
// and OptArg() always returns the default. An update frequency of 1e3 is assumed.
// Block types registered using Register() or RegisterParam() should not
// return an error when this value is provided.
var ProtoParam Param = new(protoparam)

type protoparam struct{}

func (*protoparam) Arg(n string) float64               { return 0.5 }
func (*protoparam) OptArg(n string, d float64) float64 { return d }
func (*protoparam) TickFreq() float64                  { return 1e3 }
func (*protoparam) TickTime() float64                  { return 1e-3 }
