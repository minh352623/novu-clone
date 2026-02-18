package iam

import (
	"CONVERDA/global"
	"CONVERDA/internal/iam/application/service/impl"
	"CONVERDA/internal/iam/controller"
	"CONVERDA/internal/iam/infrastructure/email"
	"CONVERDA/internal/iam/infrastructure/persistence/repository"
	"CONVERDA/internal/middleware"

	"go.uber.org/zap"
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

	// Initialize services
	emailService, err := email.NewSESEmailService(global.Config.SES)
	if err != nil {
		global.Logger.Fatal("Failed to initialize SES email service", zap.Error(err))
	}

	authService := impl.NewAuthService(userRepo)
	tenantService := impl.NewTenantService(tenantRepo, memberRepo, roleRepo, pricingPlanRepo)
	memberService := impl.NewMemberService(memberRepo, roleRepo, tenantRepo, invitationRepo, userRepo, emailService)
	roleService := impl.NewRoleService(roleRepo)
	pricingPlanService := impl.NewPricingPlanService(pricingPlanRepo)
	tokenService := impl.NewTokenService()

	// Middleware Adapters
	membershipChecker := middleware.NewMembershipCheckerAdapter(memberRepo)
	permissionChecker := middleware.NewPermissionCheckerAdapter(roleRepo)
	tenantMembershipMiddleware := middleware.TenantMembershipMiddleware(membershipChecker)

	// Controllers
	authController := controller.NewAuthController(authService, tokenService)
	tenantController := controller.NewTenantController(tenantService, memberService)
	roleController := controller.NewRoleController(roleService, memberService)
	pricingPlanController := controller.NewPricingPlanController(pricingPlanService)

	// Register Routes using the controller's register function
	controller.RegisterIAMRoutes(
		router,
		authController,
		tenantController,
		roleController,
		pricingPlanController,
		authMiddleware,
		tenantMembershipMiddleware,
		permissionChecker,
	)

	global.Logger.Info("IAM module initialized")
}
