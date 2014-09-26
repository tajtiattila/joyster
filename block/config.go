package block

type Config interface {
	Float64(string) float64
	OptFloat64(string, float64) float64
	Int(string) int
	OptInt(string, int) int
	Bool(string) bool
	OptBool(string, bool) bool

	TickFreq() float64 // ticks per seconds: 1e6/float64(c.UpdateMicros)
	TickTime() float64 // time in seconds elapsed duting one tick: float64(c.UpdateMicros) / 1e6
}
