package controller

import (
	"CONVERDA/internal/workflow/application/service"
	"CONVERDA/internal/workflow/controller/dto"
	"CONVERDA/internal/workflow/domain"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WorkflowController struct {
	svc        service.WorkflowService
	triggerSvc service.TriggerService
}

func NewWorkflowController(svc service.WorkflowService, triggerSvc service.TriggerService) *WorkflowController {
	return &WorkflowController{svc: svc, triggerSvc: triggerSvc}
}

// CreateWorkflow godoc
// @Summary     Create a new workflow
// @Tags        Workflows
// @Accept      json
// @Produce     json
// @Param       environment_id query string true "Environment ID"
// @Param       body body dto.CreateWorkflowRequest true "Workflow data"
// @Success     201 {object} dto.WorkflowResponse
// @Router      /workflows [post]
func (c *WorkflowController) CreateWorkflow(ctx *gin.Context) (interface{}, error) {
	envIDStr := ctx.Query("environment_id")
	envID, err := uuid.Parse(envIDStr)
	if err != nil {
		return nil, response.NewBadRequestError("Invalid environment_id")
	}

	var req dto.CreateWorkflowRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError("Invalid request body")
	}

	resp, err := c.svc.CreateWorkflow(ctx.Request.Context(), envID, req)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}
	return resp, nil
}

// GetWorkflow godoc
// @Summary     Get workflow by ID
// @Tags        Workflows
// @Produce     json
// @Param       id path string true "Workflow ID"
// @Success     200 {object} dto.WorkflowResponse
// @Router      /workflows/{id} [get]
func (c *WorkflowController) GetWorkflow(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid workflow ID")
	}

	wf, err := c.svc.GetWorkflow(ctx.Request.Context(), id)
	if err != nil {
		if err == domain.ErrWorkflowNotFound {
			return nil, response.NewNotFoundError("Workflow not found")
		}
		return nil, response.NewInternalServerError(err.Error())
	}
	return wf, nil
}

// ListWorkflows godoc
// @Summary     List workflows by environment
// @Tags        Workflows
// @Produce     json
// @Param       environment_id query string true "Environment ID"
// @Success     200 {array} dto.WorkflowResponse
// @Router      /workflows [get]
func (c *WorkflowController) ListWorkflows(ctx *gin.Context) (interface{}, error) {
	envIDStr := ctx.Query("environment_id")
	envID, err := uuid.Parse(envIDStr)
	if err != nil {
		return nil, response.NewBadRequestError("Invalid environment_id")
	}

	resp, err := c.svc.ListWorkflows(ctx.Request.Context(), envID)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}
	return resp, nil
}

// UpdateWorkflow godoc
// @Summary     Update a workflow
// @Tags        Workflows
// @Accept      json
// @Produce     json
// @Param       id path string true "Workflow ID"
// @Param       body body dto.UpdateWorkflowRequest true "Update data"
// @Success     200 {object} dto.WorkflowResponse
// @Router      /workflows/{id} [patch]
func (c *WorkflowController) UpdateWorkflow(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid workflow ID")
	}

	var req dto.UpdateWorkflowRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError("Invalid request body")
	}

	resp, err := c.svc.UpdateWorkflow(ctx.Request.Context(), id, req)
	if err != nil {
		if err == domain.ErrWorkflowNotFound {
			return nil, response.NewNotFoundError("Workflow not found")
		}
		return nil, response.NewInternalServerError(err.Error())
	}
	return resp, nil
}

// DeleteWorkflow godoc
// @Summary     Delete a workflow
// @Tags        Workflows
// @Param       id path string true "Workflow ID"
// @Success     204
// @Router      /workflows/{id} [delete]
func (c *WorkflowController) DeleteWorkflow(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid workflow ID")
	}

	if err := c.svc.DeleteWorkflow(ctx.Request.Context(), id); err != nil {
		if err == domain.ErrWorkflowNotFound {
			return nil, response.NewNotFoundError("Workflow not found")
		}
		if err == domain.ErrWorkflowActive {
			return nil, response.NewBadRequestError("Cannot delete active workflow")
		}
		return nil, response.NewInternalServerError(err.Error())
	}
	return nil, nil
}

