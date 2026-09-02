package function

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultStageTimeout = 5 * time.Second

type Reading struct {
	DeviceID       string  `json:"device_id"`
	TemperatureC   float64 `json:"temperature_c"`
	BatteryPercent int     `json:"battery_percent"`
}

type TemperatureResult struct {
	ValueC     float64 `json:"value_c"`
	ThresholdC float64 `json:"threshold_c"`
	Alert      bool    `json:"alert"`
}

type BatteryResult struct {
	ValuePercent     int  `json:"value_percent"`
	ThresholdPercent int  `json:"threshold_percent"`
	Alert            bool `json:"alert"`
}

type Response struct {
	DeviceID    string            `json:"device_id"`
	Status      string            `json:"status"`
	Temperature TemperatureResult `json:"temperature"`
	Battery     BatteryResult     `json:"battery"`
	DurationMs  int64             `json:"duration_ms"`
}

type callResult struct {
	function string
	body     []byte
	status   int
	err      error
}

func Handle(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	input, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	timeout, err := configuredStageTimeout()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gateway := os.Getenv("gateway_url")
	if gateway == "" {
		gateway = "http://gateway.openfaas:8080"
	}

	client := &http.Client{Timeout: timeout}

	validated, status, err := invoke(
		r.Context(), client, gateway, "validate-reading", input,
	)
	if err != nil {
		message := fmt.Sprintf("failed to invoke validate-reading: %s", err)
		http.Error(w, message, http.StatusBadGateway)
		return
	}
	if status != http.StatusOK {
		message := fmt.Sprintf("validate-reading failed: %s", validated)
		http.Error(w, message, status)
		return
	}

	var reading Reading
	if err := json.Unmarshal(validated, &reading); err != nil {
		message := fmt.Sprintf(
			"unexpected response from validate-reading: %s",
			err,
		)
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	functions := []string{"temperature-check", "battery-check"}
	results := make(chan callResult, len(functions))

	for _, function := range functions {
		go func(name string) {
			body, status, err := invoke(
				r.Context(), client, gateway, name, validated,
			)
			results <- callResult{
				function: name,
				body:     body,
				status:   status,
				err:      err,
			}
		}(function)
	}

	completed := make(map[string]callResult, len(functions))
	for range functions {
		result := <-results
		completed[result.function] = result
	}

	for _, function := range functions {
		result := completed[function]
		if result.err != nil {
			message := fmt.Sprintf(
				"failed to invoke %s: %s",
				function,
				result.err,
			)
			http.Error(w, message, http.StatusBadGateway)
			return
		}
		if result.status != http.StatusOK {
			message := fmt.Sprintf("%s failed: %s", function, result.body)
			http.Error(w, message, result.status)
			return
		}
	}

	var temperature TemperatureResult
	if err := json.Unmarshal(
		completed["temperature-check"].body,
		&temperature,
	); err != nil {
		message := fmt.Sprintf(
			"unexpected response from temperature-check: %s",
			err,
		)
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	var battery BatteryResult
	if err := json.Unmarshal(
		completed["battery-check"].body,
		&battery,
	); err != nil {
		message := fmt.Sprintf(
			"unexpected response from battery-check: %s",
			err,
		)
		http.Error(w, message, http.StatusBadGateway)
		return
	}

	workflowStatus := "ok"
	if temperature.Alert || battery.Alert {
		workflowStatus = "alert"
	}

	response := Response{
		DeviceID:    reading.DeviceID,
		Status:      workflowStatus,
		Temperature: temperature,
		Battery:     battery,
		DurationMs:  time.Since(start).Milliseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func configuredStageTimeout() (time.Duration, error) {
	value := os.Getenv("stage_timeout")
	if value == "" {
		return defaultStageTimeout, nil
	}

	timeout, err := time.ParseDuration(value)
	if err != nil || timeout <= 0 {
		return 0, fmt.Errorf("invalid stage_timeout %q", value)
	}

	return timeout, nil
}

func invoke(
	ctx context.Context,
	client *http.Client,
	gateway string,
	function string,
	body []byte,
) ([]byte, int, error) {
	url := strings.TrimRight(gateway, "/") + "/function/" + function
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("create request for %s: %w", function, err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("invoke %s: %w", function, err)
	}
	defer res.Body.Close()

	out, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("read response from %s: %w", function, err)
	}

	return out, res.StatusCode, nil
}
