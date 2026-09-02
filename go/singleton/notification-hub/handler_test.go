package function

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublishBroadcastsNotification(t *testing.T) {
	resetSubscribers()

	messages := make(chan string, 1)
	subscribers[messages] = struct{}{}

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("deployment complete"))
	recorder := httptest.NewRecorder()
	Handle(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d", recorder.Code)
	}
	if got := recorder.Body.String(); got != "{\"delivered\":1}\n" {
		t.Fatalf("unexpected response: %q", got)
	}
	if got := <-messages; got != "deployment complete" {
		t.Fatalf("unexpected notification: %q", got)
	}
}

func TestPublishRejectsEmptyNotification(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("  "))
	recorder := httptest.NewRecorder()
	Handle(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("want status 400, got %d", recorder.Code)
	}
}

func resetSubscribers() {
	subscribersMu.Lock()
	defer subscribersMu.Unlock()
	subscribers = make(map[chan string]struct{})
}
