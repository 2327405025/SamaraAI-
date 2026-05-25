package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func WriteData(w http.ResponseWriter, flusher http.Flusher, payload string) {
	for _, line := range strings.Split(payload, "\n") {
		fmt.Fprintf(w, "data: %s\n", line)
	}
	fmt.Fprint(w, "\n")
	flusher.Flush()
}

func WriteJSON(w http.ResponseWriter, flusher http.Flusher, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "data: %s\n\n", b)
	flusher.Flush()
	return nil
}

func WriteDone(w http.ResponseWriter, flusher http.Flusher) {
	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}
