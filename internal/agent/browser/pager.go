package browser

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

type Pager struct {
	page playwright.Page
	log  *slog.Logger
}

func (p *Pager) Close() error {
	return p.page.Close()
}

func (p *Pager) Navigate(url string) error {
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}

	p.log.Info("Navigating to " + url)
	_, err := p.page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(30000),
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	return err
}

func (p *Pager) ClickElement(description string) error {
	p.log.Info("🖱️ Attempting to click: " + description)

	// Пытаемся найти элемент различными стратегиями
	element, err := p.findElementByMultipleStrategies(description)
	if err != nil {
		return fmt.Errorf("element not found: %v", err)
	}

	// Проверяем видимость
	visible, err := element.IsVisible()
	if err != nil || !visible {
		return fmt.Errorf("element not visible")
	}

	// Проверяем, что элемент кликабелен
	enabled, err := element.IsEnabled()
	if err != nil || !enabled {
		return fmt.Errorf("element not enabled")
	}

	// Кликаем
	if err := element.Click(); err != nil {
		return fmt.Errorf("click failed: %v", err)
	}

	p.log.Info("✅ Successfully clicked: " + description)
	return nil
}

// findElementByMultipleStrategies ищет элемент используя multiple стратегии
func (p *Pager) findElementByMultipleStrategies(description string) (playwright.ElementHandle, error) {
	strategies := []struct {
		name     string
		findFunc func(string) (playwright.ElementHandle, error)
	}{
		{"exact text", p.findElementByExactText},
		{"partial text", p.findElementByPartialText},
		{"placeholder", p.findElementByPlaceholder},
		{"aria-label", p.findElementByAriaLabel},
		{"button by type", p.findElementByButtonType},
		{"link by href", p.findElementByLinkHref},
		{"input by type", p.findElementByInputType},
		{"css class", p.findElementByCSSClass},
		{"data attributes", p.findElementByDataAttributes},
		{"role attribute", p.findElementByRole},
		{"form elements", p.findElementByFormAttributes},
		{"generic clickable", p.findGenericClickableElement},
	}

	for _, strategy := range strategies {
		element, err := strategy.findFunc(description)
		if err == nil && element != nil {
			p.log.Debug(fmt.Sprintf("✅ Found element using strategy: %s", strategy.name))
			return element, nil
		}
		p.log.Debug(fmt.Sprintf("❌ Strategy failed: %s - %v", strategy.name, err))
	}

	return nil, fmt.Errorf("element '%s' not found with any strategy", description)
}

// findElementByExactText ищет по точному тексту
func (p *Pager) findElementByExactText(text string) (playwright.ElementHandle, error) {
	selector := fmt.Sprintf("text='%s'", text)
	return p.page.QuerySelector(selector)
}

// findElementByPartialText ищет по частичному совпадению текста
func (p *Pager) findElementByPartialText(text string) (playwright.ElementHandle, error) {
	selector := fmt.Sprintf("text=/.*%s.*/i", text)
	return p.page.QuerySelector(selector)
}

// findElementByPlaceholder ищет по placeholder атрибуту
func (p *Pager) findElementByPlaceholder(text string) (playwright.ElementHandle, error) {
	selectors := []string{
		fmt.Sprintf("[placeholder*='%s']", strings.ToLower(text)),
		fmt.Sprintf("[placeholder*='%s']", text),
	}

	for _, selector := range selectors {
		element, err := p.page.QuerySelector(selector)
		if err == nil && element != nil {
			return element, nil
		}
	}
	return nil, fmt.Errorf("no element with placeholder containing '%s'", text)
}

// findElementByAriaLabel ищет по aria-label атрибуту
func (p *Pager) findElementByAriaLabel(text string) (playwright.ElementHandle, error) {
	selectors := []string{
		fmt.Sprintf("[aria-label*='%s']", strings.ToLower(text)),
		fmt.Sprintf("[aria-label*='%s']", text),
		fmt.Sprintf("[aria-labelledby*='%s']", strings.ToLower(text)),
	}

	for _, selector := range selectors {
		element, err := p.page.QuerySelector(selector)
		if err == nil && element != nil {
			return element, nil
		}
	}
	return nil, fmt.Errorf("no element with aria-label containing '%s'", text)
}

