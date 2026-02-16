package impl

import (
	"context"
	"errors"
	"time"

	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"

	"github.com/google/uuid"
)

type jobScheduler struct {
	repo         repository.NotificationJobRepository
	notifService service.NotificationService
	tmplManager  *service.TemplateManager
}

func NewJobScheduler(repo repository.NotificationJobRepository, notifService service.NotificationService, tmplManager *service.TemplateManager) service.JobScheduler {
	return &jobScheduler{
		repo:         repo,
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

	if err := s.repo.Create(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *jobScheduler) CancelJob(ctx context.Context, envID uuid.UUID, id uuid.UUID) error {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}
	if job.EnvironmentID != envID {
		return errors.New("job not found in this environment")
	}

	if job.Status == entity.JobStatusCompleted || job.Status == entity.JobStatusFailed || job.Status == entity.JobStatusCancelled {
		return errors.New("cannot cancel finished job")
	}

	job.Status = entity.JobStatusCancelled
	job.UpdatedAt = time.Now()
	return s.repo.Update(ctx, job)
}

func (s *jobScheduler) GetJob(ctx context.Context, envID uuid.UUID, id uuid.UUID) (*entity.NotificationJob, error) {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if job != nil && job.EnvironmentID != envID {
		return nil, nil
	}
	return job, nil
}

func (s *jobScheduler) ListJobs(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationJob, int64, error) {
	return s.repo.List(ctx, envID, limit, offset)
}

func (s *jobScheduler) ProcessJob(ctx context.Context, jobID uuid.UUID) error {
	// This should ideally run in background or handle context timeout
	job, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job == nil {
		return errors.New("job not found")
	}

	if job.Status != entity.JobStatusPending && job.Status != entity.JobStatusScheduled {
		return errors.New("job is not pending")
	}

	// Fetch Template Code
	var templateCode string
	if job.TemplateID != nil {
		tmpl, err := s.tmplManager.GetTemplate(ctx, job.EnvironmentID, *job.TemplateID)
		if err != nil {
			return err // Or mark job as failed
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
		return errors.New("template code not found")
	}

	now := time.Now()
	job.StartedAt = &now
	job.Status = entity.JobStatusProcessing
	if err := s.repo.Update(ctx, job); err != nil {
		return err
	}

	recipientsList, ok := job.RecipientsData["recipients"].([]interface{})
	if !ok {
		// Log error or fail
		job.Status = entity.JobStatusFailed
		msg := "invalid recipients data"
		job.ErrorMessage = &msg
		return s.repo.Update(ctx, job)
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

	completedAt := time.Now()
	job.CompletedAt = &completedAt
	job.TotalCount = len(recipientsList) // Update total just in case
	job.SuccessCount = success
	job.FailedCount = failed
	job.Status = entity.JobStatusCompleted
	if failed > 0 && success == 0 {
		job.Status = entity.JobStatusFailed
		msg := "all recipients failed"
		job.ErrorMessage = &msg
	}

	return s.repo.Update(ctx, job)
}

func safeUUIDString(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}
