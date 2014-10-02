package parser

func sort(ctx *context) error {
	// check parameters
	for _, blk := range ctx.vblk {
		if err := blk.Type.Validate(blk.Param, ctx.config); err != nil {
			lno := ctx.blklno[blk]
			return errf("block '%s' defined on line %d: %s", blk.Name, lno, err)
		}
	}

	// create set up links and create dependency map
	dm := make(map[*Blk]int)
	rm := make(map[*Blk]map[*Blk]bool)
	var err error
	for _, c := range ctx.vlink {
		if err = c.markdep(ctx, func(blk, dep *Blk) {
			if rm[dep] == nil {
				rm[dep] = make(map[*Blk]bool)
			}
			if !rm[dep][blk] {
				rm[dep][blk] = true
				dm[blk]++
			}
		}); err != nil {
			return err
		}
		if err = c.setup(ctx); err != nil {
			return err
		}
	}

	// sort blocks
	work := ctx.vblk
	next := make([]*Blk, 0, len(work))
	done := make([]*Blk, 0, len(work))

	for len(work) != 0 {
		progress := false
		for _, blk := range work {
			n := dm[blk]
			if n < 0 {
				panic("how?")
			}
			if n == 0 {
				done = append(done, blk)
				for rdep := range rm[blk] {
					dm[rdep]--
				}
				rm[blk] = nil
				progress = true
			} else {
				next = append(next, blk)
			}
		}

		if !progress {
			blk := work[0]
			return errf("circular dependency on block '%s' defined on line %d", blk.Name, ctx.blklno[blk])
		}
		work, next = next, work[:0]
	}

	ctx.vblk = done

	// validate block inputs
	for _, blk := range ctx.vblk {
		if blk.Type.MustHaveInput() && len(blk.Type.Input()) != len(blk.Inputs) {
			return errf("block '%s' defined on line %d has only %d of %d inputs set",
				blk.Name, ctx.blklno[blk], len(blk.Inputs), len(blk.Type.Input().Names()))
		}
		for _, p := range blk.Type.Input() {
			if input, ok := blk.Inputs[p.Name]; ok {
				pt := input.Type()
				if !Match(pt, p.Type) {
					return errf("block '%s' type mismatch for %s on line %d: want %s, have %s",
						blk.Name, nice(inport, p.Name), ctx.blklno[blk], PortStr(p.Type), PortStr(pt))
				}
			}
		}
	}

	return nil
}
