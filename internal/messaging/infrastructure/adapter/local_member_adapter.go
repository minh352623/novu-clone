package adapter

import (
	"context"

	iamRepo "CONVERDA/internal/iam/domain/repository"
	"CONVERDA/internal/messaging/domain/repository"

	"github.com/google/uuid"
)

type localMemberAdapter struct {
	memberRepo iamRepo.TenantMemberRepository
}

// NewLocalMemberAdapter creates a new local adapter for MemberReader
func NewLocalMemberAdapter(memberRepo iamRepo.TenantMemberRepository) repository.MemberReader {
	return &localMemberAdapter{memberRepo: memberRepo}
}

func (a *localMemberAdapter) GetMember(ctx context.Context, id uuid.UUID) (*repository.MemberInfo, error) {
	member, err := a.memberRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, nil
	}

	displayName := ""
	if member.User != nil && member.User.FullName != nil {
		displayName = *member.User.FullName
	}

	return &repository.MemberInfo{
		ID:          member.ID,
		TenantID:    member.TenantID,
		UserID:      member.UserID,
		RoleID:      member.RoleID,
		DisplayName: displayName,
	}, nil
}
