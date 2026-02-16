package mapper

import (
	"encoding/json"

	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/infrastructure/persistence/model"

	"gorm.io/datatypes"
)

// PricingPlanMapper handles mapping between entity and model
type PricingPlanMapper struct{}

func NewPricingPlanMapper() *PricingPlanMapper {
	return &PricingPlanMapper{}
}

func (m *PricingPlanMapper) ToDomain(model *model.PricingPlanModel) *entity.PricingPlan {
	if model == nil {
		return nil
	}
	return &entity.PricingPlan{
		ID:             model.ID,
		Name:           model.Name,
		Slug:           model.Slug,
		MonthlyCredits: model.MonthlyCredits,
		Price:          model.Price,
		Currency:       model.Currency,
		Description:    model.Description,
		IsActive:       model.IsActive,
		IsDefault:      model.IsDefault,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}
}

func (m *PricingPlanMapper) ToModel(entity *entity.PricingPlan) *model.PricingPlanModel {
	if entity == nil {
		return nil
	}
	return &model.PricingPlanModel{
		ID:             entity.ID,
		Name:           entity.Name,
		Slug:           entity.Slug,
		MonthlyCredits: entity.MonthlyCredits,
		Price:          entity.Price,
		Currency:       entity.Currency,
		Description:    entity.Description,
		IsActive:       entity.IsActive,
		IsDefault:      entity.IsDefault,
		CreatedAt:      entity.CreatedAt,
		UpdatedAt:      entity.UpdatedAt,
	}
}

// TenantMapper handles mapping between entity and model
type TenantMapper struct {
	planMapper *PricingPlanMapper
}

func NewTenantMapper() *TenantMapper {
	return &TenantMapper{
		planMapper: NewPricingPlanMapper(),
	}
}

func (m *TenantMapper) ToDomain(model *model.TenantModel) *entity.Tenant {
	if model == nil {
		return nil
	}
	tenant := &entity.Tenant{
		ID:            model.ID,
		Name:          model.Name,
		Slug:          model.Slug,
		PricingPlanID: model.PricingPlanID,
		PlanStartDate: model.PlanStartDate,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
	if model.PricingPlan != nil {
		tenant.PricingPlan = m.planMapper.ToDomain(model.PricingPlan)
	}
	return tenant
}

func (m *TenantMapper) ToModel(entity *entity.Tenant) *model.TenantModel {
	if entity == nil {
		return nil
	}
	return &model.TenantModel{
		ID:            entity.ID,
		Name:          entity.Name,
		Slug:          entity.Slug,
		PricingPlanID: entity.PricingPlanID,
		PlanStartDate: entity.PlanStartDate,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
	}
}

// UserMapper handles mapping between entity and model
type UserMapper struct{}

func NewUserMapper() *UserMapper {
	return &UserMapper{}
}

func (m *UserMapper) ToDomain(model *model.UserModel) *entity.User {
	if model == nil {
		return nil
	}
	return &entity.User{
		ID:           model.ID,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		FullName:     model.FullName,
		IsRootAdmin:  model.IsRootAdmin,
		LastLoginAt:  model.LastLoginAt,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func (m *UserMapper) ToModel(entity *entity.User) *model.UserModel {
	if entity == nil {
		return nil
	}
	return &model.UserModel{
		ID:           entity.ID,
		Email:        entity.Email,
		PasswordHash: entity.PasswordHash,
		FullName:     entity.FullName,
		IsRootAdmin:  entity.IsRootAdmin,
		LastLoginAt:  entity.LastLoginAt,
		CreatedAt:    entity.CreatedAt,
		UpdatedAt:    entity.UpdatedAt,
	}
}

// RoleMapper handles mapping between entity and model
type RoleMapper struct{}

func NewRoleMapper() *RoleMapper {
	return &RoleMapper{}
}

func (m *RoleMapper) ToDomain(model *model.RoleModel) *entity.Role {
	if model == nil {
		return nil
	}

	var permissions map[string]interface{}
	if model.Permissions != nil {
		json.Unmarshal(model.Permissions, &permissions)
	}
	if permissions == nil {
		permissions = make(map[string]interface{})
	}

	return &entity.Role{
		ID:          model.ID,
		Name:        model.Name,
		Slug:        model.Slug,
		Permissions: permissions,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func (m *RoleMapper) ToModel(entity *entity.Role) *model.RoleModel {
	if entity == nil {
		return nil
	}

	permissionsJSON, _ := json.Marshal(entity.Permissions)

	return &model.RoleModel{
		ID:          entity.ID,
		Name:        entity.Name,
		Slug:        entity.Slug,
		Permissions: datatypes.JSON(permissionsJSON),
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}

// TenantMemberMapper handles mapping between entity and model
type TenantMemberMapper struct {
	tenantMapper *TenantMapper
	userMapper   *UserMapper
	roleMapper   *RoleMapper
}

func NewTenantMemberMapper() *TenantMemberMapper {
	return &TenantMemberMapper{
		tenantMapper: NewTenantMapper(),
		userMapper:   NewUserMapper(),
		roleMapper:   NewRoleMapper(),
	}
}

func (m *TenantMemberMapper) ToDomain(model *model.TenantMemberModel) *entity.TenantMember {
	if model == nil {
		return nil
	}
	member := &entity.TenantMember{
		ID:        model.ID,
		TenantID:  model.TenantID,
		UserID:    model.UserID,
		RoleID:    model.RoleID,
		AppID:     model.AppID,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
	if model.Tenant != nil {
		member.Tenant = m.tenantMapper.ToDomain(model.Tenant)
	}
	if model.User != nil {
		member.User = m.userMapper.ToDomain(model.User)
	}
	if model.Role != nil {
		member.Role = m.roleMapper.ToDomain(model.Role)
	}
	return member
}

func (m *TenantMemberMapper) ToModel(entity *entity.TenantMember) *model.TenantMemberModel {
	if entity == nil {
		return nil
	}
	return &model.TenantMemberModel{
		ID:        entity.ID,
		TenantID:  entity.TenantID,
		UserID:    entity.UserID,
		RoleID:    entity.RoleID,
		AppID:     entity.AppID,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	}
}
