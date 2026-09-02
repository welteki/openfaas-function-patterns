package function

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Reading struct {
	DeviceID       string  `json:"device_id"`
	TemperatureC   float64 `json:"temperature_c"`
	BatteryPercent int     `json:"battery_percent"`
}

func Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var reading Reading
	if err := json.NewDecoder(r.Body).Decode(&reading); err != nil {
		http.Error(w, "expected a JSON sensor reading", http.StatusBadRequest)
		return
	}

	reading.DeviceID = strings.TrimSpace(reading.DeviceID)
	if reading.DeviceID == "" {
		http.Error(w, "device_id is required", http.StatusBadRequest)
		return
	}
	if reading.TemperatureC < -100 || reading.TemperatureC > 200 {
		http.Error(
			w,
			"temperature_c must be between -100 and 200",
			http.StatusBadRequest,
		)
		return
	}
	if reading.BatteryPercent < 0 || reading.BatteryPercent > 100 {
		http.Error(
			w,
			"battery_percent must be between 0 and 100",
			http.StatusBadRequest,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reading)
}
