package core

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"context"

	"github.com/vishenosik/ai-cherry-bro/internal/agent/ai"
	_ctx "github.com/vishenosik/ai-cherry-bro/internal/context"
	"github.com/vishenosik/ai-cherry-bro/internal/entity"
	"github.com/vishenosik/concurrency"
	"github.com/vishenosik/gocherry/pkg/logs"
)

type Browser interface {
	NewPage() (Page, error)
}

type Page interface {
	ExtractPageState() (string, error)
	ScrollPage() error
	Wait(seconds int)
	ClickElement(description string) error
	TypeText(description string, text string) error
	Navigate(url string) error
	Close() error
}

type AiClient interface {
	Call(messages []entity.AiMessage) (*entity.AiResponse, error)
}

type ContextManager interface {
	AddToHistory(action string)
	CheckAuthRequired(task string, currentURL string) bool
	ClearHistory()
	GetAuthState(domain string) *_ctx.AuthState
	GetHistory() string
	UpdateAuthState(url string, isLoggedIn bool, username string)
}

type Security interface {
	CheckAction(action string, target string, reasoning string) bool
}

type Orchestrator struct {
	browser        Browser
	page           Page
	aiClient       AiClient
	contextManager ContextManager
	securityLayer  Security
	isRunning      bool
	currentTask    string
	maxSteps       int

	log     *slog.Logger
	pool    *concurrency.Pool
	subChan <-chan entity.PoolTask
}

func NewOrchestrator(
	browser Browser,
	aiClient AiClient,
	contextManager ContextManager,
	securityLayer Security,
	subscriptions ...chan entity.PoolTask,
) (*Orchestrator, error) {

	return &Orchestrator{
		browser:        browser,
		aiClient:       aiClient,
		contextManager: contextManager,
		securityLayer:  securityLayer,
		maxSteps:       50,

		log: logs.SetupLogger().With(logs.AppComponent("core_orchestrator")),

		pool:    concurrency.NewWorkerPool(concurrency.WithWorkersControl(1, 1, 1)),
		subChan: concurrency.MergeChannels(context.Background(), uint16(1024), subscriptions...),
	}, nil
}

func (o *Orchestrator) Start(ctx context.Context) error {
	page, err := o.browser.NewPage()
	if err != nil {
		return err
	}

	o.page = page

	if err := o.startPool(ctx); err != nil {
		return err
	}
	return nil
}

func (o *Orchestrator) Stop(ctx context.Context) error {
	err := o.stopPool(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (o *Orchestrator) RunTask(task string) {
	o.currentTask = task
	o.isRunning = true

	o.log.Info("starting task",
		slog.String("task", task),
		slog.Int("max_steps", o.maxSteps),
	)

	for step := 1; step <= o.maxSteps && o.isRunning; step++ {

		// Получаем текущее состояние страницы
		pageState, err := o.page.ExtractPageState()
		if err != nil {
			o.log.Error("Failed to extract page state", logs.Error(err))
			break
		}

		// Решаем следующее действие
		action, err := o.decideNextAction(task, pageState, o.contextManager.GetHistory())
		if err != nil {
			o.log.Error("failed to decide action", logs.Error(err))
			break
		}

		log := o.log.With(
			slog.String("action", action.Action),
			slog.String("target", action.Target),
		)

		log.Info("acting",
			slog.Int("step", step),
			slog.String("reasoning", action.Reasoning),
		)

		// Проверка безопасности для чувствительных действий
		if !o.securityLayer.CheckAction(action.Action, action.Target, action.Reasoning) {
			log.Error("action cancelled by user")
			break
		}

		// Выполняем действие
		if err := o.executeAction(action); err != nil {
			log.Error("action failed", logs.Error(err))

			// Пробуем восстановиться
			if !o.handleError(err, action) {
				break
			}
		}

		// Добавляем в историю
		o.contextManager.AddToHistory(fmt.Sprintf("%s: %s -> %s", action.Action, action.Target, action.Reasoning))

		// Проверяем завершение
		if action.Completed {
			o.log.Info("task completed successfully")
			break
		}

		// Пауза между действиями
		time.Sleep(2 * time.Second)

		if step == o.maxSteps {
			o.log.Warn("maximum steps reached. task may not be complete")
		}
	}

	o.isRunning = false
}

func (o *Orchestrator) decideNextAction(task, pageState, history string) (*entity.AiResponse, error) {
	messages := ai.BuildDecisionPrompt(task, pageState, history)
	return o.aiClient.Call(messages)
}

func (o *Orchestrator) executeAction(action *entity.AiResponse) error {
	switch action.Action {
	case "click":
		return o.page.ClickElement(action.Target)
	case "type":
		return o.page.TypeText(action.Target, action.Text)
	case "navigate":
		return o.page.Navigate(action.URL)
	case "scroll":
		return o.page.ScrollPage()
	case "wait":
		o.page.Wait(3)
		return nil
	case "complete":
		return nil
	default:
		return fmt.Errorf("unknown action: %s", action.Action)
	}
}

func (o *Orchestrator) handleError(
	err error,
	_ *entity.AiResponse,
) bool {
	errorMsg := err.Error()
	o.log.Warn("handling error", logs.Error(err))

	// Стратегии восстановления
	switch {
	case strings.Contains(errorMsg, "element not found"):
		o.log.Error("Element not found, trying to scroll...")
		o.page.ScrollPage()
		return true

	case strings.Contains(errorMsg, "not visible"):
		o.log.Error("Element not visible, scrolling to view...")
		o.page.ScrollPage()
		return true

	case strings.Contains(errorMsg, "navigation"):
		o.log.Error("Navigation issue, waiting...")
		o.page.Wait(5)
		return true

	default:
		o.log.Error("❌ Unrecoverable error")
		return false
	}
}
