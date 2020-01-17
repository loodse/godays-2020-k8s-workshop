package smarthome

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"sort"
	"sync"
	"time"
)

type Shutter struct {
	Name            string
	Target, Current int
	Moving          bool
}

type ShutterClient struct {
	data    map[string]*shutter
	dataMux sync.Mutex
}

func newShutterClient() *ShutterClient {
	sc := &ShutterClient{
		data: map[string]*shutter{},
	}

	js, _ := ioutil.ReadFile("/tmp/godays2020/shutters.json")
	var shutters []Shutter
	_ = json.Unmarshal(js, &shutters)

	for _, shutter := range shutters {
		s := sc.getShutter(shutter.Name)
		s.closedPercentage = shutter.Current
	}
	return sc
}

func (sc *ShutterClient) close() {
	shutters, _ := sc.List(nil)
	js, _ := json.Marshal(shutters)
	_ = ioutil.WriteFile("/tmp/godays2020/shutters.json", js, 0700)

	for _, shutter := range sc.data {
		shutter.close()
	}
}

func (sc *ShutterClient) List(ctx context.Context) ([]Shutter, error) {
	sc.dataMux.Lock()
	defer sc.dataMux.Unlock()

	var shutters shuttersByName
	for _, shutter := range sc.data {
		shutters = append(shutters, shutter.Shutter())
	}

	sort.Sort(shutters)
	return shutters, nil
}

func (sc *ShutterClient) Get(ctx context.Context, name string) (Shutter, error) {
	sc.dataMux.Lock()
	defer sc.dataMux.Unlock()

	shutter := sc.getShutter(name).Shutter()
	return shutter, nil
}

func (sc *ShutterClient) Set(ctx context.Context, name string, percentageClosed int) error {
	sc.dataMux.Lock()
	defer sc.dataMux.Unlock()

	return sc.getShutter(name).Set(percentageClosed)
}

func (sc *ShutterClient) getShutter(name string) *shutter {
	if s, ok := sc.data[name]; ok {
		return s
	}
	sc.data[name] = newShutter(name)
	return sc.data[name]
}

// shuttersByName sorts Shutters by name
type shuttersByName []Shutter

func (items shuttersByName) Len() int {
	return len(items)
}

func (items shuttersByName) Swap(i, j int) {
	items[i], items[j] = items[j], items[i]
}

func (items shuttersByName) Less(i, j int) bool {
	return items[i].Name < items[j].Name
}

// shutter is the internal representation of a Shutter.
type shutter struct {
	name             string
	stateMux         sync.RWMutex
	closedPercentage int
	targetPercentage int
	moving           bool

	startOnce sync.Once
	requests  chan int

	incrementWait time.Duration
	maxIncrement  int
}

func newShutter(name string) *shutter {
	return &shutter{
		name:     name,
		requests: make(chan int, 10),

		// defaults
		incrementWait: 1 * time.Second,
		maxIncrement:  9,
	}
}

func (s *shutter) Set(closedPercentage int) error {
	if closedPercentage > 100 {
		return ValidationError("cannot close more than 100%%")
	}
	if closedPercentage < 0 {
		return ValidationError("cannot open more than 0%% closed")
	}

	s.startOnce.Do(func() {
		go s.worker()
	})
	s.requests <- closedPercentage
	return nil
}

func (s *shutter) Shutter() Shutter {
	s.stateMux.RLock()
	defer s.stateMux.RUnlock()
	return Shutter{
		Name:    s.name,
		Current: s.closedPercentage,
		Target:  s.targetPercentage,
		Moving:  s.moving,
	}
}

func (s *shutter) close() {
	close(s.requests)
}

func (s *shutter) worker() {
	for percentage := range s.requests {
		s.stateMux.Lock()
		s.moving = true
		s.targetPercentage = percentage
		s.stateMux.Unlock()

		for s.closedPercentage != s.targetPercentage {
			time.Sleep(s.incrementWait)

			// moving more than 10% per second would destroy the shutter ... and the window
			diff := capDiff(s.closedPercentage-percentage, s.maxIncrement)

			s.stateMux.Lock()
			s.closedPercentage -= diff
			s.stateMux.Unlock()
		}

		s.stateMux.Lock()
		s.moving = false
		s.targetPercentage = percentage
		s.stateMux.Unlock()
	}
}

func capDiff(diff, cap int) int {
	if diff > 0 {
		// limit positive move to cap
		if diff > cap {
			return cap
		}
		return diff
	}

	if diff < -cap {
		// limit negative move to cap
		return -cap
	}
	return diff
}
