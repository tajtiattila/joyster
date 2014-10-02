package block

type Profile struct {
	Blocks  []Block
	Tickers []Ticker
}

func LoadProfile(fn string) *Profile {
	return nil
}

func (p *Profile) Tick() {
	for _, t := range p.Tickers {
		t.Tick()
	}
}

func (p *Profile) Close() error {
	var firsterr error
	for _, blk := range p.Blocks {
		if c, ok := blk.(Closer); ok {
			err := c.Close()
			if firsterr == nil {
				firsterr = err
			}
		}
	}
	return firsterr
}
