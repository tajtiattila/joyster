package block

type Profile struct {
	Blocks  []Block
	Tickers []Ticker
}

func (p *Profile) Tick() {
	for _, t := range p.Tickers {
		t.Tick()
	}
}
