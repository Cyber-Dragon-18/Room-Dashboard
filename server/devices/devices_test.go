package devices

import "testing"

func TestMockBrightnessValidation(t *testing.T) {
	m := &MockLight{}
	if err := m.SetBrightness(150); err == nil {
		t.Error("expected error for 150%, got nil")
	}
	if err := m.SetBrightness(50); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	st, _ := m.State()
	if st.Brightness != 50 {
		t.Errorf("brightness = %d, want 50", st.Brightness)
	}
}

func TestRegistryUnknownDriver(t *testing.T) {
	if _, err := NewLight("banana", "x"); err == nil {
		t.Error("expected error for unknown driver")
	}
}
