package impl

import (
	"context"
	"fmt"
	"time"

	domain "CONVERDA/internal/notification/domain"

	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"

	"github.com/google/uuid"
)

type jobScheduler struct {
	repo         repository.NotificationJobRepository
	uow          repository.NotificationUnitOfWork
	notifService service.NotificationService
	tmplManager  *service.TemplateManager
}

func NewJobScheduler(repo repository.NotificationJobRepository, uow repository.NotificationUnitOfWork, notifService service.NotificationService, tmplManager *service.TemplateManager) service.JobScheduler {
	return &jobScheduler{
		repo:         repo,
		uow:          uow,
		notifService: notifService,
		tmplManager:  tmplManager,
	}
}

func (s *jobScheduler) ScheduleJob(ctx context.Context, envID uuid.UUID, tenantID *uuid.UUID, channel string, templateID *uuid.UUID, scheduledAt *time.Time, recipients []map[string]interface{}, metadata map[string]interface{}) (*entity.NotificationJob, error) {
	status := entity.JobStatusPending
	if scheduledAt != nil && scheduledAt.After(time.Now()) {
		status = entity.JobStatusScheduled
	}

	recipientsData := map[string]interface{}{
		"recipients": recipients,
	}

	job := &entity.NotificationJob{
		ID:             uuid.New(),
		EnvironmentID:  envID,
		TenantID:       tenantID,
		Channel:        channel,
		TemplateID:     templateID,
		Status:         status,
		TotalCount:     len(recipients),
		ScheduledAt:    scheduledAt,
		Metadata:       metadata,
		RecipientsData: recipientsData,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.uow.Execute(ctx, func(tx repository.NotificationTxRepository) error {
		return tx.Jobs().Create(ctx, job)
	}); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *jobScheduler) CancelJob(ctx context.Context, envID uuid.UUID, id uuid.UUID) error {
	return s.uow.Execute(ctx, func(tx repository.NotificationTxRepository) error {
		job, err := tx.Jobs().GetByID(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to fetch job %s: %w", id, err)
		}
		if job == nil {
			return nil
		}
		if job.EnvironmentID != envID {
			return domain.ErrJobNotInEnv
		}

		if job.Status == entity.JobStatusCompleted || job.Status == entity.JobStatusFailed || job.Status == entity.JobStatusCancelled {
			return domain.ErrJobNotCancellable
		}

		job.Status = entity.JobStatusCancelled
		job.UpdatedAt = time.Now()
		return tx.Jobs().Update(ctx, job)
	})
}

func (s *jobScheduler) GetJob(ctx context.Context, envID uuid.UUID, id uuid.UUID) (*entity.NotificationJob, error) {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job %s: %w", id, err)
	}
	if job != nil && job.EnvironmentID != envID {
		return nil, nil
	}
	return job, nil
}

func (s *jobScheduler) ListJobs(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationJob, int64, error) {
	jobs, total, err := s.repo.List(ctx, envID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list notification jobs: %w", err)
	}
	return jobs, total, nil
}

func (s *jobScheduler) ProcessJob(ctx context.Context, jobID uuid.UUID) error {
	// This should ideally run in background or handle context timeout
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to fetch job %s: %w", jobID, err)
	}
	if job == nil {
		return domain.ErrJobNotFound
	}

	if job.Status != entity.JobStatusPending && job.Status != entity.JobStatusScheduled {
		return domain.ErrJobNotPending
	}

	// Fetch Template Code
	var templateCode string
	if job.TemplateID != nil {
		tmpl, err := s.tmplManager.GetTemplate(ctx, job.EnvironmentID, *job.TemplateID)
		if err != nil {
			return fmt.Errorf("failed to fetch template: %w", err)
		}
		if tmpl != nil {
			templateCode = tmpl.Code
		}
	} else {
		// Try metadata
		if code, ok := job.Metadata["template_code"].(string); ok {
			templateCode = code
		}
	}

	if templateCode == "" {
		return domain.ErrTemplateCodeMissing
	}

	// Step 1: Mark job as Processing
	if err := s.uow.Execute(ctx, func(tx repository.NotificationTxRepository) error {
		now := time.Now()
		job.StartedAt = &now
		job.Status = entity.JobStatusProcessing
		return tx.Jobs().Update(ctx, job)
	}); err != nil {
		return fmt.Errorf("failed to update job status to processing: %w", err)
	}

	recipientsList, ok := job.RecipientsData["recipients"].([]interface{})
	if !ok {
		// Log error or fail
		s.uow.Execute(ctx, func(tx repository.NotificationTxRepository) error {
			job.Status = entity.JobStatusFailed
			msg := "invalid recipients data"
			job.ErrorMessage = &msg
			return tx.Jobs().Update(ctx, job)
		})
		return fmt.Errorf("invalid recipients data")
	}

	success := 0
	failed := 0

	for _, r := range recipientsList {
		recipientMap, ok := r.(map[string]interface{})
		if !ok {
			failed++
			continue
		}

		recipientAddr, _ := recipientMap["recipient"].(string)

		var data map[string]interface{}
		if d, ok := recipientMap["data"].(map[string]interface{}); ok {
			data = d
		}
		lang, _ := recipientMap["language"].(string)

		req := service.SendRequest{
			TenantID:      safeUUIDString(job.TenantID),
			EnvironmentID: job.EnvironmentID.String(),
			TemplateCode:  templateCode,
			Recipient:     recipientAddr,
			Channel:       job.Channel,
			Data:          data,
			Language:      lang,
		}

		_, err := s.notifService.Send(ctx, req)
		if err != nil {
			failed++
		} else {
			success++
		}
	}

	// Step 2: Mark job as Completed or Failed
	return s.uow.Execute(ctx, func(tx repository.NotificationTxRepository) error {
		completedAt := time.Now()
		job.CompletedAt = &completedAt
		job.TotalCount = len(recipientsList)
		job.SuccessCount = success
		job.FailedCount = failed
		job.Status = entity.JobStatusCompleted

		if failed > 0 && success == 0 {
			job.Status = entity.JobStatusFailed
			msg := "all recipients failed"
			job.ErrorMessage = &msg
		} else if failed > 0 {
			msg := fmt.Sprintf("completed with %d failures", failed)
			job.ErrorMessage = &msg
		}

		return tx.Jobs().Update(ctx, job)
	})
}

func safeUUIDString(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}
