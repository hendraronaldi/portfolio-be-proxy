package models_agents

type ResumeAgentResponse struct {
	Query   string `json:"query"`
	Message string `json:"message"`
}

type ResumeAgentRequest struct {
	Query string `json:"query"`
}
