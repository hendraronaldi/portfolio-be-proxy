package agents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"portfolio-be-proxy/config"
	models_agents "portfolio-be-proxy/models/agents"
)

func ResumeAgent(query string) (models_agents.ResumeAgentResponse, int, error) {
	var responseBody models_agents.ResumeAgentResponse
	if config.Config.ResumeAgentURL == "" {
		return responseBody, 500, fmt.Errorf("RESUME_AGENT_URL is not set")
	}

	requestBody, err := json.Marshal(models_agents.ResumeAgentRequest{
		Query: query,
	})
	if err != nil {
		return responseBody, 500, fmt.Errorf("failed to marshal request body: %w", err)
	}

	resp, err := http.Post(config.Config.ResumeAgentURL+"/query-resume", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return responseBody, resp.StatusCode, fmt.Errorf("failed to call Resume Agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return responseBody, resp.StatusCode, fmt.Errorf("Resume Agent returned non-200 status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return responseBody, resp.StatusCode, fmt.Errorf("failed to decode Resume Agent response: %w", err)
	}

	return responseBody, resp.StatusCode, nil
}
