package main

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

type Input struct {
	LX, LY, RX, RY     float32
	LTrigger, RTrigger float32
	Buttons            uint64
}

type Output struct {
	X, Y, Z, RX, RY, RZ, U, V float32
	Buttons                   uint64
}

type Status struct {
	RollToYaw          bool
	TriggerYaw         bool
	HeadLook           bool
	I                  Input
	LXf, LYf, RXf, RYf float32
	O                  Output
}

////////////////////////////////////////////////////////////////////////////////

var ErrListenerClosed = errors.New("listener closed")

type StatusUpdate struct {
	r    *StatusRecycler
	nref uint32
	st   Status
}

func (u *StatusUpdate) addref() {
	atomic.AddUint32(&u.nref, 1)
}

func (u *StatusUpdate) decref() {
	if atomic.AddUint32(&u.nref, ^uint32(0)) == 0 && u.r != nil {
		u.r.give(u)
	}
}

type StatusRecycler struct {
	chget  <-chan *StatusUpdate
	chgive chan<- *StatusUpdate
}

func newStatusRecycler() *StatusRecycler {
	out, in := make(chan *StatusUpdate), make(chan *StatusUpdate)
	r := &StatusRecycler{out, in}
	go func() {
		pool := []*StatusUpdate{&StatusUpdate{r: r}}
		for {
			i := len(pool) - 1
			select {
			case out <- pool[i]:
				if i == 0 {
					pool = append(pool, &StatusUpdate{r: r})
				}
			case s := <-in:
				pool = append(pool, s)
			}
		}
	}()
	return r
}

func (r *StatusRecycler) get() *StatusUpdate {
	u := <-r.chget
	u.nref = 1 // no race possible
	return u
}

func (r *StatusRecycler) give(u *StatusUpdate) {
	r.chgive <- u
}

type StatusListener struct {
	rd       *StatusUpdate
	chi, cho chan *StatusUpdate
	closed   bool
}

func newStatusListener() *StatusListener {
	l := &StatusListener{
		rd:  new(StatusUpdate),
		chi: make(chan *StatusUpdate),
		cho: make(chan *StatusUpdate),
	}
	go func() {
		for {
			s, ok := <-l.chi
			if !ok {
				close(l.cho)
				return
			}
		Loop:
			for {
				select {
				case ss, ok := <-l.chi:
					if !ok {
						close(l.cho)
						return
					}
					s.decref()
					s = ss
				case l.cho <- s:
					break Loop
				}
			}
		}
	}()
	return l
}

func (l *StatusListener) send(u *StatusUpdate) error {
	if l.closed {
		return ErrListenerClosed
	}
	u.addref()
	l.chi <- u
	return nil
}

func (l *StatusListener) Read() (*Status, error) {
	if d, ok := <-l.cho; ok {
		l.rd.decref()
		l.rd = d
		return &d.st, nil
	}
	return nil, io.EOF
}

func (l *StatusListener) Close() error {
	if !l.closed {
		close(l.chi)
		l.closed = true
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

var DefaultStatusDispatcher StatusDispatcher

type StatusDispatcher struct {
	m     map[*StatusListener]bool
	rcyc  *StatusRecycler
	mtx   sync.RWMutex
	last  *StatusUpdate
	delch chan *StatusListener
}

func (d *StatusDispatcher) NewListener() *StatusListener {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	if d.m == nil {
		d.m = make(map[*StatusListener]bool)
		d.rcyc = newStatusRecycler()
		d.last = d.rcyc.get()

		d.delch = make(chan *StatusListener)
		go func() {
			for {
				if l, ok := <-d.delch; ok {
					d.mtx.Lock()
					delete(d.m, l)
					d.mtx.Unlock()
				} else {
					return
				}
			}
		}()
	}
	l := newStatusListener()
	d.m[l] = true
	return l
}

func (d *StatusDispatcher) Update(s *Status) {
	if d.last == nil || d.last.st == *s {
		return
	}
	d.last.decref()
	d.last = d.rcyc.get()
	d.last.st = *s

	d.mtx.RLock()
	defer d.mtx.RUnlock()

	for l := range d.m {
		if err := l.send(d.last); err != nil {
			d.delch <- l
		}
	}
}
