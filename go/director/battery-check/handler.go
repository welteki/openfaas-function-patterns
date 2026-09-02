package function

import (
	"encoding/json"
	"net/http"
)

const thresholdPercent = 20

type Reading struct {
	BatteryPercent int `json:"battery_percent"`
}

type Response struct {
	ValuePercent     int  `json:"value_percent"`
	ThresholdPercent int  `json:"threshold_percent"`
	Alert            bool `json:"alert"`
}

func Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var reading Reading
	if err := json.NewDecoder(r.Body).Decode(&reading); err != nil {
		http.Error(w, "expected a JSON sensor reading", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{
		ValuePercent:     reading.BatteryPercent,
		ThresholdPercent: thresholdPercent,
		Alert:            reading.BatteryPercent < thresholdPercent,
	})
}
