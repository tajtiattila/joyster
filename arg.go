package main

func F(n string) FilterArgSpec {
	return FilterArgSpec{n, &numberFilterArgType{}, nil}
}

func B(n string) FilterArgSpec {
	return FilterArgSpec{n, &boolFilterArgType{}, nil}
}

func OptF(n string, v float64) FilterArgSpec {
	return FilterArgSpec{n, &numberFilterArgType{}, v}
}

func OptB(n string, v bool) FilterArgSpec {
	return FilterArgSpec{n, &boolFilterArgType{}, v}
}

func Sub(n string, m FilterMap) FilterArgSpec {
	return FilterArgSpec{n, &subFilterArgType{m}, nil}
}

func Subv(n string, m FilterMap) FilterArgSpec {
	return FilterArgSpec{n, &subvFilterArgType{m}, nil}
}

func Sig(name string, args ...FilterArgSpec) *FilterSig {
	sig := &FilterSig{Name: name, Argv: args}
	opt := false
	for _, a := range sig.Argv {
		if a.Default != nil {
			opt = true
		} else {
			if opt {
				panic("normal argument after optional")
			}
		}
	}
	return sig
}

/*func Factory(s *FilterSig, f interface{}) *FilterSig {
	s.Factory = f
	return s
}*/
