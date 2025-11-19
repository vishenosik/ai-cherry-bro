package usecase

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/vishenosik/ai-cherry-bro/internal/entity"
	"github.com/vishenosik/gocherry/pkg/logs"
)

type TaskProvider interface {
}

type provider struct {
	log    *slog.Logger
	source TaskProvider

	tasksCH chan entity.PoolTask
}

func NewTaskProvider(source TaskProvider) *provider {
	return &provider{
		source:  source,
		log:     logs.SetupLogger().With(logs.AppComponent("usecase.task_provider")),
		tasksCH: make(chan entity.PoolTask, 1024),
	}
}

func (fs *provider) NewTask(ctx context.Context, text string) (task_id string, err error) {
	task_id = uuid.New().String()

	fs.tasksCH <- entity.PoolTask{
		ID:   task_id,
		Text: text,
	}

	fs.log.Info("task created",
		slog.String("id", task_id),
		slog.String("text", text),
	)
	return task_id, nil
}

func (fs *provider) TasksChan() chan entity.PoolTask {
	return fs.tasksCH
}
