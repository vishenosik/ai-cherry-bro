// context/manager.go
package context

import (
	"strings"
)

type AuthState struct {
	IsLoggedIn   bool
	Username     string
	Domain       string
	AuthRequired bool
}

type Manager struct {
	maxTokens  int
	history    []string
	maxHistory int
	authStates map[string]*AuthState // domain -> auth state
}

func NewManager(maxTokens int) *Manager {
	return &Manager{
		maxTokens:  maxTokens,
		maxHistory: 15,
		authStates: make(map[string]*AuthState),
	}
}

func (m *Manager) UpdateAuthState(url string, isLoggedIn bool, username string) {
	domain := extractDomain(url)
	m.authStates[domain] = &AuthState{
		IsLoggedIn:   isLoggedIn,
		Username:     username,
		Domain:       domain,
		AuthRequired: false,
	}
}

func (m *Manager) CheckAuthRequired(task string, currentURL string) bool {
	// Анализируем задачу на необходимость авторизации
	authKeywords := []string{
		"мой", "мои", "моё", "my", "личн", "профиль", "profile",
		"заказы", "orders", "покупки", "purchases", "история", "history",
		"сообщения", "messages", "настройки", "settings", "аккаунт", "account",
	}

	taskLower := strings.ToLower(task)
	for _, keyword := range authKeywords {
		if strings.Contains(taskLower, keyword) {
			return true
		}
	}
	return false
}

func (m *Manager) GetAuthState(domain string) *AuthState {
	if state, exists := m.authStates[domain]; exists {
		return state
	}
	return &AuthState{
		IsLoggedIn:   false,
		AuthRequired: false,
	}
}

func extractDomain(url string) string {
	// Упрощенное извлечение домена
	if strings.Contains(url, "://") {
		parts := strings.Split(url, "/")
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	return url
}

func (m *Manager) AddToHistory(action string) {
	m.history = append(m.history, action)
	if len(m.history) > m.maxHistory {
		m.history = m.history[1:]
	}
}

func (m *Manager) GetHistory() string {
	if len(m.history) == 0 {
		return "No recent actions"
	}
	return strings.Join(m.history, "\n")
}

func (m *Manager) ClearHistory() {
	m.history = nil
}
