package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vishenosik/ai-cherry-bro/internal/agent/ai"
	"github.com/vishenosik/ai-cherry-bro/internal/agent/browser"
	"github.com/vishenosik/ai-cherry-bro/internal/agent/core"
	"github.com/vishenosik/ai-cherry-bro/internal/api"
	_context "github.com/vishenosik/ai-cherry-bro/internal/context"
	"github.com/vishenosik/ai-cherry-bro/internal/security"
	"github.com/vishenosik/ai-cherry-bro/internal/usecase"
	"github.com/vishenosik/gocherry"

	_ctx "github.com/vishenosik/gocherry/pkg/context"
	"github.com/vishenosik/gocherry/pkg/grpc"
	"github.com/vishenosik/gocherry/pkg/logs"
)

func main() {

	log := logs.SetupLogger().With(logs.AppComponent("main"))

	gocherry.Flags(os.Stdout, os.Args[1:],
		gocherry.AppFlags(os.Stdout),
		gocherry.ConfigFlags(os.Stdout),
	)

	flag.Parse()

	ctx := context.Background()

	app, err := NewApp(ctx)
	if err != nil {
		log.Error("failed to init app", logs.Error(err))
		os.Exit(1)
	}

	err = app.Start(ctx)
	if err != nil {
		log.Error("failed to start app", logs.Error(err))
		os.Exit(1)
	}

	// Graceful shut down
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	stopctx, stopCancel := context.WithTimeout(_ctx.WithStopCtx(context.Background(), <-stop), time.Second*5)
	defer stopCancel()

	app.Stop(stopctx)
}

func NewApp(ctx context.Context) (*gocherry.App, error) {

	// STORES

	// localStore := local.NewFileStore()

	// USECASES

	taskProvider := usecase.NewTaskProvider(nil)

	// API

	bsApi := api.NewBrowserServiceApi(taskProvider)

	// AGENTS

	browserAgent, err := browser.NewBrowserAgent()
	if err != nil {
		return nil, err
	}

	// SERVICES

	grpcServer, err := grpc.NewGrpcServer(
		grpc.GrpcServices{
			bsApi,
		},
		grpc.WithLogInterceptors(),
	)
	if err != nil {
		return nil, err
	}

	// TODO

	aiClient := ai.NewClient()
	contextManager := _context.NewManager(8000) // 8K токенов контекста
	securityLayer := security.NewLayer()

	// CORE

	orch, err := core.NewOrchestrator(
		browserAgent,
		aiClient,
		contextManager,
		securityLayer,
		taskProvider.TasksChan(),
	)
	if err != nil {
		return nil, err
	}

	// INIT

	app, err := gocherry.NewApp()
	if err != nil {
		return nil, err
	}

	app.AddServices(
		grpcServer,
		browserAgent,
		orch,
	)

	return app, nil
}
