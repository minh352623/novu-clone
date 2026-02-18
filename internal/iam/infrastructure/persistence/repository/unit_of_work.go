package repository

import (
	"context"

	"CONVERDA/internal/iam/domain/repository"

	"gorm.io/gorm"
)

type gormIAMUnitOfWork struct {
	db *gorm.DB
}

// NewIAMUnitOfWork creates a new GORM-based implementation of IAMUnitOfWork
func NewIAMUnitOfWork(db *gorm.DB) repository.IAMUnitOfWork {
	return &gormIAMUnitOfWork{db: db}
}

func (u *gormIAMUnitOfWork) Execute(ctx context.Context, fn func(repository.IAMTxRepository) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &gormIAMTxRepository{
			members:     NewTenantMemberRepository(tx),
			invitations: NewInvitationRepository(tx),
			users:       NewUserRepository(tx),
			roles:       NewRoleRepository(tx),
			tenants:     NewTenantRepository(tx),
		}
		return fn(txRepo)
	})
}

type gormIAMTxRepository struct {
	members     repository.TenantMemberRepository
	invitations repository.InvitationRepository
	users       repository.UserRepository
	roles       repository.RoleRepository
	tenants     repository.TenantRepository
}

func (r *gormIAMTxRepository) Members() repository.TenantMemberRepository {
	return r.members
}

func (r *gormIAMTxRepository) Invitations() repository.InvitationRepository {
	return r.invitations
}

func (r *gormIAMTxRepository) Users() repository.UserRepository {
	return r.users
}

func (r *gormIAMTxRepository) Roles() repository.RoleRepository {
	return r.roles
}

func (r *gormIAMTxRepository) Tenants() repository.TenantRepository {
	return r.tenants
}
