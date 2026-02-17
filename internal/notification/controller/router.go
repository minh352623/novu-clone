package controller

import (
	"CONVERDA/internal/notification/application/service"
	"CONVERDA/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, svc service.NotificationService, groupMgr service.GroupManager, layoutMgr service.LayoutManager, jobScheduler service.JobScheduler, tmplManager *service.TemplateManager) {
	// Notifications
	notifController := NewNotificationController(svc)
	groupController := NewGroupController(groupMgr)
	layoutController := NewLayoutController(layoutMgr)
	jobController := NewJobController(jobScheduler)
	templateContentController := NewTemplateContentController(tmplManager)

	// Group: /notifications
	g := r.Group("/notifications")
	{
		g.POST("/send", response.Wrap(notifController.SendNotification, http.StatusOK))

		// Groups
		groups := g.Group("/groups")
		{
			groups.POST("", response.Wrap(groupController.CreateGroup, http.StatusCreated))
			groups.GET("", response.Wrap(groupController.ListGroups, http.StatusOK))
			groups.GET("/:id", response.Wrap(groupController.GetGroup, http.StatusOK))
			groups.PATCH("/:id", response.Wrap(groupController.UpdateGroup, http.StatusOK))
			groups.DELETE("/:id", response.Wrap(groupController.DeleteGroup, http.StatusOK))
		}

		// Layouts
		layouts := g.Group("/layouts")
		{
			layouts.POST("", response.Wrap(layoutController.CreateLayout, http.StatusCreated))
			layouts.GET("", response.Wrap(layoutController.ListLayouts, http.StatusOK))
			layouts.GET("/:id", response.Wrap(layoutController.GetLayout, http.StatusOK))
			layouts.PATCH("/:id", response.Wrap(layoutController.UpdateLayout, http.StatusOK))
			layouts.DELETE("/:id", response.Wrap(layoutController.DeleteLayout, http.StatusOK))
		}

		// Jobs
		jobs := g.Group("/jobs")
		{
			jobs.POST("", response.Wrap(jobController.ScheduleJob, http.StatusCreated))
			jobs.GET("", response.Wrap(jobController.ListJobs, http.StatusOK))
			jobs.GET("/:id", response.Wrap(jobController.GetJob, http.StatusOK))
			jobs.POST("/:id/cancel", response.Wrap(jobController.CancelJob, http.StatusOK))
		}

		// Template Content (i18n)
		templates := g.Group("/templates")
		{
			templates.POST("/:id/content", response.Wrap(templateContentController.AddContent, http.StatusCreated))
			templates.PUT("/:id/content/:lang", response.Wrap(templateContentController.UpdateContent, http.StatusOK))
			templates.GET("/:id/languages", response.Wrap(templateContentController.ListLanguages, http.StatusOK))
		}
	}
}