// findElementByButtonType ищет кнопки по типам и тексту
func (p *Pager) findElementByButtonType(text string) (playwright.ElementHandle, error) {
	buttonSelectors := []string{
		"button",
		"input[type='submit']",
		"input[type='button']",
		"input[type='reset']",
		"[role='button']",
	}

	// Сначала ищем кнопки с нужным текстом
	for _, baseSelector := range buttonSelectors {
		selector := fmt.Sprintf("%s:has-text('%s')", baseSelector, text)
		element, err := p.page.QuerySelector(selector)
		if err == nil && element != nil {
			return element, nil
		}
	}

	// Затем ищем любые кнопки
	for _, selector := range buttonSelectors {
		element, err := p.page.QuerySelector(selector)
		if err == nil && element != nil {
			// Проверяем видимый текст кнопки
			buttonText, err := element.TextContent()
			if err == nil && strings.Contains(strings.ToLower(buttonText), strings.ToLower(text)) {
				return element, nil
			}
		}
	}

	return nil, fmt.Errorf("no button found for '%s'", text)
}

// findElementByLinkHref ищет ссылки по href и тексту
func (p *Pager) findElementByLinkHref(text string) (playwright.ElementHandle, error) {
	// Ищем ссылки с текстом
	selector := fmt.Sprintf("a:has-text('%s')", text)
	element, err := p.page.QuerySelector(selector)
	if err == nil && element != nil {
		return element, nil
	}

	// Ищем ссылки с href содержащим текст
	selector = fmt.Sprintf("a[href*='%s']", strings.ToLower(text))
	return p.page.QuerySelector(selector)
}

