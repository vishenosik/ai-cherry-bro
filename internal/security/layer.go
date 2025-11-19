package security

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Layer struct {
	sensitiveActions []string
}

func NewLayer() *Layer {
	return &Layer{
		sensitiveActions: []string{
			"buy", "purchase", "pay", "order", "checkout",
			"delete", "remove", "cancel", "unsubscribe",
			"confirm", "submit", "send", "post", "publish",
			"transfer", "withdraw", "install", "download",
		},
	}
}

func (s *Layer) CheckAction(action, target, reasoning string) bool {
	actionLower := strings.ToLower(action + " " + target + " " + reasoning)

	for _, sensitive := range s.sensitiveActions {
		if strings.Contains(actionLower, sensitive) {
			fmt.Printf("\n🚨 SECURITY ALERT 🚨\n")
			fmt.Printf("Action: %s %s\n", action, target)
			fmt.Printf("Reasoning: %s\n", reasoning)
			fmt.Printf("This appears to be a sensitive action.\n")
			fmt.Print("Do you want to proceed? (y/n): ")

			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				response := strings.TrimSpace(scanner.Text())
				return strings.ToLower(response) == "y"
			}
		}
	}

	return true
}
