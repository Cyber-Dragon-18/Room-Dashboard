package devices

// TuyaLight — driver for Tuya-based lights (your Hama light).
//
// It translates the generic Light methods into Tuya cloud commands:
//
//	SetPower      -> switch_led
//	SetBrightness -> bright_value_v2 (white mode) or the "v" of colour_data_v2 (colour mode)
//	SetColor      -> work_mode "colour" + colour_data_v2 (HSV, RGB LEDs)
//	SetWhite      -> work_mode "white" + temp_value_v2 (real white LEDs)
//	State         -> GET .../status, converted back into a LightState
//
// Everything lowercase in this file is internal: only TuyaLight, NewTuyaLight
// and the four Light methods are visible outside the package.

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Central Europe data center — must match your Tuya project.
const tuyaBaseURL = "https://openapi.tuyaeu.com"

// Tuya error code meaning "token invalid / expired".
const tuyaCodeTokenInvalid = 1010

// ---------------------------------------------------------------------------
// TuyaLight
// ---------------------------------------------------------------------------

type TuyaLight struct {
	client   *tuyaClient
	deviceID string

	// Cached colour, so SetBrightness can dim a coloured light
	// without turning it white. Updated by SetColor.
	mu      sync.Mutex
	mode    string    // "white" or "colour"
	h, s, v int       // last HSV sent (Tuya ranges: h 0-360, s/v 0-1000)
	lastCmd time.Time // when we last sent a command
}

// Tuya's status lags a moment behind commands. For this long after a
// command, State() trusts our cache instead of the (possibly stale) status.
const tuyaStatusLag = 5 * time.Second

// send sends commands and remembers when, so State() knows the
// status it reads right after may still be the old one.
func (t *TuyaLight) send(cmds ...tuyaCommand) error {
	err := t.client.sendCommands(t.deviceID, cmds...)
	if err == nil {
		t.mu.Lock()
		t.lastCmd = time.Now()
		t.mu.Unlock()
	}
	return err
}

// NewTuyaLight creates a driver for one Tuya device.
// id and secret are the project's Access ID / Access Secret.
func NewTuyaLight(id, secret, deviceID string) *TuyaLight {
	return &TuyaLight{
		client: &tuyaClient{
			id:     id,
			secret: secret,
			http:   &http.Client{Timeout: 10 * time.Second},
		},
		deviceID: deviceID,
		mode:     "white",
	}
}

func (t *TuyaLight) SetPower(on bool) error {
	return t.send(tuyaCommand{"switch_led", on})
}

func (t *TuyaLight) SetBrightness(pct int) error {
	if pct < 0 || pct > 100 {
		return fmt.Errorf("brightness must be 0-100, got %d", pct)
	}

	t.mu.Lock()
	mode, h, s := t.mode, t.h, t.s
	t.mu.Unlock()

	// In colour mode, brightness is the "v" of the colour.
	if mode == "colour" {
		v := pctToTuya(pct)
		err := t.send(
			tuyaCommand{"colour_data_v2", map[string]int{"h": h, "s": s, "v": v}},
		)
		if err == nil {
			t.mu.Lock()
			t.v = v
			t.mu.Unlock()
		}
		return err
	}

	// In white mode, it's bright_value_v2.
	err := t.send(
		tuyaCommand{"work_mode", "white"},
		tuyaCommand{"bright_value_v2", pctToTuya(pct)},
	)
	if err == nil {
		t.mu.Lock()
		t.mode = "white"
		t.mu.Unlock()
	}
	return err
}

func (t *TuyaLight) SetColor(r, g, b uint8) error {
	h, s, v := rgbToTuyaHSV(r, g, b)
	err := t.send(
		tuyaCommand{"work_mode", "colour"},
		tuyaCommand{"colour_data_v2", map[string]int{"h": h, "s": s, "v": v}},
	)
	if err == nil {
		t.mu.Lock()
		t.mode, t.h, t.s, t.v = "colour", h, s, v
		t.mu.Unlock()
	}
	return err
}

func (t *TuyaLight) SetWhite(tempPct int) error {
	if tempPct < 0 || tempPct > 100 {
		return fmt.Errorf("temperature must be 0-100, got %d", tempPct)
	}
	err := t.send(
		tuyaCommand{"work_mode", "white"},
		tuyaCommand{"temp_value_v2", tempPct * 10}, // 0-100 -> 0-1000
	)
	if err == nil {
		t.mu.Lock()
		t.mode = "white"
		t.mu.Unlock()
	}
	return err
}

