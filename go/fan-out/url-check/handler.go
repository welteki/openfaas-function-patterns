package function

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultRequestTimeout = 5 * time.Second
	maxURLLength          = 4096
)

var requestTimeout = defaultRequestTimeout

func init() {
	value := os.Getenv("request_timeout")
	if value == "" {
		return
	}

	timeout, err := time.ParseDuration(value)
	if err != nil || timeout <= 0 {
		panic(fmt.Sprintf("invalid request_timeout %q", value))
	}

	requestTimeout = timeout
}

type Response struct {
	URL         string `json:"url"`
	Reachable   bool   `json:"reachable"`
	Healthy     bool   `json:"healthy"`
	StatusCode  int    `json:"status_code,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	DurationMs  int64  `json:"duration_ms"`
	Error       string `json:"error,omitempty"`
}

func Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	input, err := io.ReadAll(io.LimitReader(r.Body, maxURLLength+1))
	if err != nil {
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return
	}
	if len(input) > maxURLLength {
		http.Error(w, "URL is too long", http.StatusBadRequest)
		return
	}

	target := strings.TrimSpace(string(input))
	parsed, err := url.ParseRequestURI(target)
	if err != nil || parsed.Host == "" {
		http.Error(
			w,
			"expected an absolute HTTP or HTTPS URL",
			http.StatusBadRequest,
		)
		return
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		http.Error(
			w,
			"expected an absolute HTTP or HTTPS URL",
			http.StatusBadRequest,
		)
		return
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		http.Error(
			w,
			"unable to create health-check request",
			http.StatusBadRequest,
		)
		return
	}
	req.Header.Set("User-Agent", "OpenFaaS URL health check")

	res, requestErr := http.DefaultClient.Do(req)
	result := Response{
		URL:        target,
		DurationMs: time.Since(start).Milliseconds(),
	}
	if requestErr != nil {
		result.Error = requestErr.Error()
		writeJSON(w, result)
		return
	}
	defer res.Body.Close()
	io.Copy(io.Discard, io.LimitReader(res.Body, 1024))

	result.Reachable = true
	result.Healthy = res.StatusCode >= 200 && res.StatusCode < 400
	result.StatusCode = res.StatusCode
	result.ContentType = res.Header.Get("Content-Type")
	writeJSON(w, result)
}

func writeJSON(w http.ResponseWriter, result Response) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
