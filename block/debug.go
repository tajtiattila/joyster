package block

import (
	"fmt"
	"io"
)

func DebugOutput(w io.Writer, p *Profile, names ...string) {
	for _, blk := range p.Blocks {
		if len(names) == 0 || has(names, p.Names[blk]) {
			fmt.Fprintf(w, "%s ", p.Names[blk])
			fmtio(w, blk.Input())
			fmt.Fprint(w, " → ")
			fmtio(w, blk.Output())
			fmt.Fprintln(w)
		}
	}
}

func has(v []string, s string) bool {
	for _, vv := range v {
		if vv == s || (len(s) > len(vv) && s[:len(vv)+1] == vv+"#") {
			return true
		}
	}
	return false
}

func fmtio(w io.Writer, bio IO) {
	if bio == nil {
		return
	}
	for _, n := range bio.Names() {
		v := bio.Value(n)
		var s string
		if v == nil {
			s = "nil"
		} else {
			switch i := v.(type) {
			case bool:
				if i {
					s = "●"
				} else {
					s = "○"
				}
			case int:
				switch {
				case (i & HatNorth) != 0:
					s = "↑"
				case (i & HatSouth) != 0:
					s = "↓"
				case (i & HatEast) != 0:
					s = "→"
				case (i & HatWest) != 0:
					s = "←"
				default:
					s = "·"
				}
			case float64:
				s = fmt.Sprintf("%+5.3f", i)
			default:
				s = "?"
			}
		}
		fmt.Fprint(w, s)
	}
}