// ToggleWorkflow godoc
// @Summary     Toggle workflow active state
// @Tags        Workflows
// @Produce     json
// @Param       id path string true "Workflow ID"
// @Success     200 {object} dto.WorkflowResponse
// @Router      /workflows/{id}/toggle [post]
func (c *WorkflowController) ToggleWorkflow(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid workflow ID")
	}

	resp, err := c.svc.ToggleWorkflow(ctx.Request.Context(), id)
	if err != nil {
		if err == domain.ErrWorkflowNotFound {
			return nil, response.NewNotFoundError("Workflow not found")
		}
		return nil, response.NewInternalServerError(err.Error())
	}
	return resp, nil
}

// --- Step endpoints ---

// AddStep godoc
// @Summary     Add a step to a workflow
// @Tags        Workflows
// @Accept      json
// @Produce     json
// @Param       id path string true "Workflow ID"
// @Param       body body dto.CreateStepRequest true "Step data"
// @Success     201 {object} dto.StepResponse
// @Router      /workflows/{id}/steps [post]
func (c *WorkflowController) AddStep(ctx *gin.Context) (interface{}, error) {
	workflowID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid workflow ID")
	}

	var req dto.CreateStepRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError("Invalid request body")
	}

	resp, err := c.svc.AddStep(ctx.Request.Context(), workflowID, req)
	if err != nil {
		if err == domain.ErrWorkflowNotFound {
			return nil, response.NewNotFoundError("Workflow not found")
		}
		return nil, response.NewInternalServerError(err.Error())
	}
	return resp, nil
}

// UpdateStep godoc
// @Summary     Update a workflow step
// @Tags        Workflows
// @Accept      json
// @Produce     json
// @Param       id path string true "Workflow ID"
// @Param       stepId path string true "Step ID"
// @Param       body body dto.UpdateStepRequest true "Update data"
// @Success     200 {object} dto.StepResponse
// @Router      /workflows/{id}/steps/{stepId} [patch]
func (c *WorkflowController) UpdateStep(ctx *gin.Context) (interface{}, error) {
	stepID, err := uuid.Parse(ctx.Param("stepId"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid step ID")
	}

	var req dto.UpdateStepRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError("Invalid request body")
	}

	resp, err := c.svc.UpdateStep(ctx.Request.Context(), stepID, req)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}
	return resp, nil
}

// DeleteStep godoc
// @Summary     Delete a workflow step
// @Tags        Workflows
// @Param       id path string true "Workflow ID"
// @Param       stepId path string true "Step ID"
// @Success     204
// @Router      /workflows/{id}/steps/{stepId} [delete]
func (c *WorkflowController) DeleteStep(ctx *gin.Context) (interface{}, error) {
	stepID, err := uuid.Parse(ctx.Param("stepId"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid step ID")
	}

	if err := c.svc.DeleteStep(ctx.Request.Context(), stepID); err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}
	return nil, nil
}

// --- Trigger endpoint ---

// TriggerWorkflow godoc
// @Summary     Trigger a workflow execution
// @Tags        Workflows
// @Accept      json
// @Produce     json
// @Param       body body dto.TriggerWorkflowRequest true "Trigger data"
// @Success     202 {object} map[string]string
// @Router      /workflows/trigger [post]
func (c *WorkflowController) TriggerWorkflow(ctx *gin.Context) (interface{}, error) {
	var req dto.TriggerWorkflowRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError("Invalid request body")
	}

	if err := c.triggerSvc.Trigger(ctx.Request.Context(),
		req.EnvironmentID, req.TriggerIdentifier, req.SubscriberKey, req.Payload); err != nil {
		if err == domain.ErrWorkflowNotFound {
			return nil, response.NewNotFoundError("Workflow not found")
		}
		return nil, response.NewInternalServerError(err.Error())
	}

	return map[string]string{"status": "triggered"}, nil
}
