package controller

import (
	"CONVERDA/internal/iam/application/service"
	"CONVERDA/internal/iam/controller/dto"
	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PricingPlanController handles HTTP requests for pricing plans
type PricingPlanController struct {
	planService service.PricingPlanService
}

// NewPricingPlanController creates a new PricingPlanController
func NewPricingPlanController(planService service.PricingPlanService) *PricingPlanController {
	return &PricingPlanController{
		planService: planService,
	}
}

// CreatePricingPlan creates a new pricing plan
// @Summary Create pricing plan
// @Description Create a new pricing plan
// @Tags Pricing Plans
// @Accept json
// @Produce json
// @Param request body dto.CreatePricingPlanRequest true "Create Pricing Plan Request"
// @Success 201 {object} dto.PricingPlanResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pricing-plans [post]
func (c *PricingPlanController) CreatePricingPlan(ctx *gin.Context) (interface{}, error) {
	var req dto.CreatePricingPlanRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	plan, err := c.planService.CreatePricingPlan(ctx.Request.Context(), req.Name, req.Slug, req.MonthlyCredits, req.Price, req.Currency)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return dto.ToPricingPlanResponse(plan), nil
}

// GetPricingPlan gets a pricing plan by ID
// @Summary Get pricing plan
// @Description Get a pricing plan by ID
// @Tags Pricing Plans
// @Accept json
// @Produce json
// @Param id path string true "Pricing Plan ID"
// @Success 200 {object} dto.PricingPlanResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pricing-plans/{id} [get]
func (c *PricingPlanController) GetPricingPlan(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid ID format")
	}

	plan, err := c.planService.GetPricingPlan(ctx.Request.Context(), id)
	if err != nil {
		if err == service.ErrPlanNotFound {
			return nil, response.NewNotFoundError("Pricing plan not found")
		}
		return nil, response.NewInternalServerError(err.Error())
	}

	return dto.ToPricingPlanResponse(plan), nil
}

// ListPricingPlans lists request pricing plans
// @Summary List pricing plans
// @Description List pricing plans with filtering and pagination
// @Tags Pricing Plans
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param search query string false "Search term"
// @Param is_active query bool false "Filter by active status"
// @Param is_default query bool false "Filter by default status"
// @Success 200 {object} dto.PaginatedResponse
// @Failure 500 {object} map[string]string
// @Router /pricing-plans [get]
func (c *PricingPlanController) ListPricingPlans(ctx *gin.Context) (interface{}, error) {
	var params dto.PaginationParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	filters := repository.PricingPlanFilters{
		Limit:  params.PageSize,
		Offset: (params.Page - 1) * params.PageSize,
		Search: params.Search,
	}

	if val, ok := ctx.GetQuery("is_active"); ok {
		active := val == "true"
		filters.IsActive = &active
	}
	if val, ok := ctx.GetQuery("is_default"); ok {
		isDefault := val == "true"
		filters.IsDefault = &isDefault
	}

	plans, total, err := c.planService.ListPricingPlans(ctx.Request.Context(), filters)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return dto.NewPaginatedResponse(dto.ToPricingPlanResponseList(plans), total, params.Page, params.PageSize), nil
}

// UpdatePricingPlan updates a pricing plan
// @Summary Update pricing plan
// @Description Update a pricing plan by ID
// @Tags Pricing Plans
// @Accept json
// @Produce json
// @Param id path string true "Pricing Plan ID"
// @Param request body dto.UpdatePricingPlanRequest true "Update Pricing Plan Request"
// @Success 200 {object} dto.PricingPlanResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pricing-plans/{id} [put]
func (c *PricingPlanController) UpdatePricingPlan(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid ID format")
	}

	var req dto.UpdatePricingPlanRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	plan, err := c.planService.GetPricingPlan(ctx.Request.Context(), id)
	if err != nil {
		if err == service.ErrPlanNotFound {
			return nil, response.NewNotFoundError("Pricing plan not found")
		}
		return nil, response.NewInternalServerError(err.Error())
	}

	// Update fields
	updatedStruct := &entity.PricingPlan{
		ID:             plan.ID,
		Name:           plan.Name,
		Slug:           plan.Slug,
		MonthlyCredits: plan.MonthlyCredits,
		Price:          plan.Price,
		Currency:       plan.Currency,
		IsActive:       plan.IsActive,
		IsDefault:      plan.IsDefault,
		CreatedAt:      plan.CreatedAt,
		// UpdatedAt will be handled by repo/gorm
	}

	if req.Name != nil {
		updatedStruct.Name = *req.Name
	}
	if req.MonthlyCredits != nil {
		updatedStruct.MonthlyCredits = *req.MonthlyCredits
	}
	if req.Price != nil {
		updatedStruct.Price = *req.Price
	}
	if req.Currency != nil {
		updatedStruct.Currency = *req.Currency
	}
	if req.IsActive != nil {
		updatedStruct.IsActive = *req.IsActive
	}

	if err := c.planService.UpdatePricingPlan(ctx.Request.Context(), updatedStruct); err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	// Fetch updated
	updatedPlan, err := c.planService.GetPricingPlan(ctx.Request.Context(), id)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return dto.ToPricingPlanResponse(updatedPlan), nil
}

// DeletePricingPlan deletes a pricing plan
// @Summary Delete pricing plan
// @Description Delete a pricing plan by ID
// @Tags Pricing Plans
// @Accept json
// @Produce json
// @Param id path string true "Pricing Plan ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pricing-plans/{id} [delete]
func (c *PricingPlanController) DeletePricingPlan(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid ID format")
	}

	if err := c.planService.DeletePricingPlan(ctx.Request.Context(), id); err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return nil, nil // Wrapper handles 204
}

// SetAsDefault sets a pricing plan as default
// @Summary Set pricing plan as default
// @Description Set a pricing plan as default
// @Tags Pricing Plans
// @Accept json
// @Produce json
// @Param id path string true "Pricing Plan ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /pricing-plans/{id}/set-default [post]
func (c *PricingPlanController) SetAsDefault(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid ID format")
	}

	if err := c.planService.SetAsDefault(ctx.Request.Context(), id); err != nil {
		if err == service.ErrPlanNotFound {
			return nil, response.NewNotFoundError("Pricing plan not found")
		}
		return nil, response.NewInternalServerError(err.Error())
	}

	return gin.H{"message": "Pricing plan set as default successfully"}, nil
}
