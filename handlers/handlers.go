package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"portfolio-be-proxy/handlers/agents"
	models_agents "portfolio-be-proxy/models/agents"
)

const payloadTooLargeMsg = "payload exceeds raw byte limit"

func requirePost(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return true
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return true
	}
	return false
}

func isBodyTooLarge(err error) bool {
	var maxErr *http.MaxBytesError
	return errors.As(err, &maxErr)
}

func readRawBody(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		if isBodyTooLarge(err) {
			http.Error(w, payloadTooLargeMsg, http.StatusRequestEntityTooLarge)
			return nil, false
		}
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return nil, false
	}
	return rawBody, true
}

func forwardBuildDate(w http.ResponseWriter, header http.Header) {
	if header != nil {
		if buildDate := header.Get("X-Index-Build-Date"); buildDate != "" {
			w.Header().Set("X-Index-Build-Date", buildDate)
		}
	}
}

func forwardSessionID(w http.ResponseWriter, r *http.Request) {
	if sid := r.Header.Get("X-Session-Id"); sid != "" {
		w.Header().Set("X-Session-Id", sid)
	}
}

func QueryCVHandler(w http.ResponseWriter, r *http.Request) {
	if requirePost(w, r) {
		return
	}
	rawBody, ok := readRawBody(w, r)
	if !ok {
		return
	}
	var probe models_agents.ResumeAgentRequest
	if err := json.NewDecoder(bytes.NewReader(rawBody)).Decode(&probe); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if probe.Query == "" {
		http.Error(w, "Missing 'query'", http.StatusBadRequest)
		return
	}
	statusCode, upstreamBody, header, err := agents.ResumeAgent(rawBody, r.Header.Get("X-User-Id"), r.Header.Get("X-Session-Id"))
	if err != nil {
		log.Printf("Error calling Resume Agent: %v", err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	forwardBuildDate(w, header)
	forwardSessionID(w, r)
	w.WriteHeader(statusCode)
	if _, err := w.Write(upstreamBody); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func FeedbackHandler(w http.ResponseWriter, r *http.Request) {
	if requirePost(w, r) {
		return
	}
	rawBody, ok := readRawBody(w, r)
	if !ok {
		return
	}
	statusCode, upstreamBody, header, err := agents.FeedbackAgent(rawBody, r.Header.Get("X-User-Id"), r.Header.Get("X-Session-Id"))
	if err != nil {
		log.Printf("Error calling Feedback endpoint: %v", err)
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	forwardBuildDate(w, header)
	forwardSessionID(w, r)
	w.WriteHeader(statusCode)
	if _, err := w.Write(upstreamBody); err != nil {
		log.Printf("Error writing feedback response: %v", err)
	}
}
