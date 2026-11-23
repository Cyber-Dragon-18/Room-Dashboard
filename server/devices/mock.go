package devices

import (
	"fmt"
	"log"
)

// MockLight is a fake light for development: it keeps its state in memory
// and logs every call, so you can build and test without the real lamp.
type MockLight struct {
	state LightState
}

func NewMockLight() *MockLight {
	return &MockLight{state: LightState{Mode: "white", Brightness: 100, Temp: 50, R: 255, G: 255, B: 255}}
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
	log.Printf("[mock] SetColor(%d, %d, %d)", r, g, b)
	m.state.Mode = "colour"
	m.state.R, m.state.G, m.state.B = r, g, b
	return nil
}

func (m *MockLight) SetWhite(tempPct int) error {
	if tempPct < 0 || tempPct > 100 {
		return fmt.Errorf("temperature must be 0-100, got %d", tempPct)
	}
	log.Printf("[mock] SetWhite(%d)", tempPct)
	m.state.Mode = "white"
	m.state.Temp = tempPct
	return nil
}

func (m *MockLight) State() (LightState, error) {
	return m.state, nil
}

var _ Light = (*MockLight)(nil)
