package worker

import (
	"context"
	"time"

	"CONVERDA/global"
	"CONVERDA/internal/messaging/domain/model/entity"
	domainRepo "CONVERDA/internal/messaging/domain/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	defaultSLAThreshold = 900 // 15 minutes in seconds
	batchSize           = 500
)

type SLAWorker struct {
	threadRepo domainRepo.ThreadRepository
	logRepo    domainRepo.AssignmentLogRepository
	appReader  domainRepo.AppReader
	interval   time.Duration
}

func NewSLAWorker(
	threadRepo domainRepo.ThreadRepository,
	logRepo domainRepo.AssignmentLogRepository,
	appReader domainRepo.AppReader,
) *SLAWorker {
	return &SLAWorker{
		threadRepo: threadRepo,
		logRepo:    logRepo,
		appReader:  appReader,
		interval:   1 * time.Minute,
	}
}

func (w *SLAWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	global.Logger.Info("SLA Worker started")

	for {
		select {
		case <-ctx.Done():
			global.Logger.Info("SLA Worker stopping")
			return
		case <-ticker.C:
			w.checkSLA(ctx)
		}
	}
}

func (w *SLAWorker) checkSLA(ctx context.Context) {
	w.internalCheckSLA(ctx)
}

func (w *SLAWorker) Exported_checkSLA(ctx context.Context) {
	w.internalCheckSLA(ctx)
}

func (w *SLAWorker) internalCheckSLA(ctx context.Context) {
	overdueFilter := false
	statuses := []string{"unassigned", "assigned"}

	for _, status := range statuses {
		offset := 0
		for {
			threads, _, err := w.threadRepo.List(ctx, domainRepo.ThreadFilter{
				Status:    &status,
				IsOverdue: &overdueFilter,
				Limit:     batchSize,
				Offset:    offset,
			})
			if err != nil {
				global.Logger.Error("SLA Worker: failed to list threads", zap.String("status", status), zap.Error(err))
				break
			}

			for _, t := range threads {
				if w.isThreadOverdue(ctx, t) {
					if err := w.threadRepo.MarkAsOverdue(ctx, t.ID); err != nil {
						global.Logger.Error("SLA Worker: failed to mark thread as overdue", zap.String("threadID", t.ID.String()), zap.Error(err))
					} else {
						global.Logger.Info("SLA Worker: thread marked as overdue", zap.String("threadID", t.ID.String()))
					}
				}
			}

			// Break if we got fewer than batchSize (no more pages)
			if len(threads) < batchSize {
				break
			}
			offset += batchSize
		}
	}
}

func (w *SLAWorker) isThreadOverdue(ctx context.Context, t *entity.Thread) bool {
	threshold := w.getSLAThreshold(ctx, t.EnvironmentID)

	var startTime time.Time
	if t.Status == entity.ThreadStatusUnassigned {
		startTime = t.CreatedAt
	} else if t.Status == entity.ThreadStatusAssigned {
		// Find last assignment time from logs to calculate response time
		log, err := w.logRepo.GetLastByThread(ctx, t.ID)
		if err == nil && log != nil {
			startTime = log.AssignedAt
		} else {
			startTime = t.UpdatedAt // Fallback to last update if log missing
		}
	} else {
		return false
	}

	return time.Since(startTime).Seconds() > float64(threshold)
}

func (w *SLAWorker) getSLAThreshold(ctx context.Context, envID uuid.UUID) int {
	if envID == uuid.Nil {
		return defaultSLAThreshold
	}

	env, err := w.appReader.GetEnvironment(ctx, envID)
	if err == nil && env != nil {
		if env.SLAThresholdSeconds > 0 {
			return env.SLAThresholdSeconds
		}
		app, err := w.appReader.GetApp(ctx, env.AppID)
		if err == nil && app != nil && app.SLAThresholdSeconds > 0 {
			return app.SLAThresholdSeconds
		}
	}

	return defaultSLAThreshold
}
