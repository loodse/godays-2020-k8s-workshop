package smarthome

import (
	"sync"
	"time"
)

type ValidationError string

func (e ValidationError) Error() string {
	return string(e)
}

type Client struct {
	lights   map[string]*light
	shutters map[string]*shutter
}

func NewClient() *Client {
	return &Client{
		lights:   map[string]*light{},
		shutters: map[string]*shutter{},
	}
}

// registers a new Light with the given name.
func (c *Client) registerLight(name string) *light {
	c.lights[name] = &light{}
	return c.lights[name]
}

func (c *Client) LightState(name string) (isOn bool, err error) {
	light, ok := c.lights[name]
	if !ok {
		light = c.registerLight(name)
	}
	return light.IsOn(), nil
}

func (c *Client) LightOn(name string) (err error) {
	light, ok := c.lights[name]
	if !ok {
		light = c.registerLight(name)
	}
	light.On()
	return nil
}

func (c *Client) LightOff(name string) (err error) {
	light, ok := c.lights[name]
	if !ok {
		light = c.registerLight(name)
	}
	light.Off()
	return nil
}

// registers a new Shutter with the given name.
func (c *Client) registerShutter(name string) *shutter {
	c.shutters[name] = newShutter()
	return c.shutters[name]
}

type ShutterState struct {
	Target   int
	Current  int
	IsMoving bool
}

// ShutterState returns the current state of a shutter.
func (c *Client) ShutterState(name string) (state ShutterState, err error) {
	shutter, ok := c.shutters[name]
	if !ok {
		shutter = c.registerShutter(name)
	}
	return shutter.State(), nil
}

// SetShutter instructs a shutter to move to a certain percentage closed.
// Shutters cannot act instantly so they take some time to move to the requested position,
// before moving to a new position the shutter will finish the move to the laste position.
// use ShutterState to monitor the current position of a shutter.
func (c *Client) SetShutter(name string, closedPercentage int) error {
	shutter, ok := c.shutters[name]
	if !ok {
		shutter = c.registerShutter(name)
	}
	return shutter.SetClosed(closedPercentage)
}

type light struct {
	on bool
}

func (l *light) IsOn() bool {
	return l.on
}

func (l *light) On() {
	l.on = true
}

func (l *light) Off() {
	l.on = false
}

type shutter struct {
	stateMux         sync.RWMutex
	closedPercentage int
	targetPercentage int
	moving           bool

	startOnce sync.Once
	requests  chan int

	incrementWait time.Duration
	maxIncrement  int
}

func newShutter() *shutter {
	return &shutter{
		requests: make(chan int, 10),

		// defaults
		incrementWait: 1 * time.Second,
		maxIncrement:  9,
	}
}

func (s *shutter) close() {
	close(s.requests)
}

func (s *shutter) SetClosed(percentage int) error {
	if percentage > 100 {
		return ValidationError("cannot close more than 100%%")
	}
	if percentage < 0 {
		return ValidationError("cannot open more than 0%% closed")
	}

	s.startOnce.Do(func() {
		go s.worker()
	})
	s.requests <- percentage
	return nil
}

func (s *shutter) State() ShutterState {
	s.stateMux.RLock()
	defer s.stateMux.RUnlock()
	return ShutterState{
		Current:  s.closedPercentage,
		Target:   s.targetPercentage,
		IsMoving: s.moving,
	}
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
