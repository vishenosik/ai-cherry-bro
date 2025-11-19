package browser

import (
	"fmt"
	"strings"
)

type ElementInfo struct {
	TagName string
	Text    string
	ID      string
	Classes string
	Visible bool
	Type    string
}

func (p *Pager) ExtractPageState() (string, error) {
	var pageContent strings.Builder

	// Извлекаем основные элементы
	elements, err := p.extractInteractiveElements()
	if err != nil {
		return "", err
	}

	// Добавляем URL
	url := p.CurrentURL()
	pageContent.WriteString(fmt.Sprintf("Current URL: %s\n\n", url))

	// Добавляем заголовок страницы
	title, err := p.page.Title()
	if err == nil {
		pageContent.WriteString(fmt.Sprintf("Page Title: %s\n\n", title))
	}

	// Группируем элементы по типам
	pageContent.WriteString("=== INTERACTIVE ELEMENTS ===\n")

	// Кнопки
	pageContent.WriteString("\n--- BUTTONS ---\n")
	for _, el := range elements {
		if el.TagName == "button" || el.Type == "button" {
			pageContent.WriteString(fmt.Sprintf("- %s\n", el.Text))
		}
	}

	// Ссылки
	pageContent.WriteString("\n--- LINKS ---\n")
	for _, el := range elements {
		if el.TagName == "a" {
			pageContent.WriteString(fmt.Sprintf("- %s\n", el.Text))
		}
	}

	// Формы
	pageContent.WriteString("\n--- FORM ELEMENTS ---\n")
	for _, el := range elements {
		if el.TagName == "input" || el.TagName == "textarea" {
			pageContent.WriteString(fmt.Sprintf("- %s [%s]\n", el.Text, el.Type))
		}
	}

	// Заголовки
	headings, _ := p.extractHeadings()
	if len(headings) > 0 {
		pageContent.WriteString("\n--- HEADINGS ---\n")
		for _, h := range headings {
			pageContent.WriteString(fmt.Sprintf("- %s\n", h))
		}
	}

	return pageContent.String(), nil
}

func (p *Pager) extractInteractiveElements() ([]ElementInfo, error) {
	script := `
    () => {
        const elements = [];
        const selectors = [
            'a', 'button', 'input', 'textarea', 
            '[role="button"]', '[onclick]', '[type="submit"]'
        ];
        
        selectors.forEach(selector => {
            document.querySelectorAll(selector).forEach(el => {
                const rect = el.getBoundingClientRect();
                const isVisible = rect.width > 0 && rect.height > 0 && 
                                rect.top >= 0 && rect.left >= 0 &&
                                rect.bottom <= window.innerHeight && 
                                rect.right <= window.innerWidth;
                
                if (isVisible) {
                    const text = el.textContent?.trim() || 
                                el.getAttribute('placeholder') ||
                                el.getAttribute('value') ||
                                el.getAttribute('aria-label') ||
                                '';
                    
                    if (text && text.length < 100) { // Ограничиваем длину текста
                        elements.push({
                            tagName: el.tagName.toLowerCase(),
                            text: text,
                            id: el.id || '',
                            classes: el.className || '',
                            visible: isVisible,
                            type: el.type || ''
                        });
                    }
                }
            });
        });
        
        return elements;
    }
    `

	result, err := p.page.Evaluate(script)
	if err != nil {
		return nil, err
	}

	// Конвертируем результат
	var elements []ElementInfo
	if items, ok := result.([]interface{}); ok {
		for _, item := range items {
			if data, ok := item.(map[string]interface{}); ok {
				element := ElementInfo{
					TagName: getString(data, "tagName"),
					Text:    getString(data, "text"),
					ID:      getString(data, "id"),
					Classes: getString(data, "classes"),
					Visible: getBool(data, "visible"),
					Type:    getString(data, "type"),
				}
				if element.Text != "" {
					elements = append(elements, element)
				}
			}
		}
	}

	return elements, nil
}

func (p *Pager) extractHeadings() ([]string, error) {
	script := `
    () => {
        const headings = [];
        for (let i = 1; i <= 6; i++) {
            document.querySelectorAll('h' + i).forEach(h => {
                const text = h.textContent?.trim();
                if (text) headings.push('H' + i + ': ' + text);
            });
        }
        return headings;
    }
    `

	result, err := p.page.Evaluate(script)
	if err != nil {
		return nil, err
	}

	var headings []string
	if items, ok := result.([]interface{}); ok {
		for _, item := range items {
			if text, ok := item.(string); ok {
				headings = append(headings, text)
			}
		}
	}

	return headings, nil
}

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}
