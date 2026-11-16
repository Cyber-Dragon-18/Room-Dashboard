package main

import "github.com/Cyber-Dragon-18/room-dashboard/server/devices"

func main() {
	var l devices.Light = &devices.MockLight{}
	l.SetPower(true)
}
