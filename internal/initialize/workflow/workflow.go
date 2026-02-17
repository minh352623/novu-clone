package workflow

import (
	"context"
	"net/http"

	notifInit "CONVERDA/internal/initialize/notification"
	"CONVERDA/internal/workflow/application/service/impl"
	"CONVERDA/internal/workflow/application/worker"
	"CONVERDA/internal/workflow/controller"
	infraRepo "CONVERDA/internal/workflow/infrastructure/persistence/repository"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InitWorkflowModule initializes the Workflow Engine module.
func InitWorkflowModule(db *gorm.DB, router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// Repositories
	workflowRepo := infraRepo.NewWorkflowRepository(db)
	execRepo := infraRepo.NewExecutionRepository(db)

	// Step handlers
	channelHandler := impl.NewChannelHandler(notifInit.NotificationService)
	delayHandler := impl.NewDelayHandler(execRepo)
	digestHandler := impl.NewDigestHandler(execRepo)

	// Services
	workflowService := impl.NewWorkflowService(workflowRepo)
	triggerService := impl.NewTriggerService(workflowRepo, execRepo, channelHandler, delayHandler)

	// Controller
	wfController := controller.NewWorkflowController(workflowService, triggerService)

	// Background worker: polls scheduled steps + digest windows
	executor := worker.NewWorkflowExecutor(execRepo, workflowRepo, channelHandler, delayHandler, digestHandler)
	go executor.Run(context.Background())

	// Routes
	wfGroup := router.Group("/workflows")
	if authMiddleware != nil {
		wfGroup.Use(authMiddleware)
	}
	{
		wfGroup.POST("", response.Wrap(wfController.CreateWorkflow, http.StatusCreated))
		wfGroup.GET("", response.Wrap(wfController.ListWorkflows, http.StatusOK))
		wfGroup.GET("/:id", response.Wrap(wfController.GetWorkflow, http.StatusOK))
		wfGroup.PATCH("/:id", response.Wrap(wfController.UpdateWorkflow, http.StatusOK))
		wfGroup.DELETE("/:id", response.Wrap(wfController.DeleteWorkflow, http.StatusNoContent))
		wfGroup.POST("/:id/toggle", response.Wrap(wfController.ToggleWorkflow, http.StatusOK))

		// Step routes
		wfGroup.POST("/:id/steps", response.Wrap(wfController.AddStep, http.StatusCreated))
		wfGroup.PATCH("/:id/steps/:stepId", response.Wrap(wfController.UpdateStep, http.StatusOK))
		wfGroup.DELETE("/:id/steps/:stepId", response.Wrap(wfController.DeleteStep, http.StatusNoContent))

		// Trigger route
		wfGroup.POST("/trigger", response.Wrap(wfController.TriggerWorkflow, http.StatusAccepted))
	}
}
