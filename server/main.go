package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Cyber-Dragon-18/room-dashboard/server/devices"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	driver := os.Getenv("DRIVER")
	deviceID := os.Getenv("TUYA_DEVICE")

	light, err := devices.NewLight(driver, deviceID)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/devices", handleListDevices)
	mux.HandleFunc("GET /api/device/{id}", handleGetDevice)
	mux.HandleFunc("POST /api/devices/{id}/power", handleSetPower)
	mux.HandleFunc("POST /api/devices/{id}/brightness", handleSetBrightness)
	mux.HandleFunc("POST /api/device/{id}/color", handleSetColour)


	func handleListDevices(w http.ResponseWriter, r *http.Request){

	}

	func handleGetDevice(w http.ResponseWriter, r *http.Request) {
	
	}

	func handleSetPower (w http.ResponseWriter, r *http.Request) {
		
	}

	func handleSetBrightness (w http.ResponseWriter, r*http.Request) {
		
	}

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))	


}
