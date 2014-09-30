package block

const DefaultTickTime = 1e-3 // 1 millisecond

type Param interface {
	Arg(string) float64
	OptArg(string, float64) float64

	TickFreq() float64 // ticks per seconds: 1e6/float64(c.UpdateMicros)
	TickTime() float64 // time in seconds elapsed duting one tick: float64(c.UpdateMicros) / 1e6
}
