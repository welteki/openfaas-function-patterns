package function

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestRecordsFromInput(t *testing.T) {
	records := recordsFromInput("https://one.example\n\n https://two.example \n")
	if len(records) != 2 {
		t.Fatalf("want 2 records, got %d", len(records))
	}
	if records[0] != "https://one.example" || records[1] != "https://two.example" {
		t.Fatalf("unexpected records: %#v", records)
	}
}

func TestHandleSubmitsEachURLWithCallback(t *testing.T) {
	var mu sync.Mutex
	var bodies []string
	var callbacks []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/async-function/url-check" {
			http.NotFound(w, r)
			return
		}

		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(body))
		callbacks = append(callbacks, r.Header.Get("X-Callback-Url"))
		callID := "call-" + string(rune('0'+len(bodies)))
		mu.Unlock()

		w.Header().Set("X-Call-Id", callID)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	t.Setenv("gateway_url", server.URL)
	t.Setenv("callback_url", "")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(
		"https://one.example\nhttps://two.example\n",
	))
	req.Header.Set("X-Callback-Url", "http://gateway.openfaas:8080/function/printer")
	recorder := httptest.NewRecorder()
	Handle(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Submitted != 2 || response.Function != "url-check" || !response.Callback {
		t.Fatalf("unexpected response: %+v", response)
	}
	if len(response.CallIDs) != 2 {
		t.Fatalf("want 2 call IDs, got %#v", response.CallIDs)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != 2 || bodies[0] != "https://one.example" || bodies[1] != "https://two.example" {
		t.Fatalf("unexpected bodies: %#v", bodies)
	}
	for _, callback := range callbacks {
		if callback != "http://gateway.openfaas:8080/function/printer" {
			t.Fatalf("callback was not forwarded: %q", callback)
		}
	}
}

func TestHandleReturnsSubmissionFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "queue unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	t.Setenv("gateway_url", server.URL)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://one.example\n"))
	recorder := httptest.NewRecorder()
	Handle(recorder, req)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("want status 502, got %d", recorder.Code)
	}
}
