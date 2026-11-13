// Tuya light test — Phase 0
//
// What it does:
//   1. Gets an access token from Tuya's cloud
//   2. Turns the light on, makes it red, then warm white, then turns it off
//
// How to run (in this folder):
//   go mod init tuya-test        (only the first time)
//   export TUYA_ID="your Access ID"
//   export TUYA_SECRET="your Access Secret"
//   export TUYA_DEVICE="your Device ID"
//   go run .
//
// On Windows PowerShell use:  $env:TUYA_ID="..."  instead of export.

package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Central Europe data center. Must match the one chosen in your Tuya project.
const baseURL = "https://openapi.tuyaeu.com"

// ---------------------------------------------------------------------------
// Helpers for the signature
// ---------------------------------------------------------------------------

// sha256Hex returns the SHA-256 hash of s as lowercase hex.
// Tuya uses it to hash the request body (an empty body is hashed as "").
// Self-test: sha256Hex("") == "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// hmacUpperHex signs msg with the secret using HMAC-SHA256.
// Tuya wants the result in UPPERCASE hex.
func hmacUpperHex(msg, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	return strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
}

// ---------------------------------------------------------------------------
// A tiny Tuya client
// ---------------------------------------------------------------------------

type tuyaClient struct {
	id     string // Access ID (Tuya calls it client_id)
	secret string // Access Secret — never print or commit this
	token  string // empty until getToken() runs
	http   *http.Client
}

// Every Tuya response has this shape. "Result" changes per endpoint,
// so we keep it as raw JSON and decode it later into the right struct.
type apiResponse struct {
	Success bool            `json:"success"`
	Code    int             `json:"code"`
	Msg     string          `json:"msg"`
	Result  json.RawMessage `json:"result"`
}

// do sends one signed request to Tuya and returns the decoded response.
//
// The signature works like this:
//
//	stringToSign = METHOD + "\n" + sha256(body) + "\n" + "" + "\n" + path
//	sign         = HMAC(client_id + access_token + t + stringToSign, secret)
//
// For the token request, access_token is still empty, so it just disappears
// from the string — that is why the same function works for both calls.
func (c *tuyaClient) do(method, path string, body []byte) (*apiResponse, error) {
	// Timestamp in MILLISECONDS, as a string
	t := strconv.FormatInt(time.Now().UnixMilli(), 10)

	// The empty "" line is for optional signed headers, which we don't use
	stringToSign := method + "\n" + sha256Hex(string(body)) + "\n" + "" + "\n" + path
	sign := hmacUpperHex(c.id+c.token+t+stringToSign, c.secret)

	// Uncomment to debug "sign invalid" errors (%q shows the \n characters):
	// fmt.Printf("stringToSign: %q\n", stringToSign)

	req, err := http.NewRequest(method, baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("client_id", c.id)
	req.Header.Set("sign", sign)
	req.Header.Set("t", t)
	req.Header.Set("sign_method", "HMAC-SHA256")
	if c.token != "" {
		req.Header.Set("access_token", c.token)
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

	var out apiResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("could not decode response: %w (body: %s)", err, raw)
	}
	// Tuya returns HTTP 200 even on errors; the real status is "success"
	if !out.Success {
		return nil, fmt.Errorf("tuya error %d: %s", out.Code, out.Msg)
	}
	return &out, nil
}

// getToken asks Tuya for an access token and stores it in the client.
// The token lasts about 2 hours; for this test that is plenty.
func (c *tuyaClient) getToken() error {
	resp, err := c.do("GET", "/v1.0/token?grant_type=1", nil)
	if err != nil {
		return err
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpireTime  int    `json:"expire_time"` // seconds
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return err
	}
	c.token = result.AccessToken
	fmt.Printf("Got token (expires in %d s)\n", result.ExpireTime)
	return nil
}

// A command is exactly what you typed in the API Explorer:
// {"code": "switch_led", "value": true}
// Value is "any" because it can be a bool, a number, a string or an object.
type command struct {
	Code  string `json:"code"`
	Value any    `json:"value"`
}

// sendCommands sends one or more commands to the device in a single request.
func (c *tuyaClient) sendCommands(deviceID string, cmds ...command) error {
	body, err := json.Marshal(map[string]any{"commands": cmds})
	if err != nil {
		return err
	}
	_, err = c.do("POST", "/v1.0/iot-03/devices/"+deviceID+"/commands", body)
	return err
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

// mustEnv reads an environment variable or stops the program with a clear message.
func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("missing environment variable %s", name)
	}
	return v
}

func main() {

	_ = godotenv.Load()

	c := &tuyaClient{
		id:     mustEnv("TUYA_ID"),
		secret: mustEnv("TUYA_SECRET"),
		http:   &http.Client{Timeout: 10 * time.Second},
	}
	device := mustEnv("TUYA_DEVICE")

	// 1. Token
	if err := c.getToken(); err != nil {
		log.Fatal("token: ", err)
	}

	// 2. On
	fmt.Println("Light off")
	if err := c.sendCommands(device, command{"switch_led", false}); err != nil {
		log.Fatal(err)
	}
	time.Sleep(2 * time.Second)
	/*
		// 3. Red — colour mode first, then HSV (h 0-360, s and v 0-1000)
		fmt.Println("Red")
		err := c.sendCommands(device,
			command{"work_mode", "colour"},
			command{"colour_data_v2", map[string]int{"h": 0, "s": 1000, "v": 1000}},
		)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(2 * time.Second)

		// 4. Warm white at 50% — brightness 10-1000, temperature 0 (warm) to 1000 (cool)
		fmt.Println("Warm white 50%")
		err = c.sendCommands(device,
			command{"work_mode", "white"},
			command{"bright_value_v2", 500},
			command{"temp_value_v2", 0},
		)
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(2 * time.Second)


		// 5. Off
		fmt.Println("Light off")
		if err := c.sendCommands(device, command{"switch_led", false}); err != nil {
			log.Fatal(err)
		}
		time.Sleep(2 * time.Second)

		// 6. On again
		fmt.Println("Light on")
		if err := c.sendCommands(device, command{"switch_led", true}); err != nil {
			log.Fatal(err)
		}
		time.Sleep(2 * time.Second)

		fmt.Println("Normal 100%")
		err = c.sendCommands(device,
			command{"work_mode", "white"},
			command{"bright_value_v2", 1000},
			command{"temp_value_v2", 850},
		)
		if err != nil {
			log.Fatal(err)
		}
	*/
	fmt.Println("Done — Phase 0 complete!")
}
