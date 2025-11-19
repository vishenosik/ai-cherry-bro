package browser

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

type Pager struct {
	page playwright.Page
}

func (p *Pager) Close() error {
	return p.page.Close()
}

func (p *Pager) Navigate(url string) error {
	if !strings.HasPrefix(url, "http") {
		url = "https://" + url
	}

	log.Printf("🌐 Navigating to: %s", url)
	_, err := p.page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(30000),
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	return err
}

func (p *Pager) ClickElement(description string) error {
	log.Printf("🖱️ Attempting to click: %s", description)

	// Ищем элемент по тексту
	element, err := p.findElementByText(description)
	if err != nil {
		return fmt.Errorf("element not found: %v", err)
	}

	// Проверяем видимость
	visible, err := element.IsVisible()
	if err != nil || !visible {
		return fmt.Errorf("element not visible")
	}

	// Кликаем
	if err := element.Click(); err != nil {
		return fmt.Errorf("click failed: %v", err)
	}

	log.Printf("✅ Successfully clicked: %s", description)
	return nil
}

func (p *Pager) TypeText(description, text string) error {
	log.Printf("⌨️ Typing in %s: %s", description, text)

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

	log.Printf("✅ Successfully typed in: %s", description)
	return nil
}

func (p *Pager) ScrollPage() error {
	log.Printf("📜 Scrolling page")
	_, err := p.page.Evaluate("window.scrollBy(0, 500)")
	return err
}

func (p *Pager) Wait(seconds int) {
	log.Printf("⏳ Waiting %d seconds", seconds)
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
