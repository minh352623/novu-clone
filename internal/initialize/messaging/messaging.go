package messaging

import (
	"CONVERDA/global"
	"CONVERDA/internal/messaging/application/service/impl"
	"CONVERDA/internal/messaging/controller"
	"CONVERDA/internal/messaging/controller/middleware"
	"CONVERDA/internal/messaging/infrastructure/gateway"
	"CONVERDA/internal/messaging/infrastructure/persistence/repository"

	"github.com/gin-gonic/gin"
)

func InitMessagingModule(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	db := global.GormDB

	// Repositories
	msgRepo := repository.NewMessageRepository(db)
	threadRepo := repository.NewThreadRepository(db)
	subRepo := repository.NewSubscriberRepository(db)
	logRepo := repository.NewAssignmentLogRepository(db)
	authRepo := repository.NewEnvironmentAuthRepository(db)

	// Services
	msgService := impl.NewConversationService(msgRepo, threadRepo, subRepo, logRepo)

	// Infrastructure
	hub := gateway.NewHub()
	go hub.Run()

	// Controllers
	convController := controller.NewConversationController(msgService, hub)

	// Middlewares
	envAuthMiddleware := middleware.EnvKeyAuth(authRepo)

	// Routes
	controller.RegisterMessagingRoutes(router, convController, authMiddleware, envAuthMiddleware)

	global.Logger.Info("Messaging module initialized")
}
