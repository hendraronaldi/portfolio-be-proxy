package models_agents

type ResumeAgentRequest struct {
	Query   string   `json:"query"`
	History []string `json:"history,omitempty"`
}
