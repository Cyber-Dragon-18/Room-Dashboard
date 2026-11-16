package devices

import (
	"fmt"
	"log"
)

type MockLight struct {
	state LightState
}

func (m *MockLight) SetPower(on bool) error {
	log.Printf("[mock] SetPower(%v)", on)
	m.state.On = on
	return nil
}

func (m *MockLight) SetBrightness(pct int) error {

	if pct < 0 || pct > 100 {
		return fmt.Errorf("brightness must be 0-100, got %d", pct)
	}

	log.Printf("[mock] SetBrightness(%d)", pct)
	m.state.Brightness = pct
	return nil
}

func (m *MockLight) SetColor(r, g, b uint8) error {
	log.Printf("[mock] SetColor(%d,%d,%d)", r, g, b)
	m.state.R = r
	m.state.B = b
	m.state.G = g
	return nil
}

func (m *MockLight) State() (LightState, error) {
	return m.state, nil
}

var _ Light = (*MockLight)(nil)
