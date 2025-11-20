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
- wait_user: wait for user interaction with browser.
- complete: Task is finished

RESPONSE FORMAT:
{
    "reasoning": "Your step-by-step reasoning",
    "action": "action_name",
    "target": "css selector, DOM element, useful tag or element description or text",
    "text": "text to type (if applicable)",
    "url": "url to navigate to (if applicable)",
    "need_approval": true/false
}

If you need to click a button or a link, use the "click" action and target exact button or link selector on the page. Preferably use interactive elements from page state.

SECURITY: Set "need_approval": true for destructive actions like purchases, deletions, etc.

If a task needs to be done with user login. Send action wait_user. Do not try to authenticate yourself.

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
