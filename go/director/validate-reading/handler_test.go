package function

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleValidReading(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"device_id":" pump-17 ","temperature_c":48.2,"battery_percent":78}`,
	))
	recorder := httptest.NewRecorder()
	Handle(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var reading Reading
	if err := json.Unmarshal(recorder.Body.Bytes(), &reading); err != nil {
		t.Fatal(err)
	}
	if reading.DeviceID != "pump-17" {
		t.Fatalf("want trimmed device ID, got %q", reading.DeviceID)
	}
}

func TestHandleRejectsInvalidBattery(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		`{"device_id":"pump-17","temperature_c":48.2,"battery_percent":101}`,
	))
	recorder := httptest.NewRecorder()
	Handle(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("want status 400, got %d", recorder.Code)
	}
}
