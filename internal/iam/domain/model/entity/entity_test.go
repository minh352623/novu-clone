package entity_test

import (
	"testing"

	"CONVERDA/internal/iam/domain/model/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		password      string
		fullName      *string
		expectedError error
	}{
		{
			name:          "valid user",
			email:         "test@example.com",
			password:      "password123",
			fullName:      nil,
			expectedError: nil,
		},
		{
			name:          "empty email",
			email:         "",
			password:      "password123",
			fullName:      nil,
			expectedError: entity.ErrUserEmailRequired,
		},
		{
			name:          "empty password",
			email:         "test@example.com",
			password:      "",
			fullName:      nil,
			expectedError: entity.ErrUserPasswordRequired,
		},
		{
			name:          "password too short",
			email:         "test@example.com",
			password:      "short",
			fullName:      nil,
			expectedError: entity.ErrUserPasswordTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := entity.NewUser(tt.email, tt.password, tt.fullName)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.NotEmpty(t, user.PasswordHash)
				assert.NotEqual(t, tt.password, user.PasswordHash) // Should be hashed
			}
		})
	}
}

func TestUser_VerifyPassword(t *testing.T) {
	password := "correctpassword123"
	user, err := entity.NewUser("test@example.com", password, nil)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		password      string
		expectedError error
	}{
		{
			name:          "correct password",
			password:      password,
			expectedError: nil,
		},
		{
			name:          "wrong password",
			password:      "wrongpassword",
			expectedError: entity.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.VerifyPassword(tt.password)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewTenant(t *testing.T) {
	tests := []struct {
		name          string
		tenantName    string
		slug          string
		expectedError error
	}{
		{
			name:          "valid tenant",
			tenantName:    "Acme Corp",
			slug:          "acme-corp",
			expectedError: nil,
		},
		{
			name:          "empty name",
			tenantName:    "",
			slug:          "acme-corp",
			expectedError: entity.ErrTenantNameRequired,
		},
		{
			name:          "empty slug",
			tenantName:    "Acme Corp",
			slug:          "",
			expectedError: entity.ErrTenantSlugRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant, err := entity.NewTenant(tt.tenantName, tt.slug)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, tenant)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tenant)
				assert.Equal(t, tt.tenantName, tenant.Name)
				assert.Equal(t, tt.slug, tenant.Slug)
			}
		})
	}
}

func TestNewRole(t *testing.T) {
	tests := []struct {
		name          string
		roleName      string
		slug          string
		expectedError error
	}{
		{
			name:          "valid role",
			roleName:      "Tenant Admin",
			slug:          "tenant_admin",
			expectedError: nil,
		},
		{
			name:          "empty name",
			roleName:      "",
			slug:          "tenant_admin",
			expectedError: entity.ErrRoleNameRequired,
		},
		{
			name:          "empty slug",
			roleName:      "Tenant Admin",
			slug:          "",
			expectedError: entity.ErrRoleSlugRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := entity.NewRole(tt.roleName, tt.slug)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, role)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, role)
				assert.Equal(t, tt.roleName, role.Name)
				assert.Equal(t, tt.slug, role.Slug)
				assert.NotNil(t, role.Permissions)
			}
		})
	}
}

func TestRole_Permissions(t *testing.T) {
	role, _ := entity.NewRole("Admin", "admin")

	// Initially no permissions
	assert.False(t, role.HasPermission("users.create"))

	// Set permission
	role.SetPermission("users.create", true)
	assert.True(t, role.HasPermission("users.create"))

	// Remove permission
	role.RemovePermission("users.create")
	assert.False(t, role.HasPermission("users.create"))
}

func TestNewTenantMember(t *testing.T) {
	tenantID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name          string
		tenantID      uuid.UUID
		userID        uuid.UUID
		roleID        *uuid.UUID
		appID         *uuid.UUID
		expectedError error
	}{
		{
			name:          "valid member",
			tenantID:      tenantID,
			userID:        userID,
			roleID:        nil,
			appID:         nil,
			expectedError: nil,
		},
		{
			name:          "nil tenant ID",
			tenantID:      uuid.Nil,
			userID:        userID,
			roleID:        nil,
			appID:         nil,
			expectedError: entity.ErrMemberTenantRequired,
		},
		{
			name:          "nil user ID",
			tenantID:      tenantID,
			userID:        uuid.Nil,
			roleID:        nil,
			appID:         nil,
			expectedError: entity.ErrMemberUserRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member, err := entity.NewTenantMember(tt.tenantID, tt.userID, tt.roleID, tt.appID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, member)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, member)
				assert.Equal(t, tt.tenantID, member.TenantID)
				assert.Equal(t, tt.userID, member.UserID)
				assert.Equal(t, tt.appID, member.AppID)
			}
		})
	}
}
