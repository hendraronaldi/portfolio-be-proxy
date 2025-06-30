package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"portfolio-be-proxy/handlers/agents"
)

func QueryCVHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query, ok := requestBody["query"].(string)
	if !ok || query == "" {
		http.Error(w, "Missing 'query' in request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received query: %s", query)

	ragResponse, err := agents.ResumeAgent(query)
	if err != nil {
		log.Printf("Error calling Resume Agent: %v", err)
		http.Error(w, fmt.Sprintf("Error processing request: %v", err), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(ragResponse)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
		http.Error(w, "Error processing response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(resp); err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
	log.Printf("Response sent successfully for query: %s", query)
}
