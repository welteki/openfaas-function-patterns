package function

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

var (
	subscribersMu sync.Mutex
	subscribers   = make(map[chan string]struct{})
)

func Handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		subscribe(w, r)
	case http.MethodPost:
		publish(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func subscribe(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(
			w,
			"streaming is not supported",
			http.StatusInternalServerError,
		)
		return
	}

	messages := make(chan string, 1)
	subscribersMu.Lock()
	subscribers[messages] = struct{}{}
	subscribersMu.Unlock()

	defer func() {
		subscribersMu.Lock()
		delete(subscribers, messages)
		subscribersMu.Unlock()
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case message := <-messages:
			fmt.Fprintf(w, "data: %s\n\n", message)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func publish(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return
	}

	message := strings.TrimSpace(string(body))
	if message == "" {
		http.Error(w, "notification must not be empty", http.StatusBadRequest)
		return
	}

	message = strings.NewReplacer("\r", " ", "\n", " ").Replace(message)
	delivered := broadcast(message)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"delivered\":%d}\n", delivered)
}

func broadcast(message string) int {
	subscribersMu.Lock()
	defer subscribersMu.Unlock()

	delivered := 0
	for subscriber := range subscribers {
		select {
		case subscriber <- message:
			delivered++
		default:
			// Do not let a slow subscriber block every other client.
		}
	}

	return delivered
}
