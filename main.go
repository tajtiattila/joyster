package main

import (
	"flag"
	"fmt"
	"github.com/tajtiattila/joyster/block"
	_ "github.com/tajtiattila/joyster/block/device/vjoy"
	_ "github.com/tajtiattila/joyster/block/device/xinput"
	_ "github.com/tajtiattila/joyster/block/logic"
	"github.com/tajtiattila/vjoy"
	"os"
	"strings"
	"time"
)

var (
	quiet   bool
	Version = "development"
)

func main() {
	var (
		prtver bool
		test   bool
		debugl string
		debug  []string
	)

	flag.BoolVar(&quiet, "quiet", false, "don't print info at startup")
	flag.BoolVar(&prtver, "version", false, "print version and exit")
	flag.BoolVar(&test, "test", false, "test config and exit")
	flag.StringVar(&debugl, "debug", "", "comma separated list of blocks to debug blocks")
	//flag.BoolVar(webgui, "web", false, "enable web gui")
	//flag.String(addr, "addr", ":7489", "web gui address")  // "JY"
	//flag.String(sharedir, "share", "share", "share directory") // "JY"
	flag.Parse()

	if debugl != "" {
		debug = strings.Split(debugl, ",")
	}

	if flag.NArg() > 1 {
		abort("exactly one config parameter required")
	}

	if !quiet {
		fmt.Println("joyster version:", Version)
		fmt.Println("vJoy version:", vjoy.Version())
		fmt.Println("  Product:       ", vjoy.ProductString())
		fmt.Println("  Manufacturer:  ", vjoy.ManufacturerString())
		fmt.Println("  Serial number: ", vjoy.SerialNumberString())
	}

	fn := "joyster.cfg"
	if flag.NArg() == 1 {
		fn = flag.Arg(0)
	}

	exit := test
	if prtver {
		fmt.Println(Version)
		exit = true
	}
	if exit {
		return
	}

	prof, err := block.Load(fn)
	if err != nil {
		abort(err)
	}
	defer func() {
		// prof might be changed by autoload
		prof.Close()
	}()

	var chdbg <-chan time.Time
	if len(debug) != 0 {
		chdbg = time.Tick(time.Second / 5)
	} else {
		chdbg = make(chan time.Time)
	}

	chcfg := autoloadconfig(fn)
	d := prof.D
	cht := time.Tick(d)
	for {
		select {
		case nprof := <-chcfg:
			if nprof.D != d {
				d = nprof.D
				cht = time.Tick(d)
			}
			prof.Close()
			prof = nprof
		case <-cht:
			prof.Tick()
		case <-chdbg:
			fmt.Println()
			block.DebugOutput(os.Stdout, prof, debug...)
		}
	}
}

func abort(a ...interface{}) {
	fmt.Println(a...)
	os.Exit(1)
}

func autoloadconfig(fn string) <-chan *block.Profile {
	ch := make(chan *block.Profile)
	if fi, err := os.Stat(fn); err == nil {
		t := fi.ModTime()
		go func() {
			for {
				if fi, err := os.Stat(fn); err == nil && fi.ModTime().After(t) {
					t = fi.ModTime()
					if prof, err := block.Load(fn); err == nil {
						ch <- prof
						if !quiet {
							fmt.Println("new config loaded")
						}
					} else {
						fmt.Println(err)
					}
				}
				time.Sleep(time.Second)
			}
		}()
	} else {
		panic("autoloadcfg " + fn + err.Error())
	}
	return ch
}
