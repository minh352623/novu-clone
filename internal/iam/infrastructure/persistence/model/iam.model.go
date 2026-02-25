package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// PricingPlanModel is the GORM model for pricing_plans table
type PricingPlanModel struct {
	ID                  uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	Name                string    `gorm:"column:name;type:text;not null"`
	Slug                string    `gorm:"column:slug;type:text;uniqueIndex;not null"`
	MonthlyCredits      int64     `gorm:"column:monthly_credits;default:0"`
	Price               float64   `gorm:"column:price;type:numeric(10,2);default:0.00"`
	Currency            string    `gorm:"column:currency;type:text;default:'USD'"`
	Description         *string   `gorm:"column:description;type:text"`
	IsActive            bool      `gorm:"column:is_active;default:true"`
	IsDefault           bool      `gorm:"column:is_default;default:false"`
	MaxApps             int       `gorm:"column:max_apps;type:integer;default:0"`
	MaxMembers          int       `gorm:"column:max_members;type:integer;default:0"`
	MaxWorkflows        int       `gorm:"column:max_workflows;type:integer;default:0"`
	MaxMessagesPerMonth int64     `gorm:"column:max_messages_per_month;type:bigint;default:0"`
	RateLimitRPM        int       `gorm:"column:rate_limit_rpm;type:integer;default:0"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (PricingPlanModel) TableName() string {
	return "pricing_plans"
}

// TenantModel is the GORM model for tenants table
type TenantModel struct {
	ID            uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	Name          string     `gorm:"column:name;type:text;not null"`
	Slug          string     `gorm:"column:slug;type:text;uniqueIndex;not null"`
	Status        string     `gorm:"column:status;type:text;default:'active'"`
	PricingPlanID *uuid.UUID `gorm:"column:pricing_plan_id;type:uuid"`
	PlanStartDate time.Time  `gorm:"column:plan_start_date;autoCreateTime"`
	CreatedAt     time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;autoUpdateTime"`

	// Relationships
	PricingPlan *PricingPlanModel `gorm:"foreignKey:PricingPlanID"`
}

func (TenantModel) TableName() string {
	return "tenants"
}

// UserModel is the GORM model for users table
type UserModel struct {
	ID           uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	Email        string     `gorm:"column:email;type:text;uniqueIndex;not null"`
	PasswordHash string     `gorm:"column:password_hash;type:text;not null"`
	FullName     *string    `gorm:"column:full_name;type:text"`
	IsRootAdmin  bool       `gorm:"column:is_root_admin;default:false"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (UserModel) TableName() string {
	return "users"
}

// RoleModel is the GORM model for roles table
type RoleModel struct {
	ID          uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	Name        string         `gorm:"column:name;type:text;not null"`
	Slug        string         `gorm:"column:slug;type:text;not null"`
	Permissions datatypes.JSON `gorm:"column:permissions;type:jsonb;default:'{}'"`
	CreatedAt   time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

func (RoleModel) TableName() string {
	return "roles"
}

// TenantMemberModel is the GORM model for tenant_members table
type TenantMemberModel struct {
	ID        uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	TenantID  uuid.UUID  `gorm:"column:tenant_id;type:uuid;not null"`
	UserID    uuid.UUID  `gorm:"column:user_id;type:uuid;not null"`
	RoleID    *uuid.UUID `gorm:"column:role_id;type:uuid"`
	AppID     *uuid.UUID `gorm:"column:app_id;type:uuid"` // Nullable
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`

	// Relationships
	Tenant *TenantModel `gorm:"foreignKey:TenantID"`
	User   *UserModel   `gorm:"foreignKey:UserID"`
	Role   *RoleModel   `gorm:"foreignKey:RoleID"`
}

func (TenantMemberModel) TableName() string {
	return "tenant_members"
}
