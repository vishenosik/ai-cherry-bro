package browser

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/playwright-community/playwright-go"
)

type BrowserAgent struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	context playwright.BrowserContext

	isRunning atomic.Bool
}

func NewBrowserAgent() (*BrowserAgent, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %v", err)
	}

	ba := &BrowserAgent{
		pw: pw,
	}

	ba.isRunning.Store(true)

	return ba, nil

}

func (ba *BrowserAgent) Start(ctx context.Context) error {

	// Запускаем видимый браузер
	browser, err := ba.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("could not launch browser: %v", err)
	}

	// Persistent context для сохранения сессий
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		NoViewport: playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("could not create context: %v", err)
	}

	ba.browser = browser
	ba.context = context

	return nil
}

func (ba *BrowserAgent) Stop(ctx context.Context) error {
	ba.isRunning.Store(false)

	if err := ba.browser.Close(); err != nil {
		return fmt.Errorf("could not close browser: %v", err)
	}
	if err := ba.context.Close(); err != nil {
		return fmt.Errorf("could not close context: %v", err)
	}
	return nil
}

func (ba *BrowserAgent) Test(ctx context.Context) error {
	page, err := ba.context.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %v", err)
	}

	if _, err = page.Goto("https://news.ycombinator.com"); err != nil {
		return fmt.Errorf("could not goto: %v", err)
	}

	entries, err := page.Locator(".athing").All()
	if err != nil {
		return fmt.Errorf("could not get entries: %v", err)
	}

	for i, entry := range entries {
		title, err := entry.Locator("td.title > span > a").TextContent()
		if err != nil {
			return fmt.Errorf("could not get text content: %v", err)
		}
		fmt.Printf("%d: %s\n", i+1, title)
	}

	err = page.Locator(".pagetop > a", playwright.PageLocatorOptions{
		HasText: "jobs",
	}).Click()
	if err != nil {
		return fmt.Errorf("could not get entries: %v", err)
	}
	return nil
}
