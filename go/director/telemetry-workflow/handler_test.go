package function

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestHandleCombinesParallelChecks(t *testing.T) {
	started := make(chan string, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/function/validate-reading":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"device_id":"pump-17","temperature_c":82.4,"battery_percent":12}`))
		case "/function/temperature-check":
			started <- "temperature-check"
			<-release
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"value_c":82.4,"threshold_c":75,"alert":true}`))
		case "/function/battery-check":
			started <- "battery-check"
			<-release
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"value_percent":12,"threshold_percent":20,"alert":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("gateway_url", server.URL)
	t.Setenv("stage_timeout", "1s")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"device_id":"pump-17","temperature_c":82.4,"battery_percent":12}`,
	))
	recorder := httptest.NewRecorder()
	done := make(chan struct{})

	go func() {
		Handle(recorder, req)
		close(done)
	}()

	seen := map[string]bool{}
	for len(seen) < 2 {
		select {
		case function := <-started:
			seen[function] = true
		case <-time.After(time.Second):
			t.Fatal("parallel checks did not start together")
		}
	}

	releaseOnce.Do(func() { close(release) })
	<-done

	if recorder.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Status != "alert" {
		t.Fatalf("want alert status, got %q", response.Status)
	}
	if !response.Temperature.Alert || !response.Battery.Alert {
		t.Fatalf("want both checks to alert: %+v", response)
	}
}

func TestHandleReturnsOK(t *testing.T) {
	server := newStageServer(t, false)
	defer server.Close()

	t.Setenv("gateway_url", server.URL)
	t.Setenv("stage_timeout", "1s")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"device_id":"pump-17","temperature_c":48.2,"battery_percent":78}`,
	))
	recorder := httptest.NewRecorder()
	Handle(recorder, req)

	var response Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Status != "ok" {
		t.Fatalf("want ok status, got %q", response.Status)
	}
}

func TestHandleStopsAfterValidationFailure(t *testing.T) {
	var checks atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/function/validate-reading" {
			http.Error(w, "device_id is required", http.StatusBadRequest)
			return
		}
		checks.Add(1)
	}))
	defer server.Close()

	t.Setenv("gateway_url", server.URL)
	t.Setenv("stage_timeout", "1s")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"temperature_c":20}`))
	recorder := httptest.NewRecorder()
	Handle(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("want status 400, got %d", recorder.Code)
	}
	if checks.Load() != 0 {
		t.Fatalf("want no checks after validation failure, got %d", checks.Load())
	}
}

func TestHandleTimesOutDownstreamCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/function/validate-reading" {
			w.Write([]byte(`{"device_id":"pump-17","temperature_c":48.2,"battery_percent":78}`))
			return
		}
		time.Sleep(50 * time.Millisecond)
		w.Write([]byte(`{"alert":false}`))
	}))
	defer server.Close()

	t.Setenv("gateway_url", server.URL)
	t.Setenv("stage_timeout", "5ms")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"device_id":"pump-17","temperature_c":48.2,"battery_percent":78}`,
	))
	recorder := httptest.NewRecorder()
	Handle(recorder, req)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("want status 502, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "Client.Timeout") {
		t.Fatalf("want timeout error, got %q", recorder.Body.String())
	}
}

func newStageServer(t *testing.T, alert bool) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/function/validate-reading":
			w.Write([]byte(`{"device_id":"pump-17","temperature_c":48.2,"battery_percent":78}`))
		case "/function/temperature-check":
			json.NewEncoder(w).Encode(TemperatureResult{ValueC: 48.2, ThresholdC: 75, Alert: alert})
		case "/function/battery-check":
			json.NewEncoder(w).Encode(BatteryResult{ValuePercent: 78, ThresholdPercent: 20, Alert: alert})
		default:
			http.NotFound(w, r)
		}
	}))
}
