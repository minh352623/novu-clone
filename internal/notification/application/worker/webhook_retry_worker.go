package worker

import (
	"context"
	"time"

	"CONVERDA/global"
	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/domain/repository"
)

const (
	retryBatchSize    = 50
	retryWorkerTicker = 30 * time.Second
)

type WebhookRetryWorker struct {
	dispatcher service.WebhookDispatcher
	logRepo    repository.WebhookLogRepository
	interval   time.Duration
}

func NewWebhookRetryWorker(
	dispatcher service.WebhookDispatcher,
	logRepo repository.WebhookLogRepository,
) *WebhookRetryWorker {
	return &WebhookRetryWorker{
		dispatcher: dispatcher,
		logRepo:    logRepo,
		interval:   retryWorkerTicker,
	}
}

func (w *WebhookRetryWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	global.Logger.Info("Webhook Retry Worker started")

	for {
		select {
		case <-ctx.Done():
			global.Logger.Info("Webhook Retry Worker stopping")
			return
		case <-ticker.C:
			w.processRetries(ctx)
		}
	}
}

func (w *WebhookRetryWorker) processRetries(ctx context.Context) {
	logs, err := w.logRepo.GetPendingRetries(ctx, retryBatchSize)
	if err != nil {
		global.Logger.Error("Webhook Retry Worker: failed to fetch pending retries: " + err.Error())
		return
	}

	if len(logs) == 0 {
		return
	}

	global.Logger.Info("Webhook Retry Worker: processing " + string(rune('0'+len(logs))) + " pending retries")

	for _, log := range logs {
		w.dispatcher.RetryWebhook(ctx, log)
	}
}
