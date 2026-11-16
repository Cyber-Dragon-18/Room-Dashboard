package devices

type Light interface {
	SetPower(on bool) error
	SetBrightness(pct int) error
	SetColor(r, g, b uint8) error
	State() (LightState, error)
}

type LightState struct {
	On         bool
	Brightness int
	R, G, B    uint8
}
