package function

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleBatteryThreshold(t *testing.T) {
	tests := []struct {
		name  string
		value int
		alert bool
	}{
		{name: "below", value: 19, alert: true},
		{name: "at threshold", value: 20, alert: false},
		{name: "above", value: 21, alert: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
				`{"battery_percent":`+jsonNumber(test.value)+`}`,
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

func jsonNumber(value int) string {
	data, _ := json.Marshal(value)
	return string(data)
}
