package controller

import (
	"net/http"

	"CONVERDA/internal/middleware"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
)

// RegisterIAMRoutes registers all IAM module routes
func RegisterIAMRoutes(
	router *gin.RouterGroup,
	authController *AuthController,
	tenantController *TenantController,
	roleController *RoleController,
	pricingPlanController *PricingPlanController,
	authMiddleware gin.HandlerFunc,
	tenantMembershipMiddleware gin.HandlerFunc,
	permissionChecker middleware.PermissionChecker,
) {
	// Auth routes (public)
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", response.Wrap(authController.Register, http.StatusCreated))
		authGroup.POST("/login", response.Wrap(authController.Login, http.StatusOK))
		authGroup.POST("/refresh", response.Wrap(authController.RefreshToken, http.StatusOK))
	}

	// User routes (authenticated)
	userGroup := router.Group("/users")
	if authMiddleware != nil {
		userGroup.Use(authMiddleware)
	}
	{
		userGroup.GET("/me", response.Wrap(authController.Me, http.StatusOK))
		userGroup.PUT("/change-password", response.Wrap(authController.ChangePassword, http.StatusOK))
	}

	// Tenant routes (authenticated + membership check for specific tenant operations)
	tenantGroup := router.Group("/tenants")
	if authMiddleware != nil {
		tenantGroup.Use(authMiddleware)
	}
	{
		// Create tenant doesn't require membership (user is creating new tenant)
		tenantGroup.POST("", response.Wrap(tenantController.CreateTenant, http.StatusCreated))
		tenantGroup.GET("", response.Wrap(tenantController.ListTenants, http.StatusOK))
		// Get my tenants - MUST BE BEFORE /:id
		tenantGroup.GET("/me", response.Wrap(tenantController.GetMyTenants, http.StatusOK))
	}

	// Tenant-specific routes (require membership + permissions)
	tenantSpecificGroup := router.Group("/tenants/:id")
	if authMiddleware != nil {
		tenantSpecificGroup.Use(authMiddleware)
	}
	if tenantMembershipMiddleware != nil {
		tenantSpecificGroup.Use(tenantMembershipMiddleware)
	}
	{
		tenantSpecificGroup.GET("", response.Wrap(tenantController.GetTenant, http.StatusOK))

		// Update/Delete require admin permission
		if permissionChecker != nil {
			tenantSpecificGroup.PUT("", middleware.RequirePermission(permissionChecker, "iam.tenants.update"), response.Wrap(tenantController.UpdateTenant, http.StatusOK))
			tenantSpecificGroup.DELETE("", middleware.RequirePermission(permissionChecker, "iam.tenants.delete"), response.Wrap(tenantController.DeleteTenant, http.StatusOK))
		} else {
			tenantSpecificGroup.PUT("", response.Wrap(tenantController.UpdateTenant, http.StatusOK))
			tenantSpecificGroup.DELETE("", response.Wrap(tenantController.DeleteTenant, http.StatusOK))
		}

		// Tenant Members
		tenantSpecificGroup.GET("/members", response.Wrap(tenantController.GetMembers, http.StatusOK))
		if permissionChecker != nil {
			tenantSpecificGroup.POST("/members", middleware.RequirePermission(permissionChecker, "iam.members.create"), response.Wrap(tenantController.AddMember, http.StatusCreated))
		} else {
			tenantSpecificGroup.POST("/members", response.Wrap(tenantController.AddMember, http.StatusCreated))
		}

		// Invitations
		if permissionChecker != nil {
			tenantSpecificGroup.POST("/invitations", middleware.RequirePermission(permissionChecker, "iam.invitations.create"), response.Wrap(tenantController.InviteMember, http.StatusCreated))
		} else {
			tenantSpecificGroup.POST("/invitations", response.Wrap(tenantController.InviteMember, http.StatusCreated))
		}
	}

	// Invitation routes (authenticated)
	invitationGroup := router.Group("/invitations")
	if authMiddleware != nil {
		invitationGroup.Use(authMiddleware)
	}
	{
		invitationGroup.POST("/accept", response.Wrap(tenantController.AcceptInvitation, http.StatusOK))
	}

	// Role routes (authenticated + permission check)
	roleGroup := router.Group("/roles")
	if authMiddleware != nil {
		roleGroup.Use(authMiddleware)
	}
	{
		roleGroup.GET("", response.Wrap(roleController.ListRoles, http.StatusOK))
		roleGroup.GET("/:id", response.Wrap(roleController.GetRole, http.StatusOK))

		// Write operations require permissions
		if permissionChecker != nil {
			roleGroup.POST("", middleware.RequirePermission(permissionChecker, "iam.roles.create"), response.Wrap(roleController.CreateRole, http.StatusCreated))
			roleGroup.PUT("/:id", middleware.RequirePermission(permissionChecker, "iam.roles.update"), response.Wrap(roleController.UpdateRole, http.StatusOK))
			roleGroup.DELETE("/:id", middleware.RequirePermission(permissionChecker, "iam.roles.delete"), response.Wrap(roleController.DeleteRole, http.StatusOK))
			roleGroup.POST("/:id/assign", middleware.RequirePermission(permissionChecker, "iam.roles.assign"), response.Wrap(roleController.AssignRole, http.StatusOK))
			roleGroup.POST("/:id/revoke", middleware.RequirePermission(permissionChecker, "iam.roles.revoke"), response.Wrap(roleController.RevokeRole, http.StatusOK))
		} else {
			roleGroup.POST("", response.Wrap(roleController.CreateRole, http.StatusCreated))
			roleGroup.PUT("/:id", response.Wrap(roleController.UpdateRole, http.StatusOK))
			roleGroup.DELETE("/:id", response.Wrap(roleController.DeleteRole, http.StatusOK))
			roleGroup.POST("/:id/assign", response.Wrap(roleController.AssignRole, http.StatusOK))
			roleGroup.POST("/:id/revoke", response.Wrap(roleController.RevokeRole, http.StatusOK))
		}
	}

	// Pricing Plan routes
	// Public Pricing Plans (read-only)
	pricingGroupPublic := router.Group("/pricing-plans")
	{
		pricingGroupPublic.GET("", response.Wrap(pricingPlanController.ListPricingPlans, http.StatusOK))
		pricingGroupPublic.GET("/:id", response.Wrap(pricingPlanController.GetPricingPlan, http.StatusOK))
	}

	// Private Pricing Plans (Manage - system admin only)
	pricingGroup := router.Group("/pricing-plans")
	if authMiddleware != nil {
		pricingGroup.Use(authMiddleware)
	}
	{
		if permissionChecker != nil {
			pricingGroup.POST("", middleware.RequirePermission(permissionChecker, "system.pricing.create"), response.Wrap(pricingPlanController.CreatePricingPlan, http.StatusCreated))
			pricingGroup.PUT("/:id", middleware.RequirePermission(permissionChecker, "system.pricing.update"), response.Wrap(pricingPlanController.UpdatePricingPlan, http.StatusOK))
			pricingGroup.DELETE("/:id", middleware.RequirePermission(permissionChecker, "system.pricing.delete"), response.Wrap(pricingPlanController.DeletePricingPlan, http.StatusOK))
			pricingGroup.POST("/:id/default", middleware.RequirePermission(permissionChecker, "system.pricing.update"), response.Wrap(pricingPlanController.SetAsDefault, http.StatusOK))
		} else {
			pricingGroup.POST("", response.Wrap(pricingPlanController.CreatePricingPlan, http.StatusCreated))
			pricingGroup.PUT("/:id", response.Wrap(pricingPlanController.UpdatePricingPlan, http.StatusOK))
			pricingGroup.DELETE("/:id", response.Wrap(pricingPlanController.DeletePricingPlan, http.StatusOK))
			pricingGroup.POST("/:id/default", response.Wrap(pricingPlanController.SetAsDefault, http.StatusOK))
		}
	}
}