func (t *TuyaLight) State() (LightState, error) {
	resp, err := t.client.do("GET", "/v1.0/iot-03/devices/"+t.deviceID+"/status", nil)
	if err != nil {
		return LightState{}, err
	}

	// The status is a list like [{"code":"switch_led","value":true}, ...]
	var items []struct {
		Code  string          `json:"code"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(resp.Result, &items); err != nil {
		return LightState{}, err
	}

	var (
		st     LightState
		mode   = "white"
		bright = 1000
		temp   = 500
		hsv    tuyaHSV
		hasHSV bool
	)
	for _, it := range items {
		switch it.Code {
		case "switch_led":
			json.Unmarshal(it.Value, &st.On)
		case "work_mode":
			json.Unmarshal(it.Value, &mode)
		case "bright_value_v2":
			json.Unmarshal(it.Value, &bright)
		case "temp_value_v2":
			json.Unmarshal(it.Value, &temp)
		case "colour_data_v2":
			if c, ok := parseTuyaHSV(it.Value); ok {
				hsv, hasHSV = c, true
			}
		}
	}

	st.Temp = temp / 10 // 0-1000 -> 0-100

	// The lamp remembers its last colour even in white mode, so always
	// report it: switching back to the Colour tab then restores that colour.
	st.R, st.G, st.B = 255, 255, 255
	if hasHSV {
		st.R, st.G, st.B = tuyaHSVToRGB(hsv.H, hsv.S, 1000)
	}

	// Each mode has its OWN brightness on the lamp.
	if mode == "colour" && hasHSV {
		st.Mode = "colour"
		st.Brightness = tuyaToPct(hsv.V)
	} else {
		st.Mode = "white"
		st.Brightness = tuyaToPct(bright)
	}

	// Only refresh the cache from the status when no command was sent
	// recently — otherwise the status may still show the old state, and
	// caching it would make SetBrightness switch a coloured light to white.
	t.mu.Lock()
	if time.Since(t.lastCmd) > tuyaStatusLag {
		if mode == "colour" && hasHSV {
			t.mode, t.h, t.s, t.v = "colour", hsv.H, hsv.S, hsv.V
		} else {
			t.mode = "white"
		}
	}
	t.mu.Unlock()
	return st, nil
}

// Compile-time check: TuyaLight must implement Light.
var _ Light = (*TuyaLight)(nil)

// ---------------------------------------------------------------------------
// Conversions
// ---------------------------------------------------------------------------

// pctToTuya maps 0-100 % to Tuya's 10-1000 range (10 is the minimum).
func pctToTuya(pct int) int {
	return 10 + pct*990/100
}

// tuyaToPct maps Tuya's 10-1000 back to 0-100 %.
func tuyaToPct(v int) int {
	if v <= 10 {
		return 0
	}
	return int(math.Round(float64(v-10) * 100 / 990))
}

// rgbToTuyaHSV converts RGB (0-255) to Tuya HSV (h 0-360, s and v 0-1000).
// Standard RGB->HSV formula, with s and v scaled from 0-1 to 0-1000.
// Test: (255,0,0) -> (0,1000,1000), (0,0,255) -> (240,1000,1000)
func rgbToTuyaHSV(r, g, b uint8) (h, s, v int) {
	rf, gf, bf := float64(r)/255, float64(g)/255, float64(b)/255
	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	delta := max - min

	var hue float64
	switch {
	case delta == 0:
		hue = 0
	case max == rf:
		hue = 60 * math.Mod((gf-bf)/delta, 6)
	case max == gf:
		hue = 60 * ((bf-rf)/delta + 2)
	default:
		hue = 60 * ((rf-gf)/delta + 4)
	}
	if hue < 0 {
		hue += 360
	}

	var sat float64
	if max > 0 {
		sat = delta / max
	}

	h = int(math.Round(hue)) % 360
	s = int(math.Round(sat * 1000))
	v = int(math.Round(max * 1000))
	if v < 10 {
		v = 10 // Tuya's minimum brightness
	}
	return h, s, v
}

// tuyaHSVToRGB converts Tuya HSV back to RGB (inverse of the above).
func tuyaHSVToRGB(h, s, v int) (r, g, b uint8) {
	hf := float64(h)
	sf := float64(s) / 1000
	vf := float64(v) / 1000

	c := vf * sf
	x := c * (1 - math.Abs(math.Mod(hf/60, 2)-1))
	m := vf - c

	var rf, gf, bf float64
	switch {
	case hf < 60:
		rf, gf, bf = c, x, 0
	case hf < 120:
		rf, gf, bf = x, c, 0
	case hf < 180:
		rf, gf, bf = 0, c, x
	case hf < 240:
		rf, gf, bf = 0, x, c
	case hf < 300:
		rf, gf, bf = x, 0, c
	default:
		rf, gf, bf = c, 0, x
	}
	to8 := func(f float64) uint8 { return uint8(math.Round((f + m) * 255)) }
	return to8(rf), to8(gf), to8(bf)
}

type tuyaHSV struct {
	H int `json:"h"`
	S int `json:"s"`
	V int `json:"v"`
}

// parseTuyaHSV reads colour_data_v2 from a status response.
// Tuya sometimes sends it as an object, sometimes as a JSON string
// containing the object — this handles both.
func parseTuyaHSV(raw json.RawMessage) (tuyaHSV, bool) {
	var c tuyaHSV
	if err := json.Unmarshal(raw, &c); err == nil {
		return c, true
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if err := json.Unmarshal([]byte(s), &c); err == nil {
			return c, true
		}
	}
	return c, false
}

// ---------------------------------------------------------------------------
// Tuya cloud client (from your Phase 0 test, plus token refresh)
// ---------------------------------------------------------------------------

type tuyaClient struct {
	id     string
	secret string
	http   *http.Client

	mu        sync.Mutex // protects token and expiresAt (many requests can run at once)
	token     string
	expiresAt time.Time
}

type tuyaResponse struct {
	Success bool            `json:"success"`
	Code    int             `json:"code"`
	Msg     string          `json:"msg"`
	Result  json.RawMessage `json:"result"`
}

// tuyaError keeps Tuya's error code so we can react to specific errors.
type tuyaError struct {
	Code int
	Msg  string
}

func (e *tuyaError) Error() string { return fmt.Sprintf("tuya error %d: %s", e.Code, e.Msg) }

type tuyaCommand struct {
	Code  string `json:"code"`
	Value any    `json:"value"`
}

func (c *tuyaClient) sendCommands(deviceID string, cmds ...tuyaCommand) error {
	body, err := json.Marshal(map[string]any{"commands": cmds})
	if err != nil {
		return err
	}
	_, err = c.do("POST", "/v1.0/iot-03/devices/"+deviceID+"/commands", body)
	return err
}

// do makes sure there is a valid token, sends the request, and retries
// once with a fresh token if Tuya says the token is invalid.
func (c *tuyaClient) do(method, path string, body []byte) (*tuyaResponse, error) {
	token, err := c.validToken()
	if err != nil {
		return nil, fmt.Errorf("get token: %w", err)
	}

	resp, err := c.signedRequest(method, path, body, token)

	var te *tuyaError
	if errors.As(err, &te) && te.Code == tuyaCodeTokenInvalid {
		c.mu.Lock()
		c.token = "" // force a refresh
		c.mu.Unlock()
		if token, err = c.validToken(); err != nil {
			return nil, fmt.Errorf("refresh token: %w", err)
		}
		return c.signedRequest(method, path, body, token)
	}
	return resp, err
}

// validToken returns the cached token, or fetches a new one if it is
// missing or expires within the next minute.
func (c *tuyaClient) validToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Until(c.expiresAt) > time.Minute {
		return c.token, nil
	}

	// Token request is signed WITHOUT an access token.
	resp, err := c.signedRequest("GET", "/v1.0/token?grant_type=1", nil, "")
	if err != nil {
		return "", err
	}
	var result struct {
		AccessToken string `json:"access_token"`
		ExpireTime  int    `json:"expire_time"` // seconds
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return "", err
	}
	c.token = result.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(result.ExpireTime) * time.Second)
	return c.token, nil
}

// signedRequest is your Phase 0 do(): sign, send, decode.
//
//	stringToSign = METHOD + "\n" + sha256(body) + "\n" + "" + "\n" + path
//	sign         = HMAC(client_id + access_token + t + stringToSign, secret)
func (c *tuyaClient) signedRequest(method, path string, body []byte, token string) (*tuyaResponse, error) {
	t := strconv.FormatInt(time.Now().UnixMilli(), 10)
	stringToSign := method + "\n" + sha256Hex(string(body)) + "\n" + "" + "\n" + path
	sign := hmacUpperHex(c.id+token+t+stringToSign, c.secret)

	req, err := http.NewRequest(method, tuyaBaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("client_id", c.id)
	req.Header.Set("sign", sign)
	req.Header.Set("t", t)
	req.Header.Set("sign_method", "HMAC-SHA256")
	if token != "" {
		req.Header.Set("access_token", token)
	}
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var out tuyaResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode response: %w (body: %s)", err, raw)
	}
	if !out.Success {
		return nil, &tuyaError{Code: out.Code, Msg: out.Msg}
	}
	return &out, nil
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func hmacUpperHex(msg, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	return strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
}
