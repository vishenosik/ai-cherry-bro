package core

import (
	"context"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/vishenosik/concurrency"
	"github.com/vishenosik/gocherry/pkg/logs"
)

func (o *Orchestrator) startPool(ctx context.Context) error {
	o.pool.Start(ctx)

	metrics := o.pool.GetMetrics()

	o.log.Info("pool started",
		slog.Int("workers_current", int(metrics.WorkersCurrent)),
		slog.Int("workers_max", int(metrics.WorkersMax)),
		slog.Int("workers_min", int(metrics.WorkersMin)),
	)

	go func() {
		for task := range o.subChan {

			_, err := o.pool.AddTask(
				concurrency.Task{
					ID: task.ID,
					Func: func() {
						o.RunTask(task.Text)
					},
					Priority: concurrency.Priority(0),
				},
			)

			if err != nil {
				o.log.Error("pool error", logs.Error(err))
				if errors.Is(err, concurrency.ErrPoolClosed) {
					return
				}
			}
		}
		o.log.Warn("subs exited")
	}()
	return nil
}

func (o *Orchestrator) stopPool(ctx context.Context) error {
	o.pool.Stop(ctx)
	o.log.Info("pool stopped")
	return nil
}
