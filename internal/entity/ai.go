package entity

type AiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AiResponse struct {
	Reasoning    string `json:"reasoning"`
	Action       string `json:"action"`
	Target       string `json:"target,omitempty"`
	Text         string `json:"text,omitempty"`
	URL          string `json:"url,omitempty"`
	NeedApproval bool   `json:"need_approval,omitempty"`
	Completed    bool   `json:"completed,omitempty"`
}
