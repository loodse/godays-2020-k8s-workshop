package smarthome

import (
	"testing"
	"time"
)

func TestShutter(t *testing.T) {
	tests := []struct {
		Name             string
		Start, SetClosed int
	}{
		{
			Name:  "from 0% to 43%",
			Start: 0, SetClosed: 43,
		},
		{
			Name:  "from 93% to 21%",
			Start: 93, SetClosed: 21,
		},
		{
			Name:  "from 21% to 0%",
			Start: 21, SetClosed: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			s := newShutter()
			s.closedPercentage = test.Start
			s.incrementWait = 0
			defer s.close()

			err := s.SetClosed(test.SetClosed)
			if err != nil {
				t.Error("unexpected error calling .SetClosed")
			}

			timer := time.NewTimer(500 * time.Millisecond)
			defer timer.Stop()

			var p int
			for {
				select {
				case <-timer.C:
					t.Errorf("timeout waiting for shutter to close to %d%%, is: %d", test.SetClosed, p)
					return
				default:
					p = s.State().Current
					if p == test.SetClosed {
						return
					}
				}
			}
		})
	}
}

func Test_capDiff(t *testing.T) {
	// if c := capDiff(-43, 10)
}
