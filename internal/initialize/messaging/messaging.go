package messaging

import (
	"CONVERDA/global"
	appsRepo "CONVERDA/internal/apps/infrastructure/persistence/repository"
	iamRepo "CONVERDA/internal/iam/infrastructure/persistence/repository"
	"CONVERDA/internal/messaging/application/service/impl"
	"CONVERDA/internal/messaging/application/worker"
	"CONVERDA/internal/messaging/controller"
	"CONVERDA/internal/messaging/controller/middleware"
	"CONVERDA/internal/messaging/infrastructure/adapter"
	"CONVERDA/internal/messaging/infrastructure/gateway"
	"CONVERDA/internal/messaging/infrastructure/persistence/repository"
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitMessagingModule(v1 *gin.RouterGroup, db *gorm.DB, authMiddleware gin.HandlerFunc) {

	// Repositories
	msgRepo := repository.NewMessageRepository(db)
	threadRepo := repository.NewThreadRepository(db)
	subRepo := repository.NewSubscriberRepository(db)
	logRepo := repository.NewAssignmentLogRepository(db)
	authRepo := repository.NewEnvironmentAuthRepository(db)
	memberRepo := iamRepo.NewTenantMemberRepository(db)
	appRepo := appsRepo.NewAppRepository(db)
	envRepo := appsRepo.NewEnvironmentRepository(db)
	messagingUoW := repository.NewMessagingUnitOfWork(db)

	// Adapters
	appReader := adapter.NewLocalAppAdapter(appRepo, envRepo)
	memberReader := adapter.NewLocalMemberAdapter(memberRepo)

	// Infrastructure
	hub := gateway.NewHub()
	go hub.Run()

	// Services
	msgService := impl.NewConversationService(msgRepo, threadRepo, subRepo, logRepo, memberReader, appReader, hub, messagingUoW)

	slaWorker := worker.NewSLAWorker(threadRepo, logRepo, appReader)
	go slaWorker.Run(context.Background())

	// Controllers
	convController := controller.NewConversationController(msgService, hub)

	// Middlewares
	envAuthMiddleware := middleware.EnvKeyAuth(authRepo)

	// Routes
	controller.RegisterMessagingRoutes(v1, convController, authMiddleware, envAuthMiddleware)

	global.Logger.Info("Messaging module initialized")
}
