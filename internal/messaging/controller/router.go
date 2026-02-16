package controller

import (
	"net/http"

	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
)

func RegisterMessagingRoutes(
	router *gin.RouterGroup,
	conversationController *ConversationController,
	authMiddleware gin.HandlerFunc,
	envAuthMiddleware gin.HandlerFunc,
) {
	group := router.Group("/conversations")

	// Public or Webhook/ApiKey Protected
	group.POST("/inbound", envAuthMiddleware, response.Wrap(conversationController.InboundMessage, http.StatusCreated))

	// WebSocket
	group.GET("/ws", conversationController.ServeWS)

	// Agent Protected
	if authMiddleware != nil {
		group.Use(authMiddleware)
	}
	{
		group.GET("", response.Wrap(conversationController.ListConversations, http.StatusOK))
		group.GET("/:id", response.Wrap(conversationController.GetThread, http.StatusOK))

		// Assignment & Status
		group.PATCH("/:id/assign", response.Wrap(conversationController.AssignThread, http.StatusOK))
		group.POST("/:id/unassign", response.Wrap(conversationController.UnassignThread, http.StatusOK))
		group.PATCH("/assign", response.Wrap(conversationController.BulkAssignThreads, http.StatusOK)) // Bulk
		group.POST("/:id/resolve", response.Wrap(conversationController.ResolveThread, http.StatusOK))

		// Internal Chat (Direct & Group)
		group.POST("/direct", response.Wrap(conversationController.CreateDirectChat, http.StatusOK))
		group.POST("/group", response.Wrap(conversationController.CreateGroupChat, http.StatusCreated))
		group.PATCH("/group/:id", response.Wrap(conversationController.UpdateGroupChat, http.StatusOK))
		group.POST("/group/:id/participants", response.Wrap(conversationController.AddGroupParticipants, http.StatusOK))
		group.DELETE("/group/:id/participants/:memberID", response.Wrap(conversationController.RemoveGroupParticipant, http.StatusOK))
		group.POST("/:id/read", response.Wrap(conversationController.MarkAsRead, http.StatusOK))

		// Messages & Notes
		group.POST("/:id/messages", response.Wrap(conversationController.ReplyMessage, http.StatusCreated))
		group.POST("/:id/notes", response.Wrap(conversationController.AddInternalNote, http.StatusCreated))
	}

	// Dashboards Group
	dashboardGroup := router.Group("/dashboards")
	if authMiddleware != nil {
		dashboardGroup.Use(authMiddleware)
	}
	{
		dashboardGroup.GET("/team", response.Wrap(conversationController.GetTeamDashboard, http.StatusOK))
		dashboardGroup.GET("/partner", response.Wrap(conversationController.GetPartnerDashboard, http.StatusOK))
		dashboardGroup.GET("/agent/:member_id", response.Wrap(conversationController.GetAgentDashboard, http.StatusOK))
	}
}
