package parser

type Blk struct {
	Name   string
	Type   Type
	Param  Param
	Inputs map[string]Port
}

type Port struct {
	Source *Blk
	Sel    string
}