// findElementByInputType ищет input элементы по типу
func (p *Pager) findElementByInputType(description string) (playwright.ElementHandle, error) {
	inputTypes := map[string][]string{
		"search":   {"search", "find", "query"},
		"email":    {"email", "mail"},
		"password": {"password", "pass", "pwd"},
		"text":     {"text", "input", "field", "enter"},
		"username": {"username", "login", "user"},
	}

	for inputType, keywords := range inputTypes {
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(description), keyword) {
				selectors := []string{
					fmt.Sprintf("input[type='%s']", inputType),
					fmt.Sprintf("input[placeholder*='%s']", keyword),
				}

				for _, selector := range selectors {
					element, err := p.page.QuerySelector(selector)
					if err == nil && element != nil {
						return element, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("no input element found for '%s'", description)
}

// findElementByCSSClass ищет по CSS классам
func (p *Pager) findElementByCSSClass(description string) (playwright.ElementHandle, error) {
	commonClasses := map[string][]string{
		"button":   {"btn", "button", "submit", "cta", "action"},
		"search":   {"search", "find", "query"},
		"login":    {"login", "signin", "auth"},
		"menu":     {"menu", "nav", "navigation"},
		"close":    {"close", "exit", "cancel"},
		"next":     {"next", "continue", "forward"},
		"previous": {"prev", "previous", "back"},
	}

	descriptionLower := strings.ToLower(description)

	for _, classKeywords := range commonClasses {
		for _, keyword := range classKeywords {
			if strings.Contains(descriptionLower, keyword) {
				// Ищем элементы с классами содержащими ключевые слова
				selectors := []string{
					fmt.Sprintf("[class*='%s']", keyword),
					fmt.Sprintf(".%s", keyword),
				}

				for _, selector := range selectors {
					elements, err := p.page.QuerySelectorAll(selector)
					if err == nil && len(elements) > 0 {
						// Возвращаем первый видимый элемент
						for _, element := range elements {
							if visible, _ := element.IsVisible(); visible {
								return element, nil
							}
						}
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("no element found by CSS classes for '%s'", description)
}

// findElementByDataAttributes ищет по data-атрибутам
func (p *Pager) findElementByDataAttributes(description string) (playwright.ElementHandle, error) {
	dataAttributes := []string{
		"data-testid", "data-qa", "data-test", "data-id",
		"data-action", "data-target", "data-role",
	}

	descriptionLower := strings.ToLower(strings.ReplaceAll(description, " ", "-"))

	for _, attr := range dataAttributes {
		selectors := []string{
			fmt.Sprintf("[%s*='%s']", attr, descriptionLower),
			fmt.Sprintf("[%s*='%s']", attr, strings.ToLower(description)),
		}

		for _, selector := range selectors {
			element, err := p.page.QuerySelector(selector)
			if err == nil && element != nil {
				return element, nil
			}
		}
	}

	return nil, fmt.Errorf("no element found by data attributes for '%s'", description)
}

// findElementByRole ищет по ARIA role атрибуту
func (p *Pager) findElementByRole(description string) (playwright.ElementHandle, error) {
	roleMapping := map[string][]string{
		"button":  {"button", "submit", "link"},
		"search":  {"search", "searchbox"},
		"menu":    {"menu", "navigation"},
		"link":    {"link"},
		"textbox": {"textbox"},
	}

	descriptionLower := strings.ToLower(description)

	for role, keywords := range roleMapping {
		for _, keyword := range keywords {
			if strings.Contains(descriptionLower, keyword) {
				selector := fmt.Sprintf("[role='%s']", role)
				element, err := p.page.QuerySelector(selector)
				if err == nil && element != nil {
					return element, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no element found by role for '%s'", description)
}

// findElementByFormAttributes ищет form элементы
func (p *Pager) findElementByFormAttributes(description string) (playwright.ElementHandle, error) {
	formSelectors := []string{
		"form",
		"button[type='submit']",
		"input[type='submit']",
		"button[form]",
	}

	descriptionLower := strings.ToLower(description)

	if strings.Contains(descriptionLower, "form") ||
		strings.Contains(descriptionLower, "submit") ||
		strings.Contains(descriptionLower, "send") {

		for _, selector := range formSelectors {
			element, err := p.page.QuerySelector(selector)
			if err == nil && element != nil {
				return element, nil
			}
		}
	}

	return nil, fmt.Errorf("no form element found for '%s'", description)
}

// findGenericClickableElement ищет generic кликабельные элементы
func (p *Pager) findGenericClickableElement(description string) (playwright.ElementHandle, error) {
	// Ищем все потенциально кликабельные элементы
	clickableSelectors := []string{
		"button", "a", "input[type='button']", "input[type='submit']",
		"[onclick]", "[role='button']", "[tabindex]:not([tabindex='-1'])",
	}

	var allElements []playwright.ElementHandle

	for _, selector := range clickableSelectors {
		elements, err := p.page.QuerySelectorAll(selector)
		if err == nil {
			allElements = append(allElements, elements...)
		}
	}

	// Фильтруем по текстовому содержанию
	for _, element := range allElements {
		text, err := element.TextContent()
		if err != nil {
			continue
		}

		if strings.Contains(strings.ToLower(strings.TrimSpace(text)), strings.ToLower(description)) {
			return element, nil
		}

		// Также проверяем другие атрибуты
		attrs := []string{"title", "aria-label", "placeholder", "value"}
		for _, attr := range attrs {
			attrValue, _ := element.GetAttribute(attr)
			if strings.Contains(strings.ToLower(attrValue), strings.ToLower(description)) {
				return element, nil
			}
		}
	}

	return nil, fmt.Errorf("no generic clickable element found for '%s'", description)
}

func (p *Pager) TypeText(description, text string) error {
	p.log.Info(fmt.Sprintf("⌨️ Typing in %s: %s", description, text))

	element, err := p.findElementByText(description)
	if err != nil {
		// Пробуем найти input field
		element, err = p.page.QuerySelector("input, textarea")
		if err != nil {
			return fmt.Errorf("no input field found: %v", err)
		}
	}

	if err := element.Fill(text); err != nil {
		return fmt.Errorf("typing failed: %v", err)
	}

	p.log.Info("Successfully typed in " + description)
	return nil
}

func (p *Pager) ScrollPage() error {
	p.log.Info("Scrolling page")
	_, err := p.page.Evaluate("window.scrollBy(0, 500)")
	return err
}

func (p *Pager) Wait(seconds int) {
	p.log.Info(fmt.Sprintf("Waiting %d seconds", seconds))
	time.Sleep(time.Duration(seconds) * time.Second)
}

func (p *Pager) CurrentURL() string {
	url := p.page.URL()
	return url
}

func (p *Pager) TakeScreenshot(filename string) error {
	_, err := p.page.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String(filename),
		FullPage: playwright.Bool(true),
	})
	return err
}

func (p *Pager) findElementByText(text string) (playwright.ElementHandle, error) {
	// Ищем по точному тексту
	selector := fmt.Sprintf("text=%s", text)
	element, err := p.page.QuerySelector(selector)
	if err == nil && element != nil {
		return element, nil
	}

	// Ищем по частичному совпадению
	selector = fmt.Sprintf("text=/.*%s.*/i", text)
	element, err = p.page.QuerySelector(selector)
	if err == nil && element != nil {
		return element, nil
	}

	// Ищем по атрибутам
	selectors := []string{
		fmt.Sprintf("[placeholder*='%s']", strings.ToLower(text)),
		fmt.Sprintf("[value*='%s']", strings.ToLower(text)),
		fmt.Sprintf("[aria-label*='%s']", strings.ToLower(text)),
	}

	for _, sel := range selectors {
		element, err := p.page.QuerySelector(sel)
		if err == nil && element != nil {
			return element, nil
		}
	}

	return nil, fmt.Errorf("element with text '%s' not found", text)
}
