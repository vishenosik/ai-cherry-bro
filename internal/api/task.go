package api

import (
	"context"
	"log/slog"

	browser_task_v1 "github.com/vishenosik/ai-cherry-bro/gen/grpc/v1/browser_task"
	"github.com/vishenosik/gocherry/pkg/logs"
	"google.golang.org/grpc"
)

type BrowserTaskUsecase interface {
	NewTask(ctx context.Context, text string) (task_id string, err error)
}

type BrowserServiceApi struct {
	browser_task_v1.UnimplementedBrowserTaskServiceServer
	svc BrowserTaskUsecase
	// log is a structured logger for the application.
	log *slog.Logger
}

func NewBrowserServiceApi(svc BrowserTaskUsecase) *BrowserServiceApi {
	return &BrowserServiceApi{
		svc: svc,
		log: logs.SetupLogger().With(logs.AppComponent("browser_task_api")),
	}
}

func (bsa *BrowserServiceApi) RegisterService(server *grpc.Server) {
	browser_task_v1.RegisterBrowserTaskServiceServer(server, bsa)
}

func (bsa *BrowserServiceApi) NewTask(ctx context.Context, req *browser_task_v1.NewTaskReq) (*browser_task_v1.NewTaskResp, error) {
	task_id, err := bsa.svc.NewTask(ctx, req.TaskText)
	if err != nil {
		return nil, err
	}
	return &browser_task_v1.NewTaskResp{
		TaskId: task_id,
	}, nil
}
