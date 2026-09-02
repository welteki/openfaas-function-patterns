package function

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleTemperatureThreshold(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		alert bool
	}{
		{name: "below", value: 74.9, alert: false},
		{name: "at threshold", value: 75, alert: false},
		{name: "above", value: 75.1, alert: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
				`{"temperature_c":`+jsonNumber(test.value)+`}`,
			))
			recorder := httptest.NewRecorder()
			Handle(recorder, req)

			var response Response
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Alert != test.alert {
				t.Fatalf("want alert %t, got %t", test.alert, response.Alert)
			}
		})
	}
}

func jsonNumber(value float64) string {
	data, _ := json.Marshal(value)
	return string(data)
}
