package ai

import (
	"fmt"

	"github.com/vishenosik/ai-cherry-bro/internal/entity"
)

func BuildDecisionPrompt(task, pageState, history string) []entity.AiMessage {
	systemPrompt := `You are an autonomous web browsing AI agent. Your goal is to complete tasks by interacting with web pages.

AVAILABLE ACTIONS:
- click: Click on an element (button, link, etc.)
- type: Type text into an input field  
- navigate: Go to a new URL
- scroll: Scroll the page to see more content
- wait: Wait for page to load
- complete: Task is finished

RESPONSE FORMAT:
{
    "reasoning": "Your step-by-step reasoning",
    "action": "action_name",
    "target": "element description or text",
    "text": "text to type (if applicable)",
    "url": "url to navigate to (if applicable)",
    "need_approval": true/false
}

SECURITY: Set "need_approval": true for destructive actions like purchases, deletions, etc.

BE SPECIFIC: Describe exactly what element to interact with based on the visible text and context.`

	userPrompt := fmt.Sprintf(`TASK: %s

CURRENT PAGE STATE:
%s

RECENT HISTORY:
%s

Based on the current page and task, decide the next action. Be precise about what element to interact with.`,
		task, pageState, history)

	return []entity.AiMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
}
