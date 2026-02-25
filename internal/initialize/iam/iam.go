package iam

import (
	"CONVERDA/global"
	"CONVERDA/internal/iam/application/service/impl"
	"CONVERDA/internal/iam/controller"
	"CONVERDA/internal/iam/infrastructure/email"
	"CONVERDA/internal/iam/infrastructure/persistence/repository"
	"CONVERDA/internal/middleware"
	"CONVERDA/pkg/gdpr"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// InitIAMModule initializes the IAM module
func InitIAMModule(router *gin.RouterGroup, db *gorm.DB, authMiddleware gin.HandlerFunc) {

	// Repositories
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	memberRepo := repository.NewTenantMemberRepository(db)
	invitationRepo := repository.NewInvitationRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	pricingPlanRepo := repository.NewPricingPlanRepository(db)
	iamUoW := repository.NewIAMUnitOfWork(db)

	// Initialize services
	emailService, err := email.NewSESEmailService(global.Config.SES)
	if err != nil {
		global.Logger.Fatal("Failed to initialize SES email service", "error", err)
	}

	authService := impl.NewAuthService(userRepo)
	tenantService := impl.NewTenantService(tenantRepo, memberRepo, roleRepo, pricingPlanRepo, iamUoW)
	memberService := impl.NewMemberService(memberRepo, roleRepo, tenantRepo, invitationRepo, userRepo, emailService, iamUoW)
	roleService := impl.NewRoleService(roleRepo)
	pricingPlanService := impl.NewPricingPlanService(pricingPlanRepo)
	tokenService := impl.NewTokenService()

	// Middleware Adapters
	membershipChecker := middleware.NewMembershipCheckerAdapter(memberRepo, tenantRepo)
	permissionChecker := middleware.NewPermissionCheckerAdapter(roleRepo)
	tenantMembershipMiddleware := middleware.TenantMembershipMiddleware(membershipChecker)

	// GDPR Service + Providers
	gdprService := gdpr.NewService(
		gdpr.NewProfileProvider(db),
		gdpr.NewMessageProvider(db),
		gdpr.NewSubscriptionProvider(db),
		gdpr.NewAuditLogProvider(db),
	)

	// Controllers
	authController := controller.NewAuthController(authService, tokenService)
	tenantController := controller.NewTenantController(tenantService, memberService)
	roleController := controller.NewRoleController(roleService, memberService)
	pricingPlanController := controller.NewPricingPlanController(pricingPlanService)
	gdprController := controller.NewGDPRController(gdprService)

	// Register Routes using the controller's register function
	controller.RegisterIAMRoutes(
		router,
		authController,
		tenantController,
		roleController,
		pricingPlanController,
		gdprController,
		authMiddleware,
		tenantMembershipMiddleware,
		permissionChecker,
	)

	global.Logger.Info("IAM module initialized")
}
