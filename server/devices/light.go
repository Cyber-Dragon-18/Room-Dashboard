package devices

type Light interface {
	SetPower(on bool) error
	SetBrightness(pct int) error
	SetColor(r, g, b uint8) error
	State() (LightState, error)
}

type LightState struct {
	On         bool  `json:"on"`
	Brightness int   `json:"brightness"`
	R          uint8 `json:"r"`
	G          uint8 `json:"g"`
	B          uint8 `json:"b"`
}
