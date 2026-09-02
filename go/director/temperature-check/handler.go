package function

import (
	"encoding/json"
	"net/http"
)

const thresholdC = 75.0

type Reading struct {
	TemperatureC float64 `json:"temperature_c"`
}

type Response struct {
	ValueC     float64 `json:"value_c"`
	ThresholdC float64 `json:"threshold_c"`
	Alert      bool    `json:"alert"`
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
		ValueC:     reading.TemperatureC,
		ThresholdC: thresholdC,
		Alert:      reading.TemperatureC > thresholdC,
	})
}
