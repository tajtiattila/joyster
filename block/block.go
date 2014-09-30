package block

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

type BlockFactory func(Param) (Block, error)

type TypeMap map[string]BlockFactory

var DefaultTypeMap = make(TypeMap)

func Register(name string, fn func() Block) {
	if _, ok := DefaultTypeMap[name]; ok {
		panic("duplicate name: " + name)
	}
	DefaultTypeMap[name] = func(Param) (Block, error) {
		return fn(), nil
	}
}

func RegisterParam(name string, fn BlockFactory) {
	if _, ok := DefaultTypeMap[name]; ok {
		panic("duplicate name: " + name)
	}
	DefaultTypeMap[name] = fn
}
