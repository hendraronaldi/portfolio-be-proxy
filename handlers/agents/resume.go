package agents

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"portfolio-be-proxy/config"
)

// 60s generation budget: upstream RAG generation may take up to 60s.
var agentClient = &http.Client{Timeout: 60 * time.Second}

func requireBackendURL() error {
	if config.Config.ResumeAgentURL == "" {
		return fmt.Errorf("RESUME_AGENT_URL is not set")
	}
	return nil
}

func postRaw(path string, rawBody []byte, userID, sessionID string) (int, []byte, http.Header, error) {
	if err := requireBackendURL(); err != nil {
		return 500, nil, nil, err
	}
	req, err := http.NewRequest("POST", config.Config.ResumeAgentURL+path, bytes.NewReader(rawBody))
	if err != nil {
		return 500, nil, nil, fmt.Errorf("failed to build upstream request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if userID != "" {
		req.Header.Set("X-User-Id", userID)
	}
	if sessionID != "" {
		req.Header.Set("X-Session-Id", sessionID)
	}
	resp, err := agentClient.Do(req)
	if err != nil {
		return 500, nil, nil, fmt.Errorf("failed to call upstream: %w", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 500, nil, resp.Header, fmt.Errorf("failed to read upstream response: %w", err)
	}
	return resp.StatusCode, bodyBytes, resp.Header, nil
}

func ResumeAgent(rawBody []byte, userID, sessionID string) (int, []byte, http.Header, error) {
	return postRaw("/query-resume/", rawBody, userID, sessionID)
}

func FeedbackAgent(rawBody []byte, userID, sessionID string) (int, []byte, http.Header, error) {
	return postRaw("/feedback/", rawBody, userID, sessionID)
}
