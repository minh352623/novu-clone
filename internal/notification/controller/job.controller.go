package controller

import (
	"context"
	"net/http"
	"strconv"

	"CONVERDA/global"
	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/controller/dto"
	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/pkg/response" // Assuming standardized response package

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type JobController struct {
	scheduler service.JobScheduler
}

func NewJobController(scheduler service.JobScheduler) *JobController {
	return &JobController{scheduler: scheduler}
}

func (c *JobController) ScheduleJob(ctx *gin.Context) (interface{}, error) {
	var req dto.ScheduleJobRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}
	// Fallback to body if not in context/header?
	if envID == uuid.Nil && req.EnvironmentID != "" {
		if id, err := uuid.Parse(req.EnvironmentID); err == nil {
			envID = id
		}
	}
	if envID == uuid.Nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "environment_id is required", nil)
	}

	var templateID *uuid.UUID
	if req.TemplateID != "" {
		id, err := uuid.Parse(req.TemplateID)
		if err != nil {
			return nil, response.NewAPIError(http.StatusBadRequest, "invalid template_id", err)
		}
		templateID = &id
	}

	// Assuming TenantID is available in context or optional
	var tenantID *uuid.UUID

	job, err := c.scheduler.ScheduleJob(ctx, envID, tenantID, req.Channel, templateID, req.ScheduledAt, req.Recipients, req.Metadata)
	if err != nil {
		return nil, err
	}

	// Trigger processing immediately if not scheduled
	if job.Status == entity.JobStatusPending {
		// Asynchronous processing recommended. For MVP synchronous or fire-and-forget.
		// Go routine?
		go func() {
			defer func() {
				if r := recover(); r != nil {
					global.Logger.Error("job_controller: panic recovered in background processing", "panic", r, "jobID", job.ID.String())
				}
			}()
			// Create a background context with timeout?
			bgCtx := context.Background()
			_ = c.scheduler.ProcessJob(bgCtx, job.ID)
		}()
		job.Status = entity.JobStatusProcessing // Optimistic update for response or fetch fresh?
	}

	return toJobResponse(job), nil
}

func (c *JobController) ListJobs(ctx *gin.Context) (interface{}, error) {
	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}
	if envID == uuid.Nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "environment_id is required", nil)
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	jobs, total, err := c.scheduler.ListJobs(ctx, envID, limit, offset)
	if err != nil {
		return nil, err
	}

	jobResponses := make([]*dto.JobResponse, 0, len(jobs))
	for _, j := range jobs {
		jobResponses = append(jobResponses, toJobResponse(j))
	}

	return dto.ListJobsResponse{
		Jobs:  jobResponses,
		Total: total,
	}, nil
}

func (c *JobController) GetJob(ctx *gin.Context) (interface{}, error) {
	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "invalid id", err)
	}

	job, err := c.scheduler.GetJob(ctx, envID, id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, response.NewAPIError(http.StatusNotFound, "job not found", nil)
	}

	return toJobResponse(job), nil
}

func (c *JobController) CancelJob(ctx *gin.Context) (interface{}, error) {
	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "invalid id", err)
	}

	if err := c.scheduler.CancelJob(ctx, envID, id); err != nil {
		return nil, err
	}

	return map[string]string{"status": "cancelled"}, nil
}

func toJobResponse(job *entity.NotificationJob) *dto.JobResponse {
	if job == nil {
		return nil
	}
	return &dto.JobResponse{
		ID:             job.ID,
		EnvironmentID:  job.EnvironmentID,
		Channel:        job.Channel,
		TemplateID:     job.TemplateID,
		Status:         string(job.Status),
		TotalCount:     job.TotalCount,
		SuccessCount:   job.SuccessCount,
		FailedCount:    job.FailedCount,
		ScheduledAt:    job.ScheduledAt,
		StartedAt:      job.StartedAt,
		CompletedAt:    job.CompletedAt,
		ErrorMessage:   job.ErrorMessage,
		Metadata:       job.Metadata,
		RecipientsData: job.RecipientsData,
		CreatedAt:      job.CreatedAt,
		UpdatedAt:      job.UpdatedAt,
	}
}
