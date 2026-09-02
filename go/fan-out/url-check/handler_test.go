package function

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandleHealthyURL(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()

	recorder := invokeURLCheck(t, target.URL)

	var response Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !response.Reachable || !response.Healthy || response.StatusCode != http.StatusNoContent {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestHandleUnhealthyStatus(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer target.Close()

	recorder := invokeURLCheck(t, target.URL)

	var response Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !response.Reachable || response.Healthy || response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestHandleRejectsInvalidURL(t *testing.T) {
	recorder := invokeURLCheck(t, "file:///etc/passwd")
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("want status 400, got %d", recorder.Code)
	}
}

func TestHandleReportsRequestTimeout(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	originalTimeout := requestTimeout
	requestTimeout = 5 * time.Millisecond
	t.Cleanup(func() { requestTimeout = originalTimeout })

	recorder := invokeURLCheck(t, target.URL)

	var response Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Reachable || response.Error == "" {
		t.Fatalf("want unreachable result with an error: %+v", response)
	}
}

func invokeURLCheck(t *testing.T, target string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(target))
	recorder := httptest.NewRecorder()
	Handle(recorder, req)
	return recorder
}
