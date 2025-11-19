package core

import (
	"fmt"
	"log"
	"log/slog"
	"strings"
	"time"

	"context"

	"github.com/vishenosik/ai-cherry-bro/internal/agent/ai"
	_ctx "github.com/vishenosik/ai-cherry-bro/internal/context"
	"github.com/vishenosik/ai-cherry-bro/internal/entity"
	"github.com/vishenosik/ai-cherry-bro/internal/security"
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

type Orchestrator struct {
	browser        Browser
	page           Page
	aiClient       *ai.Client
	contextManager *_ctx.Manager
	securityLayer  *security.Layer
	isRunning      bool
	currentTask    string
	maxSteps       int

	log     *slog.Logger
	pool    *concurrency.Pool
	subChan <-chan entity.PoolTask
}

func NewOrchestrator(
	browser Browser,
	aiClient *ai.Client,
	contextManager *_ctx.Manager,
	securityLayer *security.Layer,
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

	fmt.Printf("🎯 Starting task: %s\n", task)
	fmt.Printf("📝 Maximum steps: %d\n", o.maxSteps)

	for step := 1; step <= o.maxSteps && o.isRunning; step++ {
		fmt.Printf("\n--- Step %d ---\n", step)

		// Получаем текущее состояние страницы
		pageState, err := o.page.ExtractPageState()
		if err != nil {
			log.Printf("❌ Failed to extract page state: %v", err)
			break
		}

		// Получаем историю действий
		history := o.contextManager.GetHistory()

		// Решаем следующее действие
		action, err := o.decideNextAction(task, pageState, history)
		if err != nil {
			log.Printf("❌ Failed to decide action: %v", err)
			break
		}

		fmt.Printf("🤔 Reasoning: %s\n", action.Reasoning)
		fmt.Printf("⚡ Action: %s", action.Action)
		if action.Target != "" {
			fmt.Printf(" -> %s", action.Target)
		}
		if action.Text != "" {
			fmt.Printf(" (text: %s)", action.Text)
		}
		fmt.Println()

		// Проверка безопасности для чувствительных действий
		if !o.securityLayer.CheckAction(action.Action, action.Target, action.Reasoning) {
			fmt.Println("❌ Action cancelled by user")
			break
		}

		// Выполняем действие
		if err := o.executeAction(action); err != nil {
			log.Printf("❌ Action failed: %v", err)

			// Пробуем восстановиться
			if !o.handleError(err, action) {
				break
			}
		}

		// Добавляем в историю
		o.contextManager.AddToHistory(fmt.Sprintf("%s: %s -> %s", action.Action, action.Target, action.Reasoning))

		// Проверяем завершение
		if action.Completed {
			fmt.Println("✅ Task completed successfully!")
			break
		}

		// Пауза между действиями
		time.Sleep(2 * time.Second)

		if step == o.maxSteps {
			fmt.Println("⚠️ Maximum steps reached. Task may not be complete.")
		}
	}

	o.isRunning = false
}

func (o *Orchestrator) decideNextAction(task, pageState, history string) (*entity.AiResponse, error) {
	messages := ai.BuildDecisionPrompt(task, pageState, history)
	return o.aiClient.CallAI(messages)
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

func (o *Orchestrator) handleError(err error, failedAction *entity.AiResponse) bool {
	errorMsg := err.Error()
	fmt.Printf("🔄 Handling error: %s\n", errorMsg)

	// Стратегии восстановления
	switch {
	case strings.Contains(errorMsg, "element not found"):
		fmt.Println("🔍 Element not found, trying to scroll...")
		o.page.ScrollPage()
		return true
	case strings.Contains(errorMsg, "not visible"):
		fmt.Println("👀 Element not visible, scrolling to view...")
		o.page.ScrollPage()
		return true
	case strings.Contains(errorMsg, "navigation"):
		fmt.Println("🌐 Navigation issue, waiting...")
		o.page.Wait(5)
		return true
	default:
		fmt.Println("❌ Unrecoverable error")
		return false
	}
}
