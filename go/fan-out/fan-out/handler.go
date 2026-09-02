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

const (
	targetFunction = "url-check"
	submitTimeout  = 30 * time.Second
)

type Response struct {
	Submitted int      `json:"submitted"`
	Function  string   `json:"function"`
	Callback  bool     `json:"callback"`
	CallIDs   []string `json:"call_ids,omitempty"`
}

// Handle takes a HTTP request body and splits it into one record per line.
// Each record is submitted as an asynchronous invocation of the target
// function, then a summary is returned to the caller without waiting for
// the function invocations to complete.
func Handle(w http.ResponseWriter, r *http.Request) {
	input, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	gateway := os.Getenv("gateway_url")
	if gateway == "" {
		gateway = "http://gateway.openfaas:8080"
	}

	// Forward the callback URL to every asynchronous invocation. A header on
	// the batch request overrides the environment variable.
	callback := strings.TrimSpace(r.Header.Get("X-Callback-Url"))
	if callback == "" {
		callback = strings.TrimSpace(os.Getenv("callback_url"))
	}

	records := recordsFromInput(string(input))
	if len(records) == 0 {
		http.Error(
			w,
			"expected one record per line in the request body",
			http.StatusBadRequest,
		)
		return
	}

	submitted := 0
	var callIDs []string

	for i, record := range records {
		callID, err := submit(
			r.Context(), gateway, targetFunction, record, callback,
		)
		if err != nil {
			message := fmt.Sprintf(
				"record %d of %d: %s",
				i+1,
				len(records),
				err,
			)
			http.Error(w, message, http.StatusBadGateway)
			return
		}

		submitted++
		if callID != "" {
			callIDs = append(callIDs, callID)
		}
	}

	res := Response{
		Submitted: submitted,
		Function:  targetFunction,
		Callback:  callback != "",
		CallIDs:   callIDs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func recordsFromInput(input string) []string {
	var records []string

	for _, record := range strings.Split(strings.TrimSpace(input), "\n") {
		record = strings.TrimSpace(record)
		if record != "" {
			records = append(records, record)
		}
	}

	return records
}

func submit(
	ctx context.Context,
	gateway string,
	targetFunction string,
	record string,
	callback string,
) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, submitTimeout)
	defer cancel()

	url := strings.TrimRight(gateway, "/") + "/async-function/" + targetFunction
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewReader([]byte(record)),
	)
	if err != nil {
		return "", fmt.Errorf("unable to invoke %s: %w", targetFunction, err)
	}
	req.Header.Set("Content-Type", "text/plain")
	if callback != "" {
		req.Header.Set("X-Callback-Url", callback)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error invoking %s: %w", targetFunction, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		out, err := io.ReadAll(res.Body)
		if err != nil {
			return "", fmt.Errorf(
				"unexpected status %d from %s",
				res.StatusCode,
				targetFunction,
			)
		}

		return "", fmt.Errorf(
			"unexpected status %d from %s: %s",
			res.StatusCode,
			targetFunction,
			string(out),
		)
	}

	// the X-Call-Id header can be used to track or cancel the record
	return res.Header.Get("X-Call-Id"), nil
}
