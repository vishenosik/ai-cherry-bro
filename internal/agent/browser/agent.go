package browser

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/playwright-community/playwright-go"
	"github.com/vishenosik/ai-cherry-bro/internal/agent/core"
	"github.com/vishenosik/gocherry/pkg/logs"
)

type BrowserAgent struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	context playwright.BrowserContext
	log     *slog.Logger

	isRunning atomic.Bool
}

func NewBrowserAgent() (*BrowserAgent, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %v", err)
	}

	ba := &BrowserAgent{
		pw:  pw,
		log: logs.SetupLogger().With(logs.AppComponent("browser")),
	}

	// Запускаем видимый браузер
	browser, err := ba.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("could not launch browser: %v", err)
	}

	// Persistent context для сохранения сессий
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		NoViewport: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create context: %v", err)
	}

	ba.browser = browser
	ba.context = context

	ba.isRunning.Store(true)

	return ba, nil

}

func (ba *BrowserAgent) Close(ctx context.Context) error {
	ba.isRunning.Store(false)

	if err := ba.browser.Close(); err != nil {
		return fmt.Errorf("could not close browser: %v", err)
	}
	if err := ba.context.Close(); err != nil {
		return fmt.Errorf("could not close context: %v", err)
	}
	return nil
}

func (ba *BrowserAgent) NewPage() (core.Page, error) {

	page, err := ba.context.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %v", err)
	}

	return &Pager{
		page: page,
		log:  ba.log,
	}, nil
}
